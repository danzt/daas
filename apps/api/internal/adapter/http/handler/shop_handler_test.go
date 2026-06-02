//go:build integration

// Integration tests for ShopHandler. They spin up a real PostgreSQL instance
// via testcontainers and exercise the full middleware chain: RLS + public catalog
// handler. Build with: go test -tags integration ./internal/adapter/http/handler/...
package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/adapter/http/handler"
	"github.com/danzt/daas/api/internal/adapter/http/middleware"
	"github.com/danzt/daas/api/internal/app"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// shopTestCtx creates an Echo context that has the tenant_id already set
// (simulating PublicTenantMiddleware having run).
func shopTestCtx(t *testing.T, method, path, tenantID string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(middleware.ContextKeyTenantID), tenantID)
	return c, rec
}

// decodeBody unmarshals the response body into target.
func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode response body: %v\nbody: %s", err, rec.Body.String())
	}
}

// seedProduct inserts one product row for the given tenant and returns its ID.
// category param is a display name only (not used as FK — pass "" if not needed).
// active maps to the `active` column.
func seedProduct(t *testing.T, pool *pgxpool.Pool, tenantID uuid.UUID, name, _ string, price float64, active bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO products (id, tenant_id, name, description, category_id,
		                      is_fiscal, internal_price, active)
		VALUES ($1, $2, $3, $4, NULL, FALSE, $5, $6)`,
		id, tenantID, name, "description for "+name, price, active,
	)
	if err != nil {
		t.Fatalf("seedProduct %q: %v", name, err)
	}
	return id
}

// seedTenant inserts a minimal tenant row and returns its ID and slug.
func seedTenant(t *testing.T, pool *pgxpool.Pool, name string) (uuid.UUID, string) {
	t.Helper()
	slug := fmt.Sprintf("test-%s", uuid.New().String()[:8])
	var id uuid.UUID
	err := pool.QueryRow(context.Background(), `
		INSERT INTO tenants (name, slug, status, country_code)
		VALUES ($1, $2, 'active', 'VE')
		RETURNING id`,
		name, slug,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedTenant: %v", err)
	}
	return id, slug
}

// activateRLS sets app.tenant_id on the pool's session so that RLS policies
// apply to all subsequent queries on the same connection.
// is_local=false (third arg) means "session-scoped" — persists for the
// lifetime of the connection, not just the current transaction.
// With pool_max_conns=1, this reliably applies to all ShopService queries.
func activateRLS(t *testing.T, pool *pgxpool.Pool, tenantID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"SELECT set_config('app.tenant_id', $1, false)", tenantID.String())
	if err != nil {
		t.Fatalf("activateRLS: %v", err)
	}
}

// newShopHandler creates a ShopHandler backed by the real pool (test helper).
func newShopHandler(pool *pgxpool.Pool) *handler.ShopHandler {
	svc := app.NewShopService(pool)
	return handler.NewShopHandler(svc)
}

// --------------------------------------------------------------------------
// S6-T8 Tests — CATALOG-TS-01..10
// --------------------------------------------------------------------------

// CATALOG-TS-01: ListProducts returns only active products.
func TestShopHandler_ListProducts_ActiveOnly(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test Active Only")

	seedProduct(t, pool, tenantID, "Product A", "Cat1", 10.0, true)
	seedProduct(t, pool, tenantID, "Product B", "Cat1", 20.0, true)
	seedProduct(t, pool, tenantID, "Product C", "Cat1", 30.0, true)
	seedProduct(t, pool, tenantID, "Inactive Product", "Cat1", 5.0, false)

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products", tenantID.String())
	activateRLS(t, pool, tenantID)

	if err := h.ListProducts(c); err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d\nbody: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data []any `json:"data"`
		Meta struct {
			TotalCount int `json:"total_count"`
		} `json:"meta"`
	}
	decodeBody(t, rec, &resp)

	if len(resp.Data) != 3 {
		t.Errorf("expected 3 active products, got %d", len(resp.Data))
	}
	if resp.Meta.TotalCount != 3 {
		t.Errorf("expected total_count=3, got %d", resp.Meta.TotalCount)
	}
}

// CATALOG-TS-02: Pagination defaults — page=1, per_page=20.
func TestShopHandler_ListProducts_Pagination_Default(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test Pagination")

	for i := range 25 {
		seedProduct(t, pool, tenantID, fmt.Sprintf("Product %d", i+1), "Cat", float64(i+1)*10, true)
	}

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products", tenantID.String())
	activateRLS(t, pool, tenantID)

	if err := h.ListProducts(c); err != nil {
		t.Fatalf("ListProducts: %v", err)
	}

	var resp struct {
		Data []any `json:"data"`
		Meta struct {
			TotalCount int `json:"total_count"`
			Page       int `json:"page"`
			PerPage    int `json:"per_page"`
			TotalPages int `json:"total_pages"`
		} `json:"meta"`
	}
	decodeBody(t, rec, &resp)

	if len(resp.Data) != 20 {
		t.Errorf("expected 20 items (default per_page), got %d", len(resp.Data))
	}
	if resp.Meta.TotalCount != 25 {
		t.Errorf("expected total_count=25, got %d", resp.Meta.TotalCount)
	}
	if resp.Meta.Page != 1 {
		t.Errorf("expected page=1, got %d", resp.Meta.Page)
	}
	if resp.Meta.PerPage != 20 {
		t.Errorf("expected per_page=20, got %d", resp.Meta.PerPage)
	}
	if resp.Meta.TotalPages != 2 {
		t.Errorf("expected total_pages=2, got %d", resp.Meta.TotalPages)
	}
}

// CATALOG-TS-04: page beyond total → 200 with empty data, correct meta.
func TestShopHandler_ListProducts_PageBeyondTotal(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test Page Beyond")

	for i := range 10 {
		seedProduct(t, pool, tenantID, fmt.Sprintf("Product %d", i+1), "Cat", float64(i+1)*5, true)
	}

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products?page=99", tenantID.String())
	activateRLS(t, pool, tenantID)

	if err := h.ListProducts(c); err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Data []any `json:"data"`
		Meta struct {
			TotalCount int `json:"total_count"`
			Page       int `json:"page"`
		} `json:"meta"`
	}
	decodeBody(t, rec, &resp)

	if len(resp.Data) != 0 {
		t.Errorf("expected empty data for page 99, got %d items", len(resp.Data))
	}
	if resp.Meta.TotalCount != 10 {
		t.Errorf("expected total_count=10, got %d", resp.Meta.TotalCount)
	}
	if resp.Meta.Page != 99 {
		t.Errorf("expected page=99, got %d", resp.Meta.Page)
	}
}

// CATALOG-TS-05: category filter.
func TestShopHandler_ListProducts_CategoryFilter(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test Category Filter")

	// Seed products with different category names in description (category_id is FK)
	// We need to seed real category IDs. Simplification: use a sub-test approach
	// using categories inline via the product_categories table.
	ctx := context.Background()
	elecID := uuid.New()
	clothID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO product_categories (id, tenant_id, name) VALUES ($1, $2, 'Electrónica'), ($3, $2, 'Ropa')`,
		elecID, tenantID, clothID)
	if err != nil {
		t.Fatalf("seed categories: %v", err)
	}

	for i := range 2 {
		id := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO products (id, tenant_id, name, is_fiscal, internal_price, active, category_id)
			VALUES ($1, $2, $3, FALSE, 50.0, TRUE, $4)`,
			id, tenantID, fmt.Sprintf("Electronics Product %d", i+1), elecID)
		if err != nil {
			t.Fatalf("seed electronics product: %v", err)
		}
	}
	for i := range 3 {
		id := uuid.New()
		_, err := pool.Exec(ctx, `
			INSERT INTO products (id, tenant_id, name, is_fiscal, internal_price, active, category_id)
			VALUES ($1, $2, $3, FALSE, 30.0, TRUE, $4)`,
			id, tenantID, fmt.Sprintf("Clothing Product %d", i+1), clothID)
		if err != nil {
			t.Fatalf("seed clothing product: %v", err)
		}
	}

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products?category=Electr%C3%B3nica", tenantID.String())
	activateRLS(t, pool, tenantID)

	if err := h.ListProducts(c); err != nil {
		t.Fatalf("ListProducts: %v", err)
	}

	var resp struct {
		Data []any `json:"data"`
		Meta struct {
			TotalCount int `json:"total_count"`
		} `json:"meta"`
	}
	decodeBody(t, rec, &resp)

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 electronics products, got %d", len(resp.Data))
	}
}

// CATALOG-TS-06: full-text search via q param.
func TestShopHandler_ListProducts_QSearch(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test Q Search")

	seedProduct(t, pool, tenantID, "Zapato Rojo", "Calzado", 50.0, true)
	seedProduct(t, pool, tenantID, "Camisa Azul", "Ropa", 30.0, true)
	seedProduct(t, pool, tenantID, "Zapato Deportivo", "Calzado", 70.0, true)

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products?q=zapato", tenantID.String())
	activateRLS(t, pool, tenantID)

	if err := h.ListProducts(c); err != nil {
		t.Fatalf("ListProducts: %v", err)
	}

	var resp struct {
		Data []any `json:"data"`
	}
	decodeBody(t, rec, &resp)

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 products matching 'zapato', got %d", len(resp.Data))
	}
}

// CATALOG-TS-03 + per_page validation: per_page=200 → 400.
func TestShopHandler_ListProducts_InvalidPerPage(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test Invalid PerPage")

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products?per_page=200", tenantID.String())

	if err := h.ListProducts(c); err != nil {
		// handler may return error or write response directly
		t.Logf("ListProducts returned error (acceptable): %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("per_page=200 should return 400, got %d\nbody: %s", rec.Code, rec.Body.String())
	}

	var errBody struct {
		Error string `json:"error"`
		Field string `json:"field"`
	}
	decodeBody(t, rec, &errBody)
	if errBody.Field != "per_page" {
		t.Errorf("expected field='per_page', got %q", errBody.Field)
	}
}

// CATALOG-TS-07: GetProduct happy path.
func TestShopHandler_GetProduct_HappyPath(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test GetProduct")
	productID := seedProduct(t, pool, tenantID, "Laptop Pro", "Electronics", 999.0, true)

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products/"+productID.String(), tenantID.String())
	c.SetParamNames("id")
	c.SetParamValues(productID.String())
	activateRLS(t, pool, tenantID)

	if err := h.GetProduct(c); err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d\nbody: %s", rec.Code, rec.Body.String())
	}

	var product map[string]any
	decodeBody(t, rec, &product)

	if product["id"] != productID.String() {
		t.Errorf("expected product id=%s, got %v", productID, product["id"])
	}
	if product["name"] != "Laptop Pro" {
		t.Errorf("expected name='Laptop Pro', got %v", product["name"])
	}
}

// CATALOG-TS-08: inactive product returns 404.
func TestShopHandler_GetProduct_InactiveReturns404(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test GetProduct Inactive")
	productID := seedProduct(t, pool, tenantID, "Inactive Product", "Cat", 50.0, false)

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products/"+productID.String(), tenantID.String())
	c.SetParamNames("id")
	c.SetParamValues(productID.String())
	activateRLS(t, pool, tenantID)

	if err := h.GetProduct(c); err != nil {
		t.Logf("GetProduct returned error (acceptable): %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for inactive product, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
}

// CATALOG-TS-09: cross-tenant isolation — tenant-B product is 404 for tenant-A.
func TestShopHandler_CrossTenantIsolation(t *testing.T) {
	pool := setupTestDB(t)
	tenantA, _ := seedTenant(t, pool, "Tenant A")
	tenantB, _ := seedTenant(t, pool, "Tenant B")

	// Seed a product for tenant B.
	productID := seedProduct(t, pool, tenantB, "Tenant B Product", "Cat", 100.0, true)

	// Request as tenant A — should get 404.
	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products/"+productID.String(), tenantA.String())
	c.SetParamNames("id")
	c.SetParamValues(productID.String())
	activateRLS(t, pool, tenantA) // RLS active for tenant A

	if err := h.GetProduct(c); err != nil {
		t.Logf("GetProduct returned error (acceptable): %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("cross-tenant: expected 404, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
}

// CATALOG-TS-10: sensitive fields not exposed in response.
func TestShopHandler_SensitiveFieldsNotExposed(t *testing.T) {
	pool := setupTestDB(t)
	tenantID, _ := seedTenant(t, pool, "Test Sensitive Fields")
	productID := seedProduct(t, pool, tenantID, "Secure Product", "Cat", 150.0, true)

	h := newShopHandler(pool)
	c, rec := shopTestCtx(t, http.MethodGet, "/products/"+productID.String(), tenantID.String())
	c.SetParamNames("id")
	c.SetParamValues(productID.String())
	activateRLS(t, pool, tenantID)

	if err := h.GetProduct(c); err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	sensitiveFields := []string{"cost_price", "supplier_id", "fiscal_id", "created_by"}
	for _, field := range sensitiveFields {
		if contains(body, field) {
			t.Errorf("response body must NOT contain %q, but does.\nbody: %s", field, body)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
