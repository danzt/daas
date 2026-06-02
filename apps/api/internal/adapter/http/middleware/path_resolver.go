package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// SlugLookup is a thin interface that abstracts the DB call for resolving
// a tenant slug to a tenant UUID. The production implementation calls the
// SECURITY DEFINER function lookup_tenant_by_slug($1). Tests inject a mock.
type SlugLookup interface {
	LookupTenantBySlug(ctx context.Context, slug string) (string, error)
}

// poolSlugLookup wraps *pgxpool.Pool to implement SlugLookup.
type poolSlugLookup struct {
	pool *pgxpool.Pool
}

// LookupTenantBySlug executes the SECURITY DEFINER function that returns the
// tenant UUID for an active slug, or ErrTenantNotFound if no match.
func (p *poolSlugLookup) LookupTenantBySlug(ctx context.Context, slug string) (string, error) {
	var tenantID *string
	err := p.pool.QueryRow(ctx, "SELECT lookup_tenant_by_slug($1)", slug).Scan(&tenantID)
	if err != nil {
		return "", err
	}
	if tenantID == nil {
		return "", ErrTenantNotFound
	}
	return *tenantID, nil
}

// cacheEntry holds a cached tenant ID with an expiry timestamp.
type cacheEntry struct {
	tenantID  string
	expiresAt time.Time
}

// PathResolver resolves the tenant for a request by reading the ":tenantSlug"
// Echo path parameter and calling lookup_tenant_by_slug via the DB.
//
// Resolved slugs are cached in a sync.Map with a configurable TTL (default
// 5 minutes). Negative results (unknown / inactive slugs) are NEVER cached so
// that a tenant activation is visible immediately on the next request.
type PathResolver struct {
	lookup SlugLookup
	cache  sync.Map
	ttl    time.Duration
}

// NewPathResolver creates a PathResolver backed by the given pool.
// TTL defaults to 5 minutes; use NewPathResolverWithLookup for testing.
func NewPathResolver(pool *pgxpool.Pool) *PathResolver {
	return &PathResolver{
		lookup: &poolSlugLookup{pool: pool},
		ttl:    5 * time.Minute,
	}
}

// NewPathResolverWithLookup creates a PathResolver with an injectable
// SlugLookup and explicit TTL. Intended for unit tests.
func NewPathResolverWithLookup(lookup SlugLookup, ttl time.Duration) *PathResolver {
	return &PathResolver{
		lookup: lookup,
		ttl:    ttl,
	}
}

// Resolve reads ":tenantSlug" from the Echo path params, checks the
// sync.Map cache, and falls back to the DB on miss or expiry.
// Returns ErrTenantNotFound when the slug maps to no active tenant.
func (r *PathResolver) Resolve(c echo.Context) (string, error) {
	slug := c.Param("tenantSlug")

	// Check cache — valid only when the entry exists AND has not expired.
	if raw, ok := r.cache.Load(slug); ok {
		entry := raw.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.tenantID, nil
		}
		// Expired — remove stale entry and fall through to DB.
		r.cache.Delete(slug)
	}

	// Cache miss or expired: query the DB.
	tenantID, err := r.lookup.LookupTenantBySlug(c.Request().Context(), slug)
	if err != nil {
		// Do NOT cache negative results (ErrTenantNotFound or any DB error).
		return "", err
	}

	// Cache the positive result with TTL.
	r.cache.Store(slug, cacheEntry{
		tenantID:  tenantID,
		expiresAt: time.Now().Add(r.ttl),
	})

	return tenantID, nil
}
