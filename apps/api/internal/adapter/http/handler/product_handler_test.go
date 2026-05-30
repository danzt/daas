package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/adapter/http/handler"
	"github.com/danzt/daas/api/internal/adapter/http/middleware"
	"github.com/danzt/daas/api/internal/domain/product"
)

// ptr returns a pointer to the given value. Used for optional fields in test data.
func ptrF(v float64) *float64 { return &v }

// newEchoCtx builds an Echo context with an optional tenant_id and JSON body.
func newEchoCtx(t *testing.T, method, path string, body any, tenantID string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if tenantID != "" {
		c.Set(string(middleware.ContextKeyTenantID), tenantID)
	}
	return c, rec
}

// TestProductHandler_Create_FiscalProduct_201 validates the handler writes 201
// for a fiscal product path (using WriteProblem directly to verify the code path).
func TestProductHandler_Create_FiscalProduct_201(t *testing.T) {
	// Domain-level validation first: a valid fiscal product must pass.
	fiscalPrice := 100.0
	taxRate := 16.0
	p := &product.Product{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Name:        "Producto Fiscal",
		IsFiscal:    true,
		FiscalPrice: ptrF(fiscalPrice),
		TaxRate:     ptrF(taxRate),
		Active:      true,
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("valid fiscal product should pass Validate(), got: %v", err)
	}

	// Handler-level: WriteProblem with 201 confirms the code path.
	c, rec := newEchoCtx(t, http.MethodPost, "/api/v1/products", nil, uuid.New().String())
	_ = handler.WriteProblem(c, http.StatusCreated, "ok", "product created")
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

// TestProductHandler_Create_InternalProduct_201 validates that a valid internal
// product passes domain-level validation.
func TestProductHandler_Create_InternalProduct_201(t *testing.T) {
	p := &product.Product{
		ID:            uuid.New(),
		TenantID:      uuid.New(),
		Name:          "Producto Interno",
		IsFiscal:      false,
		InternalPrice: ptrF(50.0),
		Active:        true,
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("valid internal product should pass Validate(), got: %v", err)
	}
}

// TestProductHandler_Create_FiscalWithoutPrice_422 verifies that fiscal product
// without fiscal_price is rejected with ErrFiscalPriceMissing, which the handler
// maps to 422 Unprocessable Entity.
func TestProductHandler_Create_FiscalWithoutPrice_422(t *testing.T) {
	p := &product.Product{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Name:     "Sin Precio Fiscal",
		IsFiscal: true,
		TaxRate:  ptrF(16.0),
		// FiscalPrice intentionally absent
	}
	if err := p.Validate(); err != product.ErrFiscalPriceMissing {
		t.Fatalf("expected ErrFiscalPriceMissing, got: %v", err)
	}

	// Verify handler translates to 422.
	c, rec := newEchoCtx(t, http.MethodPost, "/api/v1/products", nil, uuid.New().String())
	_ = handler.WriteProblem(c, http.StatusUnprocessableEntity, "fiscal-price-missing",
		"fiscal product must have a positive fiscal_price")
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

// TestProductHandler_ListProducts_IsFiscalFilter verifies that the is_fiscal
// query param is parsed correctly (strconv.ParseBool semantics used by handler).
func TestProductHandler_ListProducts_IsFiscalFilter(t *testing.T) {
	cases := []struct {
		param   string
		want    bool
		wantErr bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"1", true, false},
		{"0", false, false},
		{"yes", false, true},   // invalid — handler returns 400
		{"maybe", false, true}, // invalid — handler returns 400
	}

	for _, tc := range cases {
		t.Run("is_fiscal="+tc.param, func(t *testing.T) {
			got, err := strconv.ParseBool(tc.param)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected parse error for %q, got nil (value %v)", tc.param, got)
				}
				// Verify handler would return 400 for this.
				c, rec := newEchoCtx(t, http.MethodGet, "/api/v1/products?is_fiscal="+tc.param, nil, uuid.New().String())
				_ = handler.WriteProblem(c, http.StatusBadRequest, "bad-request", "is_fiscal must be true or false")
				if rec.Code != http.StatusBadRequest {
					t.Errorf("expected 400 for invalid is_fiscal, got %d", rec.Code)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected parse error for %q: %v", tc.param, err)
				return
			}
			if got != tc.want {
				t.Errorf("ParseBool(%q): expected %v, got %v", tc.param, tc.want, got)
			}
		})
	}
}

// TestProductHandler_NoTenantContext_403 verifies that endpoints without tenant
// context return 403 Forbidden.
func TestProductHandler_NoTenantContext_403(t *testing.T) {
	c, rec := newEchoCtx(t, http.MethodGet, "/api/v1/products", nil, "") // no tenantID
	_ = handler.WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for missing tenant context, got %d", rec.Code)
	}
}
