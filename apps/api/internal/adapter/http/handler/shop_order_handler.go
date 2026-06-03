package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/shop"
)

// ShopOrderHandler exposes HTTP endpoints for the public storefront checkout
// flow and the tenant admin order lifecycle.
type ShopOrderHandler struct {
	svc *app.ShopOrderService
}

// NewShopOrderHandler creates a ShopOrderHandler backed by the given pool.
func NewShopOrderHandler(pool *pgxpool.Pool) *ShopOrderHandler {
	return &ShopOrderHandler{svc: app.NewShopOrderService(pool)}
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type checkoutLineReq struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type checkoutReq struct {
	CustomerName    string            `json:"customer_name"`
	CustomerEmail   string            `json:"customer_email"`
	CustomerPhone   string            `json:"customer_phone"`
	ShippingAddress string            `json:"shipping_address"`
	ShippingCity    string            `json:"shipping_city"`
	ShippingNotes   string            `json:"shipping_notes"`
	Notes           string            `json:"notes"`
	Lines           []checkoutLineReq `json:"lines"`
}

// ─── Public endpoints ─────────────────────────────────────────────────────────

// Checkout handles POST /t/:tenantSlug/shop/v1/checkout
//
// Public. No auth required. Tenant comes from PublicTenantMiddleware context.
// Returns 201 with the full order JSON including the one-time access_token.
func (h *ShopOrderHandler) Checkout(c echo.Context) error {
	tenantID, err := getPublicTenantID(c)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "tenant_context_missing"})
	}

	var body checkoutReq
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_body"})
	}

	lines, err := buildCheckoutLines(body.Lines)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_param",
			"message": err.Error(),
		})
	}

	order, err := h.svc.Checkout(c.Request().Context(), tenantID, shop.CreateOrderRequest{
		CustomerName:    body.CustomerName,
		CustomerEmail:   body.CustomerEmail,
		CustomerPhone:   body.CustomerPhone,
		ShippingAddress: body.ShippingAddress,
		ShippingCity:    body.ShippingCity,
		ShippingNotes:   body.ShippingNotes,
		Notes:           body.Notes,
		Lines:           lines,
	})
	if err != nil {
		return mapShopOrderError(c, err)
	}
	return c.JSON(http.StatusCreated, order)
}

// GetOrder handles GET /t/:tenantSlug/shop/v1/orders/:id?access_token=...
//
// Public. Validates access_token against the order.
// Returns 404 if not found or token mismatch (intentionally opaque).
// The response does NOT include the access_token field.
func (h *ShopOrderHandler) GetOrder(c echo.Context) error {
	tenantID, err := getPublicTenantID(c)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "tenant_context_missing"})
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid_param",
			"field": "id",
		})
	}

	token := c.QueryParam("access_token")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "access_token_required"})
	}

	order, err := h.svc.GetPublicOrder(c.Request().Context(), tenantID, orderID, token)
	if err != nil {
		return mapShopOrderError(c, err)
	}
	// access_token is already cleared by the service; ensure it's not leaked.
	order.AccessToken = ""
	return c.JSON(http.StatusOK, order)
}

// CancelOrder handles POST /t/:tenantSlug/shop/v1/orders/:id/cancel?access_token=...
//
// Public. Customer can cancel if the order status is pending.
func (h *ShopOrderHandler) CancelOrder(c echo.Context) error {
	tenantID, err := getPublicTenantID(c)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "tenant_context_missing"})
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid_param",
			"field": "id",
		})
	}

	token := c.QueryParam("access_token")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "access_token_required"})
	}

	order, err := h.svc.CustomerCancel(c.Request().Context(), tenantID, orderID, token)
	if err != nil {
		return mapShopOrderError(c, err)
	}
	order.AccessToken = ""
	return c.JSON(http.StatusOK, order)
}

// ─── Tenant admin endpoints ───────────────────────────────────────────────────

