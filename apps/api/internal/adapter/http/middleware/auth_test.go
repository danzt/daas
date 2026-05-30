package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
)

// makeTestKeyPair generates an RSA key pair and a JWKS server for testing.
func makeTestKeyPair(t *testing.T) (*rsa.PrivateKey, *httptest.Server) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	pubKey, err := jwk.PublicKeyOf(privateKey)
	if err != nil {
		t.Fatalf("get public key: %v", err)
	}
	if err := pubKey.Set(jwk.AlgorithmKey, jwa.RS256()); err != nil {
		t.Fatalf("set algorithm: %v", err)
	}
	if err := pubKey.Set(jwk.KeyIDKey, "test-key-id"); err != nil {
		t.Fatalf("set key id: %v", err)
	}

	set := jwk.NewSet()
	if err := set.AddKey(pubKey); err != nil {
		t.Fatalf("add key to set: %v", err)
	}

	jwksBytes, err := json.Marshal(set)
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksBytes)
	}))

	return privateKey, srv
}

// signToken creates a signed JWT with the given private key and claims.
func signToken(t *testing.T, privateKey *rsa.PrivateKey, sub, tenantID string, expiry time.Time) string {
	t.Helper()

	privJWK, err := jwk.Import(privateKey)
	if err != nil {
		t.Fatalf("import private key: %v", err)
	}
	if err := privJWK.Set(jwk.AlgorithmKey, jwa.RS256()); err != nil {
		t.Fatalf("set algorithm: %v", err)
	}
	if err := privJWK.Set(jwk.KeyIDKey, "test-key-id"); err != nil {
		t.Fatalf("set key id: %v", err)
	}

	tok, err := jwt.NewBuilder().
		Subject(sub).
		Expiration(expiry).
		IssuedAt(time.Now()).
		Build()
	if err != nil {
		t.Fatalf("build token: %v", err)
	}

	if tenantID != "" {
		appMeta := map[string]interface{}{"tenant_id": tenantID}
		if err := tok.Set("app_metadata", appMeta); err != nil {
			t.Fatalf("set app_metadata: %v", err)
		}
	}

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256(), privJWK))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return string(signed)
}

func TestAuthMiddleware_ValidJWT(t *testing.T) {
	privateKey, srv := makeTestKeyPair(t)
	defer srv.Close()

	m := mw.NewAuthMiddleware(srv.URL)

	e := echo.New()
	called := false
	handler := m.Handle()(func(c echo.Context) error {
		called = true
		uid := c.Get("supabase_uid").(string)
		tid := c.Get("tenant_id").(string)
		if uid != "user-123" {
			t.Errorf("expected supabase_uid=user-123, got %s", uid)
		}
		if tid != "tenant-456" {
			t.Errorf("expected tenant_id=tenant-456, got %s", tid)
		}
		return c.JSON(http.StatusOK, "ok")
	})

	token := signToken(t, privateKey, "user-123", "tenant-456", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_ExpiredJWT(t *testing.T) {
	privateKey, srv := makeTestKeyPair(t)
	defer srv.Close()

	m := mw.NewAuthMiddleware(srv.URL)

	e := echo.New()
	handler := m.Handle()(func(c echo.Context) error {
		t.Fatal("next handler should not be called for expired JWT")
		return nil
	})

	// Token expired 1 hour ago.
	token := signToken(t, privateKey, "user-123", "tenant-456", time.Now().Add(-time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = handler(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired JWT, got %d", rec.Code)
	}
}

func TestAuthMiddleware_MissingTenantID(t *testing.T) {
	privateKey, srv := makeTestKeyPair(t)
	defer srv.Close()

	m := mw.NewAuthMiddleware(srv.URL)

	e := echo.New()
	handler := m.Handle()(func(c echo.Context) error {
		t.Fatal("next handler should not be called when tenant_id is missing")
		return nil
	})

	// tenantID="" means app_metadata has no tenant_id.
	token := signToken(t, privateKey, "user-123", "", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = handler(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for missing tenant_id, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidSignature(t *testing.T) {
	_, srv := makeTestKeyPair(t)
	defer srv.Close()

	// Create a DIFFERENT key to sign the token (signature mismatch).
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate other RSA key: %v", err)
	}

	m := mw.NewAuthMiddleware(srv.URL)

	e := echo.New()
	handler := m.Handle()(func(c echo.Context) error {
		t.Fatal("next handler should not be called for invalid signature")
		return nil
	})

	token := signToken(t, otherKey, "user-123", "tenant-456", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = handler(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid signature, got %d", rec.Code)
	}
}
