package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/sale"
)

// SaleHandler exposes HTTP endpoints for the sales order lifecycle.
type SaleHandler struct {
	svc *app.SaleService
}

func NewSaleHandler(pool *pgxpool.Pool) *SaleHandler {
	return &SaleHandler{svc: app.NewSaleService(pool)}
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type createSaleLineReq struct {
	ProductID   string  `json:"product_id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type createSaleOrderReq struct {
	CustomerName     string              `json:"customer_name"`
	CustomerIDType   string              `json:"customer_id_type"`
	CustomerIDNumber string              `json:"customer_id_number"`
	Notes            string              `json:"notes"`
	Lines            []createSaleLineReq `json:"lines"`
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

func (h *SaleHandler) CreateOrder(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	var body createSaleOrderReq
	if err := c.Bind(&body); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid body")
	}
	lines, err := buildSaleLines(body.Lines)
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", err.Error())
	}
	order, err := h.svc.CreateOrder(c.Request().Context(), tenantID, supabaseUID, sale.CreateOrderRequest{
		CustomerName:     body.CustomerName,
		CustomerIDType:   body.CustomerIDType,
		CustomerIDNumber: body.CustomerIDNumber,
		Notes:            body.Notes,
		Lines:            lines,
	})
	if err != nil {
		return mapSaleError(c, err)
	}
	return c.JSON(http.StatusCreated, order)
}

func (h *SaleHandler) ListOrders(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	var status *sale.OrderStatus
	if s := c.QueryParam("status"); s != "" {
		st := sale.OrderStatus(s)
		status = &st
	}
	orders, err := h.svc.ListOrders(c.Request().Context(), tenantID, status)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	if orders == nil {
		orders = []*sale.SaleOrder{}
	}
	return c.JSON(http.StatusOK, orders)
}

func (h *SaleHandler) GetOrder(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid order id")
	}
	order, err := h.svc.GetOrder(c.Request().Context(), tenantID, orderID)
	if err != nil {
		return mapSaleError(c, err)
	}
	return c.JSON(http.StatusOK, order)
}

func (h *SaleHandler) ConfirmOrder(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid order id")
	}
	order, err := h.svc.ConfirmOrder(c.Request().Context(), tenantID, orderID, supabaseUID)
	if err != nil {
		return mapSaleError(c, err)
	}
	return c.JSON(http.StatusOK, order)
}

func (h *SaleHandler) InvoiceOrder(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid order id")
	}
	order, err := h.svc.InvoiceOrder(c.Request().Context(), tenantID, orderID, supabaseUID)
	if err != nil {
		return mapSaleError(c, err)
	}
	return c.JSON(http.StatusOK, order)
}

func (h *SaleHandler) CancelOrder(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusUnauthorized, "unauthorized", "missing user context")
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid order id")
	}
	order, err := h.svc.CancelOrder(c.Request().Context(), tenantID, orderID, supabaseUID)
	if err != nil {
		return mapSaleError(c, err)
	}
	return c.JSON(http.StatusOK, order)
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func buildSaleLines(raw []createSaleLineReq) ([]sale.CreateOrderLineRequest, error) {
	lines := make([]sale.CreateOrderLineRequest, 0, len(raw))
	for _, r := range raw {
		pid, err := uuid.Parse(r.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product_id: %s", r.ProductID)
		}
		lines = append(lines, sale.CreateOrderLineRequest{
			ProductID:   pid,
			Description: r.Description,
			Quantity:    r.Quantity,
			UnitPrice:   r.UnitPrice,
		})
	}
	return lines, nil
}

func mapSaleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, sale.ErrOrderNotFound):
		return WriteProblem(c, http.StatusNotFound, "not-found", err.Error())
	case errors.Is(err, sale.ErrEmptyOrder),
		errors.Is(err, sale.ErrOrderAlreadyConfirmed),
		errors.Is(err, sale.ErrOrderAlreadyInvoiced),
		errors.Is(err, sale.ErrOrderAlreadyCancelled),
		errors.Is(err, sale.ErrOrderNotConfirmed),
		errors.Is(err, sale.ErrOrderNotDraft),
		errors.Is(err, sale.ErrInsufficientStock):
		return WriteProblem(c, http.StatusUnprocessableEntity, "unprocessable", err.Error())
	default:
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
}
