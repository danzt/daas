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

	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/invoice"
)

// InvoiceHandler exposes HTTP endpoints for internal invoicing.
type InvoiceHandler struct {
	svc *app.InternalInvoiceService
}

// NewInvoiceHandler creates an InvoiceHandler backed by the given pool.
func NewInvoiceHandler(pool *pgxpool.Pool) *InvoiceHandler {
	return &InvoiceHandler{svc: app.NewInternalInvoiceService(pool)}
}

// ─── Request / Response DTOs ─────────────────────────────────────────────────

type createInvoiceLineReq struct {
	ProductID   string  `json:"product_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type createInvoiceReq struct {
	CustomerName     string                 `json:"customer_name"`
	CustomerIDType   invoice.CustomerIDType `json:"customer_id_type"`
	CustomerIDNumber string                 `json:"customer_id_number"`
	Notes            string                 `json:"notes"`
	Lines            []createInvoiceLineReq `json:"lines"`
}

type cancelInvoiceReq struct {
	Notes string `json:"notes"`
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// Create handles POST /api/v1/invoices/internal
func (h *InvoiceHandler) Create(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}

	var body createInvoiceReq
	if err := c.Bind(&body); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid request body")
	}

	lines, err := buildLineRequests(body.Lines)
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", err.Error())
	}

	req := invoice.CreateRequest{
		CustomerName:     body.CustomerName,
		CustomerIDType:   body.CustomerIDType,
		CustomerIDNumber: body.CustomerIDNumber,
		Notes:            body.Notes,
		Lines:            lines,
	}

	inv, err := h.svc.Create(c.Request().Context(), tenantID, supabaseUID, req)
	if err != nil {
		return mapInvoiceError(c, err)
	}
	return c.JSON(http.StatusCreated, inv)
}

// List handles GET /api/v1/invoices/internal
func (h *InvoiceHandler) List(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	f := app.InvoiceListFilters{
		Limit:  50,
		Offset: 0,
	}

	if s := c.QueryParam("status"); s != "" {
		st := invoice.Status(s)
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
		invoices = []*invoice.InternalInvoice{}
	}
	return c.JSON(http.StatusOK, invoices)
}

// GetByID handles GET /api/v1/invoices/internal/:id
func (h *InvoiceHandler) GetByID(c echo.Context) error {
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
		return mapInvoiceError(c, err)
	}
	return c.JSON(http.StatusOK, inv)
}

// Issue handles POST /api/v1/invoices/internal/:id/issue
func (h *InvoiceHandler) Issue(c echo.Context) error {
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
		return mapInvoiceError(c, err)
	}
	return c.JSON(http.StatusOK, inv)
}

// Cancel handles POST /api/v1/invoices/internal/:id/cancel
func (h *InvoiceHandler) Cancel(c echo.Context) error {
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

	var body cancelInvoiceReq
	_ = c.Bind(&body)

	inv, err := h.svc.Cancel(c.Request().Context(), tenantID, invoiceID, supabaseUID, body.Notes)
	if err != nil {
		return mapInvoiceError(c, err)
	}
	return c.JSON(http.StatusOK, inv)
}

// ─── private helpers ─────────────────────────────────────────────────────────

func buildLineRequests(raw []createInvoiceLineReq) ([]invoice.CreateLineRequest, error) {
	lines := make([]invoice.CreateLineRequest, 0, len(raw))
	for _, r := range raw {
		pid, err := uuid.Parse(r.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product_id: %s", r.ProductID)
		}
		lines = append(lines, invoice.CreateLineRequest{
			ProductID:   pid,
			Description: r.Description,
			Quantity:    r.Quantity,
			UnitPrice:   r.UnitPrice,
		})
	}
	return lines, nil
}

func mapInvoiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, invoice.ErrInvoiceNotFound):
		return WriteProblem(c, http.StatusNotFound, "not-found", err.Error())
	case errors.Is(err, invoice.ErrInvoiceAlreadyIssued),
		errors.Is(err, invoice.ErrInvoiceAlreadyCancelled),
		errors.Is(err, invoice.ErrFiscalProductForbidden),
		errors.Is(err, invoice.ErrEmptyInvoice),
		errors.Is(err, invoice.ErrInvoiceNotDraft):
		return WriteProblem(c, http.StatusUnprocessableEntity, "unprocessable", err.Error())
	default:
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
}
