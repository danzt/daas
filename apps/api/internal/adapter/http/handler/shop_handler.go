package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/shop"
)

// ShopHandler handles the public storefront catalog HTTP endpoints.
// It is a thin adapter: parse → call service → format response.
// No authentication is required — PublicTenantMiddleware sets tenant context.
type ShopHandler struct {
	svc *app.ShopService
}

// NewShopHandler creates a ShopHandler backed by the given ShopService.
func NewShopHandler(svc *app.ShopService) *ShopHandler {
	return &ShopHandler{svc: svc}
}

// ----- Response types -----

type shopProductResponse struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	FiscalPrice *float64 `json:"fiscal_price,omitempty"`
	Category    *string  `json:"category,omitempty"`
	StockQty    int      `json:"stock_qty"`
	IsFiscal    bool     `json:"is_fiscal"`
	ImageURL    *string  `json:"image_url,omitempty"`
}

type paginationMeta struct {
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

type productListResponse struct {
	Data []shopProductResponse `json:"data"`
	Meta paginationMeta        `json:"meta"`
}

// ----- Helpers -----

func toShopProductResponse(p shop.ShopProduct) shopProductResponse {
	return shopProductResponse{
		ID:          p.ID.String(),
		TenantID:    p.TenantID.String(),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		FiscalPrice: p.FiscalPrice,
		Category:    p.Category,
		StockQty:    p.StockQty,
		IsFiscal:    p.IsFiscal,
		ImageURL:    p.ImageURL,
	}
}

// getPublicTenantID extracts the tenant UUID from the Echo context.
// PublicTenantMiddleware stores it under ContextKeyTenantID.
func getPublicTenantID(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get("tenant_id").(string)
	if raw == "" {
		return uuid.Nil, errors.New("tenant context missing")
	}
	return uuid.Parse(raw)
}

// ----- Handlers -----

// ListProducts handles GET /t/:tenantSlug/shop/v1/products
//
// Query params:
//   - page     integer  optional  default: 1   (min 1, else 400)
//   - per_page integer  optional  default: 20  (min 1, max 100; >100 → 400)
//   - category string   optional  exact match on category name
//   - q        string   optional  ILIKE search on name and description
//
// Response 200:
//
//	{"data": [...], "meta": {"total_count": N, "page": P, "per_page": PP, "total_pages": T}}
//
// Response 400 on invalid params:
//
//	{"error": "invalid_param", "field": "<param name>"}
func (h *ShopHandler) ListProducts(c echo.Context) error {
	tenantID, err := getPublicTenantID(c)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "tenant_context_missing",
		})
	}

	// --- Parse and validate query parameters ---

	page := 1
	if v := c.QueryParam("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid_param",
				"field": "page",
			})
		}
		page = n
	}

	perPage := 20
	if v := c.QueryParam("per_page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid_param",
				"field": "per_page",
			})
		}
		if n > 100 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid_param",
				"field": "per_page",
			})
		}
		perPage = n
	}

	filter := shop.ProductFilter{
		Page:    page,
		PerPage: perPage,
	}

	if v := c.QueryParam("category"); v != "" {
		filter.Category = &v
	}
	if v := c.QueryParam("q"); v != "" {
		filter.Q = &v
	}

	// --- Call service ---

	products, total, err := h.svc.ListPublicProducts(c.Request().Context(), tenantID, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal_error",
		})
	}

	// --- Build response ---

	data := make([]shopProductResponse, 0, len(products))
	for _, p := range products {
		data = append(data, toShopProductResponse(p))
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return c.JSON(http.StatusOK, productListResponse{
		Data: data,
		Meta: paginationMeta{
			TotalCount: total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	})
}

// GetProduct handles GET /t/:tenantSlug/shop/v1/products/:id
//
// Path param:
//   - :id  UUID of the product (must be valid UUID, else 400)
//
// Response 200: single product object
// Response 400: invalid UUID
// Response 404: product not found, inactive, or belongs to different tenant
func (h *ShopHandler) GetProduct(c echo.Context) error {
	tenantID, err := getPublicTenantID(c)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "tenant_context_missing",
		})
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "invalid_param",
			"field":   "id",
			"message": "id must be a valid UUID",
		})
	}

	p, err := h.svc.GetPublicProduct(c.Request().Context(), tenantID, productID)
	if err != nil {
		if errors.Is(err, shop.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error":   "not_found",
				"message": "Product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "internal_error",
		})
	}

	return c.JSON(http.StatusOK, toShopProductResponse(*p))
}
