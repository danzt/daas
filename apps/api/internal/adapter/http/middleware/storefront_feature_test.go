package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	mw "github.com/danzt/daas/api/internal/adapter/http/middleware"
)

// mockRow implements pgx.Row for testing StorefrontFeatureMiddleware.
type mockRow struct {
	enabled bool
	err     error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if b, ok := dest[0].(*bool); ok {
		*b = r.enabled
	}
	return nil
}

// mockQuerier implements mw.ConnQuerier with a fixed row result.
type mockQuerier struct {
	row *mockRow
}

func (q *mockQuerier) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return q.row
}

func TestStorefrontFeatureMW_EnabledCallsNext(t *testing.T) {
	const tenantID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	q := &mockQuerier{row: &mockRow{enabled: true}}
	m := mw.NewStorefrontFeatureMiddlewareWithQuerier(q)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/tienda/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(mw.ContextKeyTenantID), tenantID)

	nextCalled := false
	handler := m.Handle()(func(c echo.Context) error {
		nextCalled = true
		return c.JSON(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestStorefrontFeatureMW_DisabledReturns403(t *testing.T) {
	const tenantID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	q := &mockQuerier{row: &mockRow{enabled: false}}
	m := mw.NewStorefrontFeatureMiddlewareWithQuerier(q)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/tienda/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(mw.ContextKeyTenantID), tenantID)

	_ = m.Handle()(successHandler)(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestStorefrontFeatureMW_NoRowReturns403(t *testing.T) {
	const tenantID = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	q := &mockQuerier{row: &mockRow{err: pgx.ErrNoRows}}
	m := mw.NewStorefrontFeatureMiddlewareWithQuerier(q)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/tienda/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(mw.ContextKeyTenantID), tenantID)

	_ = m.Handle()(successHandler)(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestStorefrontFeatureMW_MissingTenantIDReturns403(t *testing.T) {
	q := &mockQuerier{row: &mockRow{enabled: true}}
	m := mw.NewStorefrontFeatureMiddlewareWithQuerier(q)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/t/tienda/shop/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No tenant_id set on context

	_ = m.Handle()(successHandler)(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