// AdminList handles GET /api/v1/shop-orders
//
// Auth required. Lists orders for the authenticated tenant.
// Optional query params: status, limit (default 20), offset (default 0).
func (h *ShopOrderHandler) AdminList(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	var status *shop.OrderStatus
	if s := c.QueryParam("status"); s != "" {
		st := shop.OrderStatus(s)
		status = &st
	}

	limit := 20
	if v := c.QueryParam("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			return WriteProblem(c, http.StatusBadRequest, "bad-request", "limit must be 1-100")
		}
		limit = n
	}

	offset := 0
	if v := c.QueryParam("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return WriteProblem(c, http.StatusBadRequest, "bad-request", "offset must be >= 0")
		}
		offset = n
	}

	orders, total, err := h.svc.ListTenantOrders(c.Request().Context(), tenantID, status, limit, offset)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  orders,
		"total": total,
	})
}

// AdminGet handles GET /api/v1/shop-orders/:id
func (h *ShopOrderHandler) AdminGet(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid order id")
	}

	order, err := h.svc.GetAdminOrder(c.Request().Context(), tenantID, orderID)
	if err != nil {
		return mapShopOrderError(c, err)
	}
	order.AccessToken = ""
	return c.JSON(http.StatusOK, order)
}

// MarkPaid handles POST /api/v1/shop-orders/:id/mark-paid
func (h *ShopOrderHandler) MarkPaid(c echo.Context) error {
	return h.adminTransition(c, func(tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
		return h.svc.MarkPaid(c.Request().Context(), tenantID, orderID)
	})
}

// MarkFulfilled handles POST /api/v1/shop-orders/:id/mark-fulfilled
func (h *ShopOrderHandler) MarkFulfilled(c echo.Context) error {
	return h.adminTransition(c, func(tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
		return h.svc.MarkFulfilled(c.Request().Context(), tenantID, orderID)
	})
}

// MarkDelivered handles POST /api/v1/shop-orders/:id/mark-delivered
func (h *ShopOrderHandler) MarkDelivered(c echo.Context) error {
	return h.adminTransition(c, func(tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
		return h.svc.MarkDelivered(c.Request().Context(), tenantID, orderID)
	})
}

// AdminCancel handles POST /api/v1/shop-orders/:id/cancel (admin version, no token)
func (h *ShopOrderHandler) AdminCancel(c echo.Context) error {
	return h.adminTransition(c, func(tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
		return h.svc.Cancel(c.Request().Context(), tenantID, orderID)
	})
}

// ─── Private helpers ──────────────────────────────────────────────────────────

type transitionFn func(tenantID, orderID uuid.UUID) (*shop.ShopOrder, error)

func (h *ShopOrderHandler) adminTransition(c echo.Context, fn transitionFn) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid order id")
	}
	order, err := fn(tenantID, orderID)
	if err != nil {
		return mapShopOrderError(c, err)
	}
	order.AccessToken = ""
	return c.JSON(http.StatusOK, order)
}

func buildCheckoutLines(raw []checkoutLineReq) ([]shop.CreateOrderLineRequest, error) {
	lines := make([]shop.CreateOrderLineRequest, 0, len(raw))
	for _, r := range raw {
		pid, err := uuid.Parse(r.ProductID)
		if err != nil {
			return nil, errors.New("invalid product_id: " + r.ProductID)
		}
		lines = append(lines, shop.CreateOrderLineRequest{
			ProductID: pid,
			Quantity:  r.Quantity,
		})
	}
	return lines, nil
}

func mapShopOrderError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, shop.ErrShopOrderNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{
			"error":   "not_found",
			"message": "Order not found",
		})
	case errors.Is(err, shop.ErrInsufficientStock):
		return c.JSON(http.StatusConflict, map[string]string{
			"error":   "insufficient_stock",
			"message": err.Error(),
		})
	case errors.Is(err, shop.ErrProductNotPurchasable):
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error":   "product_not_purchasable",
			"message": err.Error(),
		})
	case errors.Is(err, shop.ErrCustomerNameRequired),
		errors.Is(err, shop.ErrCustomerEmailRequired),
		errors.Is(err, shop.ErrShippingAddressRequired),
		errors.Is(err, shop.ErrEmptyOrder),
		errors.Is(err, shop.ErrInvalidQuantity):
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "validation_error",
			"message": err.Error(),
		})
	case errors.Is(err, shop.ErrInvalidTransition):
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"error":   "invalid_transition",
			"message": err.Error(),
		})
	case errors.Is(err, shop.ErrInvalidAccessToken):
		return c.JSON(http.StatusNotFound, map[string]string{
			"error":   "not_found",
			"message": "Order not found",
		})
	default:
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", err.Error())
	}
}
