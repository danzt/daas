package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
)

// mockPublicConn captures the SQL and args sent by PublicTenantMiddleware.
type mockPublicConn struct {
	capturedSQL  string
	capturedArgs []any
}

func (m *mockPublicConn) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	m.capturedSQL = sql
	m.capturedArgs = append(m.capturedArgs, args...)
	return pgconn.CommandTag{}, nil
}

func (m *mockPublicConn) Release() {}

// fixedResolver is a TenantResolver test double with a fixed result.
type fixedResolver struct {
	tenantID string
	err      error
}

func (f *fixedResolver) Resolve(_ echo.Context) (string, error) {
	return f.tenantID, f.err
}

// TestPublicTenantMW_UnknownSlug verifies that ErrTenantNotFound returns 404 JSON.
func TestPublicTenantMW_UnknownSlug(t *testing.T) {
	resolver := &fixedResolver{err: mw.ErrTenantNotFound}
	mwFn := mw.NewPublicTenantMiddlewareWithAcquire(resolver, func(_ context.Context) (mw.ConnExecer, error) {
		return &mockPublicConn{}, nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/bad-slug/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := mwFn.Handle()(successHandler)
	_ = handler(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// TestPublicTenantMW_SetsConfig verifies that PublicTenantMiddleware calls
// SELECT set_config('app.tenant_id', $1, true) with the correct tenantID.
func TestPublicTenantMW_SetsConfig(t *testing.T) {
	const tenantID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	resolver := &fixedResolver{tenantID: tenantID}
	conn := &mockPublicConn{}

	mwFn := mw.NewPublicTenantMiddlewareWithAcquire(resolver, func(_ context.Context) (mw.ConnExecer, error) {
		return conn, nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/tienda/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	handler := mwFn.Handle()(func(c echo.Context) error {
		nextCalled = true
		return c.JSON(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !nextCalled {
		t.Fatal("next handler not called")
	}

	const wantSQL = "SELECT set_config('app.tenant_id', $1, true)"
	if conn.capturedSQL != wantSQL {
		t.Errorf("SQL: got %q, want %q", conn.capturedSQL, wantSQL)
	}
	if len(conn.capturedArgs) == 0 {
		t.Fatal("no args passed to Exec")
	}
	if got, ok := conn.capturedArgs[0].(string); !ok || got != tenantID {
		t.Errorf("arg[0]: got %v, want %q", conn.capturedArgs[0], tenantID)
	}
}

// TestPublicTenantMW_OtherResolverError verifies that a non-ErrTenantNotFound
// error from the resolver returns 500.
func TestPublicTenantMW_OtherResolverError(t *testing.T) {
	resolver := &fixedResolver{err: errors.New("some transient error")}
	mwFn := mw.NewPublicTenantMiddlewareWithAcquire(resolver, func(_ context.Context) (mw.ConnExecer, error) {
		return &mockPublicConn{}, nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/tienda/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = mwFn.Handle()(successHandler)(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// TestPublicTenantMW_CacheControlHeader verifies that the middleware sets
// Cache-Control: no-store on the response.
func TestPublicTenantMW_CacheControlHeader(t *testing.T) {
	const tenantID = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	resolver := &fixedResolver{tenantID: tenantID}
	conn := &mockPublicConn{}

	mwFn := mw.NewPublicTenantMiddlewareWithAcquire(resolver, func(_ context.Context) (mw.ConnExecer, error) {
		return conn, nil
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/tienda/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = mwFn.Handle()(successHandler)(c)

	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control: got %q, want %q", cc, "no-store")
	}
}
