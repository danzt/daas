package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
)

// TestSubdomainResolver_ReturnsNotImplemented verifies STORE-RESOLVE-5:
// SubdomainResolver is a stub that returns ErrNotImplemented until S11.
func TestSubdomainResolver_ReturnsNotImplemented(t *testing.T) {
	r := mw.NewSubdomainResolver()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tenantID, err := r.Resolve(c)
	if tenantID != "" {
		t.Errorf("expected empty tenantID, got %q", tenantID)
	}
	if err != mw.ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented, got %v", err)
	}
}

// TestTenantResolver_Interface verifies that SubdomainResolver satisfies
// the TenantResolver interface at compile time.
func TestTenantResolver_Interface(t *testing.T) {
	var _ mw.TenantResolver = mw.NewSubdomainResolver()
}
