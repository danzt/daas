package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/adapter/http/middleware"
	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/inventory"
)

// InventoryHandler handles inventory control HTTP endpoints.
type InventoryHandler struct {
	service *app.InventoryService
}

// NewInventoryHandler creates an InventoryHandler backed by the given pool.
func NewInventoryHandler(pool *pgxpool.Pool) *InventoryHandler {
	return &InventoryHandler{service: app.NewInventoryService(pool)}
}

// ─── Response types ───────────────────────────────────────────────────────────

type stockResponse struct {
	ProductID      string  `json:"product_id"`
	TenantID       string  `json:"tenant_id"`
	QuantityOnHand float64 `json:"quantity_on_hand"`
	LastUpdatedAt  string  `json:"last_updated_at"`
}

type movementResponse struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	ProductID     string     `json:"product_id"`
	Type          string     `json:"type"`
	Quantity      float64    `json:"quantity"`
	UnitCost      *float64   `json:"unit_cost,omitempty"`
	ReferenceType string     `json:"reference_type"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	CreatedBy     string     `json:"created_by"`
	CreatedAt     string     `json:"created_at"`
}

type adjustmentRequest struct {
	ProductID string  `json:"product_id"`
	Delta     float64 `json:"delta"`
	Notes     string  `json:"notes"`
}

func toStockResponse(st *inventory.Stock) stockResponse {
	return stockResponse{
		ProductID:      st.ProductID.String(),
		TenantID:       st.TenantID.String(),
		QuantityOnHand: st.QuantityOnHand,
		LastUpdatedAt:  st.LastUpdatedAt.Format(time.RFC3339),
	}
}

func toMovementResponse(m *inventory.Movement) movementResponse {
	return movementResponse{
		ID:            m.ID.String(),
		TenantID:      m.TenantID.String(),
		ProductID:     m.ProductID.String(),
		Type:          string(m.Type),
		Quantity:      m.Quantity,
		UnitCost:      m.UnitCost,
		ReferenceType: string(m.ReferenceType),
		ReferenceID:   m.ReferenceID,
		Notes:         m.Notes,
		CreatedBy:     m.CreatedBy.String(),
		CreatedAt:     m.CreatedAt.Format(time.RFC3339),
	}
}

func getSupabaseUID(c echo.Context) (string, error) {
	uid, _ := c.Get(string(middleware.ContextKeySupabaseUID)).(string)
	if uid == "" {
		return "", errors.New("supabase_uid missing from context")
	}
	return uid, nil
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// ListStock handles GET /api/v1/inventory/stock
func (h *InventoryHandler) ListStock(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	stocks, err := h.service.ListStock(c.Request().Context(), tenantID)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to list stock")
	}

	resp := make([]stockResponse, 0, len(stocks))
	for _, st := range stocks {
		resp = append(resp, toStockResponse(st))
	}
	return c.JSON(http.StatusOK, resp)
}

// GetProductStock handles GET /api/v1/inventory/stock/:product_id
func (h *InventoryHandler) GetProductStock(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "product_id must be a valid UUID")
	}

	st, err := h.service.GetProductStock(c.Request().Context(), tenantID, productID)
	if err != nil {
		if errors.Is(err, inventory.ErrStockNotFound) {
			return WriteProblem(c, http.StatusNotFound, "not-found", "stock record not found")
		}
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to get stock")
	}
	return c.JSON(http.StatusOK, toStockResponse(st))
}

// ListMovements handles GET /api/v1/inventory/movements
// Query params: product_id, type, from, to, page, limit
func (h *InventoryHandler) ListMovements(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	f := app.MovementFilters{}

	if v := c.QueryParam("product_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return WriteProblem(c, http.StatusBadRequest, "bad-request", "product_id must be a valid UUID")
		}
		f.ProductID = &id
	}
	f.Type = c.QueryParam("type")

	if v := c.QueryParam("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return WriteProblem(c, http.StatusBadRequest, "bad-request", "from must be RFC3339 datetime")
		}
		f.From = &t
	}
	if v := c.QueryParam("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return WriteProblem(c, http.StatusBadRequest, "bad-request", "to must be RFC3339 datetime")
		}
		f.To = &t
	}

	movements, err := h.service.ListMovements(c.Request().Context(), tenantID, f)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to list movements")
	}

	resp := make([]movementResponse, 0, len(movements))
	for _, m := range movements {
		resp = append(resp, toMovementResponse(m))
	}
	return c.JSON(http.StatusOK, resp)
}

// GetMovement handles GET /api/v1/inventory/movements/:id
func (h *InventoryHandler) GetMovement(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "id must be a valid UUID")
	}

	m, err := h.service.GetMovement(c.Request().Context(), tenantID, id)
	if err != nil {
		if errors.Is(err, inventory.ErrMovementNotFound) {
			return WriteProblem(c, http.StatusNotFound, "not-found", "movement not found")
		}
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to get movement")
	}
	return c.JSON(http.StatusOK, toMovementResponse(m))
}

// CreateAdjustment handles POST /api/v1/inventory/adjustments
func (h *InventoryHandler) CreateAdjustment(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	supabaseUID, err := getSupabaseUID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "user context missing")
	}

	var req adjustmentRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "product_id must be a valid UUID")
	}

	m, err := h.service.CreateAdjustment(c.Request().Context(), tenantID, supabaseUID, inventory.AdjustmentRequest{
		ProductID: productID,
		Delta:     req.Delta,
		Notes:     req.Notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, inventory.ErrAdjustmentNotesMissing):
			return WriteProblem(c, http.StatusUnprocessableEntity, "notes-required", "adjustment notes are required")
		case errors.Is(err, inventory.ErrZeroDelta):
			return WriteProblem(c, http.StatusUnprocessableEntity, "zero-delta", "adjustment delta must not be zero")
		case errors.Is(err, inventory.ErrInsufficientStock):
			return WriteProblem(c, http.StatusUnprocessableEntity, "insufficient-stock",
				"adjustment would result in negative stock")
		case errors.Is(err, inventory.ErrStockNotFound):
			return WriteProblem(c, http.StatusNotFound, "not-found", "stock record not found for this product")
		default:
			return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to create adjustment")
		}
	}

	return c.JSON(http.StatusCreated, toMovementResponse(m))
}
