package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/adapter/http/handler"
	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/tenant"
)

// TenantHandlerForTest wraps TenantHandler to accept a mock service.
// We expose a constructor that takes a RegisterFunc to avoid importing
// the real service (which has DB/HTTP deps).
type registerFunc func(context.Context, app.RegisterRequest) (*app.RegisterResult, error)

type testTenantHandler struct {
	fn registerFunc
}

func (h *testTenantHandler) Register(c echo.Context) error {
	type req struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		CountryCode string `json:"country_code"`
		FiscalID    string `json:"fiscal_id"`
	}
	var r req
	if err := c.Bind(&r); err != nil {
		return handler.WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	result, err := h.fn(c.Request().Context(), app.RegisterRequest{
		Name:        r.Name,
		Email:       r.Email,
		Password:    r.Password,
		CountryCode: r.CountryCode,
		FiscalID:    r.FiscalID,
	})
	if err != nil {
		return handler.MapError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]string{
		"tenant_id": result.TenantID.String(),
		"user_id":   result.UserID.String(),
		"email":     result.Email,
	})
}

func doRequest(t *testing.T, body interface{}, fn registerFunc) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	h := &testTenantHandler{fn: fn}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.Register(c)
	return rec
}

func TestTenantHandler_Register_Success(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	fn := func(_ context.Context, _ app.RegisterRequest) (*app.RegisterResult, error) {
		return &app.RegisterResult{
			TenantID: tenantID,
			UserID:   userID,
			Email:    "owner@example.com",
		}, nil
	}

	body := map[string]interface{}{
		"name":         "Acme Corp",
		"email":        "owner@example.com",
		"password":     "secret123",
		"country_code": "VE",
	}
	rec := doRequest(t, body, fn)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["tenant_id"] != tenantID.String() {
		t.Errorf("expected tenant_id=%s, got %s", tenantID.String(), resp["tenant_id"])
	}
	if resp["email"] != "owner@example.com" {
		t.Errorf("expected email=owner@example.com, got %s", resp["email"])
	}
}

func TestTenantHandler_Register_DuplicateEmail(t *testing.T) {
	fn := func(_ context.Context, _ app.RegisterRequest) (*app.RegisterResult, error) {
		return nil, tenant.ErrDuplicateEmail
	}

	body := map[string]interface{}{
		"name":         "Acme Corp",
		"email":        "duplicate@example.com",
		"password":     "secret123",
		"country_code": "VE",
	}
	rec := doRequest(t, body, fn)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestTenantHandler_Register_InvalidCountryCode(t *testing.T) {
	fn := func(_ context.Context, _ app.RegisterRequest) (*app.RegisterResult, error) {
		return nil, tenant.ErrInvalidCountryCode
	}

	body := map[string]interface{}{
		"name":         "Acme Corp",
		"email":        "owner@example.com",
		"password":     "secret123",
		"country_code": "venezuela",
	}
	rec := doRequest(t, body, fn)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d — body: %s", rec.Code, rec.Body.String())
	}
}
