package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	fiscaladapter "github.com/danzt/daas/api/internal/adapter/fiscal"
	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/fiscal"
)

// FiscalInvoiceHandler exposes HTTP endpoints for fiscal invoicing.
type FiscalInvoiceHandler struct {
	svc *app.FiscalInvoiceService
}

// NewFiscalInvoiceHandler creates a FiscalInvoiceHandler backed by the given pool.
// It uses the MockAdapter; swap for a real TCP adapter in production.
func NewFiscalInvoiceHandler(pool *pgxpool.Pool) *FiscalInvoiceHandler {
	adapter := fiscaladapter.NewMockAdapter()
	return &FiscalInvoiceHandler{svc: app.NewFiscalInvoiceService(pool, adapter)}
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type createFiscalLineReq struct {
	ProductID   string  `json:"product_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
}

type createFiscalInvoiceReq struct {
	CustomerName     string                `json:"customer_name"`
	CustomerIDType   fiscal.CustomerIDType `json:"customer_id_type"`
	CustomerIDNumber string                `json:"customer_id_number"`
	Notes            string                `json:"notes"`
	Lines            []createFiscalLineReq `json:"lines"`
}

type cancelFiscalReq struct {
	Notes string `json:"notes"`
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// Create handles POST /api/v1/invoices/fiscal
func (h *FiscalInvoiceHandler) Create(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}

	var body createFiscalInvoiceReq
	if err := c.Bind(&body); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid request body")
	}

	lines, err := buildFiscalLineRequests(body.Lines)
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", err.Error())
	}

	req := fiscal.CreateRequest{
		CustomerName:     body.CustomerName,
		CustomerIDType:   body.CustomerIDType,
		CustomerIDNumber: body.CustomerIDNumber,
		Notes:            body.Notes,
		Lines:            lines,
	}

	inv, err := h.svc.Create(c.Request().Context(), tenantID, supabaseUID, req)
	if err != nil {
		return mapFiscalError(c, err)
	}
	return c.JSON(http.StatusCreated, inv)
}

// List handles GET /api/v1/invoices/fiscal
func (h *FiscalInvoiceHandler) List(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	f := app.FiscalInvoiceListFilters{Limit: 50, Offset: 0}

	if s := c.QueryParam("status"); s != "" {
		st := fiscal.Status(s)
		f.Status = &st
	}
	if v := c.QueryParam("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.From = &t
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.To = &t
		}
	}
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			f.Limit = n
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			f.Offset = n
		}
	}

	invoices, err := h.svc.List(c.Request().Context(), tenantID, f)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	if invoices == nil {
		invoices = []*fiscal.FiscalInvoice{}
	}
	return c.JSON(http.StatusOK, invoices)
}

// GetByID handles GET /api/v1/invoices/fiscal/:id
func (h *FiscalInvoiceHandler) GetByID(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid invoice id")
	}
	inv, err := h.svc.GetByID(c.Request().Context(), tenantID, invoiceID)
	if err != nil {
		return mapFiscalError(c, err)
	}
	return c.JSON(http.StatusOK, inv)
}

// Issue handles POST /api/v1/invoices/fiscal/:id/issue
func (h *FiscalInvoiceHandler) Issue(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid invoice id")
	}
	inv, err := h.svc.Issue(c.Request().Context(), tenantID, invoiceID, supabaseUID)
	if err != nil {
		return mapFiscalError(c, err)
	}
	return c.JSON(http.StatusOK, inv)
}

// Cancel handles POST /api/v1/invoices/fiscal/:id/cancel
func (h *FiscalInvoiceHandler) Cancel(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid invoice id")
	}
	var body cancelFiscalReq
	_ = c.Bind(&body)
	inv, err := h.svc.Cancel(c.Request().Context(), tenantID, invoiceID, supabaseUID, body.Notes)
	if err != nil {
		return mapFiscalError(c, err)
	}
	return c.JSON(http.StatusOK, inv)
}

// Retry handles POST /api/v1/invoices/fiscal/:id/retry
func (h *FiscalInvoiceHandler) Retry(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid invoice id")
	}
	inv, err := h.svc.Retry(c.Request().Context(), tenantID, invoiceID, supabaseUID)
	if err != nil {
		return mapFiscalError(c, err)
	}
	return c.JSON(http.StatusOK, inv)
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func buildFiscalLineRequests(raw []createFiscalLineReq) ([]fiscal.CreateLineRequest, error) {
	lines := make([]fiscal.CreateLineRequest, 0, len(raw))
	for _, r := range raw {
		pid, err := uuid.Parse(r.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product_id: %s", r.ProductID)
		}
		lines = append(lines, fiscal.CreateLineRequest{
			ProductID:   pid,
			Description: r.Description,
			Quantity:    r.Quantity,
			UnitPrice:   r.UnitPrice,
			TaxRate:     r.TaxRate,
		})
	}
	return lines, nil
}

func mapFiscalError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, fiscal.ErrInvoiceNotFound):
		return WriteProblem(c, http.StatusNotFound, "not-found", err.Error())
	case errors.Is(err, fiscal.ErrInvoiceAlreadyIssued),
		errors.Is(err, fiscal.ErrInvoiceAlreadyCancelled),
		errors.Is(err, fiscal.ErrNonFiscalProductForbidden),
		errors.Is(err, fiscal.ErrEmptyInvoice),
		errors.Is(err, fiscal.ErrInvoiceNotDraft),
		errors.Is(err, fiscal.ErrInvoiceCannotRetry):
		return WriteProblem(c, http.StatusUnprocessableEntity, "unprocessable", err.Error())
	default:
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
}
