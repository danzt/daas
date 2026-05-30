// Package middleware contains Echo HTTP middleware for authentication and
// tenant isolation. The auth middleware verifies Supabase-issued JWTs using
// JWKS and extracts tenant context.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// Context keys for values stored by auth middleware.
// Using typed string keys prevents collisions with other context values.
type contextKey string

const (
	// ContextKeySupabaseUID is the key for the Supabase user UUID in echo.Context.
	ContextKeySupabaseUID contextKey = "supabase_uid"
	// ContextKeyTenantID is the key for the tenant UUID in echo.Context.
	ContextKeyTenantID contextKey = "tenant_id"
	// ContextKeyRole is the key for the user's role string in echo.Context.
	ContextKeyRole contextKey = "role"
)

// jwksCache holds a cached JWKS key set with a TTL.
// Supabase JWKS keys rotate infrequently; 1h is a safe cache duration.
type jwksCache struct {
	mu        sync.RWMutex
	keySet    jwk.Set
	fetchedAt time.Time
	ttl       time.Duration
}

func (c *jwksCache) get() (jwk.Set, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.keySet == nil || time.Since(c.fetchedAt) > c.ttl {
		return nil, false
	}
	return c.keySet, true
}

func (c *jwksCache) set(ks jwk.Set) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keySet = ks
	c.fetchedAt = time.Now()
}

// AuthMiddleware verifies Supabase JWTs and extracts tenant context.
//
// IMPORTANT — Supabase Auth Hook required:
// The `tenant_id` claim in the JWT comes from `app_metadata.tenant_id`.
// Supabase does NOT automatically include app_metadata fields as top-level
// JWT claims. To enable this, configure an Auth Hook (database hook) in the
// Supabase dashboard that copies app_metadata into the JWT payload.
// Alternatively, app_metadata is included in the standard Supabase JWT
// under the "app_metadata" claim directly. This middleware reads from there.
//
// Flow:
// 1. Extract Bearer token from Authorization header.
// 2. Fetch JWKS from Supabase (cached 1h) and verify JWT signature.
// 3. Extract `sub` (Supabase UID) and `tenant_id` from app_metadata.
// 4. Return 401 for expired/invalid JWTs (no DB call).
// 5. Return 403 if tenant_id is missing (Auth Hook not configured).
type AuthMiddleware struct {
	supabaseURL string
	cache       *jwksCache
}

// NewAuthMiddleware creates a new AuthMiddleware that validates JWTs against
// the given Supabase project URL.
func NewAuthMiddleware(supabaseURL string) *AuthMiddleware {
	return &AuthMiddleware{
		supabaseURL: supabaseURL,
		cache: &jwksCache{
			ttl: time.Hour,
		},
	}
}

// Handle returns the Echo middleware handler function.
func (m *AuthMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := m.extractBearer(c.Request())
			if err != nil {
				return problemJSON(c, http.StatusUnauthorized, "unauthorized", err.Error())
			}

			parsed, err := m.verifyToken(c.Request().Context(), token)
			if err != nil {
				return problemJSON(c, http.StatusUnauthorized, "unauthorized", "invalid or expired JWT")
			}

			// Extract Supabase UID from `sub` claim.
			sub, ok := parsed.Subject()
			if !ok || sub == "" {
				return problemJSON(c, http.StatusUnauthorized, "unauthorized", "missing sub claim in JWT")
			}

			// Extract tenant_id from app_metadata claim.
			// Supabase embeds app_metadata as a map under the "app_metadata" key.
			tenantID, err := extractTenantID(parsed)
			if err != nil {
				return problemJSON(c, http.StatusForbidden, "forbidden",
					"tenant_id claim missing from JWT — configure Supabase Auth Hook to include app_metadata.tenant_id")
			}

			// Extract role from app_metadata if present.
			role := extractRole(parsed)

			// Store values in context for downstream middleware and handlers.
			c.Set(string(ContextKeySupabaseUID), sub)
			c.Set(string(ContextKeyTenantID), tenantID)
			c.Set(string(ContextKeyRole), role)

			return next(c)
		}
	}
}

// extractBearer pulls the Bearer token from the Authorization header.
func (m *AuthMiddleware) extractBearer(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	if !strings.HasPrefix(header, "Bearer ") {
		return "", fmt.Errorf("authorization header must use Bearer scheme")
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("empty Bearer token")
	}
	return token, nil
}

// verifyToken fetches (or uses cached) JWKS and verifies the token signature
// and expiry. Returns the parsed JWT on success.
func (m *AuthMiddleware) verifyToken(ctx context.Context, token string) (jwt.Token, error) {
	ks, ok := m.cache.get()
	if !ok {
		var err error
		ks, err = m.fetchJWKS(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}
		m.cache.set(ks)
	}

	parsed, err := jwt.Parse([]byte(token), jwt.WithKeySet(ks), jwt.WithValidate(true))
	if err != nil {
		return nil, fmt.Errorf("JWT verification failed: %w", err)
	}
	return parsed, nil
}

// fetchJWKS retrieves the public key set from Supabase's JWKS endpoint.
func (m *AuthMiddleware) fetchJWKS(ctx context.Context) (jwk.Set, error) {
	url := m.supabaseURL + "/auth/v1/.well-known/jwks.json"
	ks, err := jwk.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("JWKS fetch from %s failed: %w", url, err)
	}
	return ks, nil
}

// extractTenantID reads tenant_id from the JWT. Supabase places app_metadata
// as a top-level claim in the JWT payload. We check:
//  1. Top-level `tenant_id` claim (from a custom Auth Hook).
//  2. `app_metadata.tenant_id` nested map (standard Supabase behavior).
func extractTenantID(tok jwt.Token) (string, error) {
	// Check top-level tenant_id first (Auth Hook approach).
	var topLevel string
	if err := tok.Get("tenant_id", &topLevel); err == nil && topLevel != "" {
		return topLevel, nil
	}

	// Check app_metadata.tenant_id (standard Supabase approach).
	var appMeta map[string]interface{}
	if err := tok.Get("app_metadata", &appMeta); err == nil && appMeta != nil {
		if tid, ok := appMeta["tenant_id"]; ok {
			if s, ok := tid.(string); ok && s != "" {
				return s, nil
			}
		}
	}

	return "", fmt.Errorf("tenant_id not found in JWT claims")
}

// extractRole attempts to read the user role from app_metadata. Returns empty
// string if absent; callers should treat that as employee-level access.
func extractRole(tok jwt.Token) string {
	var appMeta map[string]interface{}
	if err := tok.Get("app_metadata", &appMeta); err == nil && appMeta != nil {
		if r, ok := appMeta["role"]; ok {
			if s, ok := r.(string); ok {
				return s
			}
		}
	}
	return ""
}

// problemJSON writes an RFC 7807 Problem Details response.
func problemJSON(c echo.Context, status int, title, detail string) error {
	return c.JSON(status, map[string]interface{}{
		"type":   "https://daas.app/errors/" + strings.ToLower(strings.ReplaceAll(title, " ", "-")),
		"title":  title,
		"status": status,
		"detail": detail,
	})
}
