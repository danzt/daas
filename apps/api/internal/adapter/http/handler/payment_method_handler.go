package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/tenant"
)

// PaymentMethodHandler exposes HTTP endpoints for tenant payment method management.
type PaymentMethodHandler struct {
	svc *app.PaymentMethodService
}

// NewPaymentMethodHandler creates a PaymentMethodHandler backed by the given pool.
func NewPaymentMethodHandler(pool *pgxpool.Pool) *PaymentMethodHandler {
	return &PaymentMethodHandler{svc: app.NewPaymentMethodService(pool)}
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type createPaymentMethodReq struct {
	Type      string         `json:"type"`
	Label     string         `json:"label"`
	Details   map[string]any `json:"details"`
	Currency  string         `json:"currency"`
	SortOrder int            `json:"sort_order"`
}

type updatePaymentMethodReq struct {
	Label     *string         `json:"label"`
	Details   *map[string]any `json:"details"`
	Currency  *string         `json:"currency"`
	Active    *bool           `json:"active"`
	SortOrder *int            `json:"sort_order"`
}

// ─── Admin endpoints ──────────────────────────────────────────────────────────

// AdminList handles GET /api/v1/payment-methods
//
// Returns all payment methods for the authenticated tenant (including inactive).
func (h *PaymentMethodHandler) AdminList(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	methods, err := h.svc.List(c.Request().Context(), tenantID)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	return c.JSON(http.StatusOK, methods)
}

// AdminCreate handles POST /api/v1/payment-methods
func (h *PaymentMethodHandler) AdminCreate(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	var body createPaymentMethodReq
	if err := c.Bind(&body); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	if body.Details == nil {
		body.Details = map[string]any{}
	}

	pm, err := h.svc.Create(c.Request().Context(), tenantID, tenant.CreatePaymentMethodRequest{
		Type:      tenant.PaymentMethodType(body.Type),
		Label:     body.Label,
		Details:   body.Details,
		Currency:  body.Currency,
		SortOrder: body.SortOrder,
	})
	if err != nil {
		return mapPaymentMethodError(c, err)
	}
	return c.JSON(http.StatusCreated, pm)
}

// AdminUpdate handles PUT /api/v1/payment-methods/:id
func (h *PaymentMethodHandler) AdminUpdate(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid payment method id")
	}

	var body updatePaymentMethodReq
	if err := c.Bind(&body); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}

	pm, err := h.svc.Update(c.Request().Context(), tenantID, id, tenant.UpdatePaymentMethodRequest{
		Label:     body.Label,
		Details:   body.Details,
		Currency:  body.Currency,
		Active:    body.Active,
		SortOrder: body.SortOrder,
	})
	if err != nil {
		return mapPaymentMethodError(c, err)
	}
	return c.JSON(http.StatusOK, pm)
}

// AdminDelete handles DELETE /api/v1/payment-methods/:id
func (h *PaymentMethodHandler) AdminDelete(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid payment method id")
	}

	if err := h.svc.Delete(c.Request().Context(), tenantID, id); err != nil {
		return mapPaymentMethodError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// ─── Public storefront endpoint ───────────────────────────────────────────────

// PublicList handles GET /t/:tenantSlug/shop/v1/payment-methods
//
// No auth required. Returns only active payment methods ordered by sort_order.
func (h *PaymentMethodHandler) PublicList(c echo.Context) error {
	tenantID, err := getPublicTenantID(c)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "tenant_context_missing"})
	}

	methods, err := h.svc.ListActiveForStorefront(c.Request().Context(), tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal_error"})
	}
	return c.JSON(http.StatusOK, methods)
}

// ─── Error mapper ─────────────────────────────────────────────────────────────

func mapPaymentMethodError(c echo.Context, err error) error {
	var mfe *tenant.MissingFieldError
	switch {
	case errors.Is(err, tenant.ErrPaymentMethodNotFound):
		return WriteProblem(c, http.StatusNotFound, "not-found", "payment method not found")
	case errors.Is(err, tenant.ErrInvalidPaymentMethodType):
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid payment method type")
	case errors.As(err, &mfe):
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "validation_error",
			"field":   mfe.Field,
			"message": mfe.Error(),
		})
	default:
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
}
