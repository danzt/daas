package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/danzt/daas/api/internal/adapter/http/middleware"
	"github.com/danzt/daas/api/internal/app"
	"github.com/danzt/daas/api/internal/domain/product"
)

// ProductHandler handles the product catalog HTTP endpoints.
// It is a thin adapter: parse → call service → format response.
type ProductHandler struct {
	service *app.ProductService
}

// NewProductHandler creates a ProductHandler backed by the given service.
func NewProductHandler(pool *pgxpool.Pool) *ProductHandler {
	return &ProductHandler{service: app.NewProductService(pool)}
}

// ----- Request / Response types -----

type createProductRequest struct {
	Name          string     `json:"name"`
	SKU           string     `json:"sku,omitempty"`
	Barcode       string     `json:"barcode,omitempty"`
	Description   string     `json:"description,omitempty"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty"`
	IsFiscal      bool       `json:"is_fiscal"`
	FiscalPrice   *float64   `json:"fiscal_price,omitempty"`
	InternalPrice *float64   `json:"internal_price,omitempty"`
	TaxRate       *float64   `json:"tax_rate,omitempty"`
}

type updateProductRequest struct {
	Name          string     `json:"name"`
	SKU           string     `json:"sku,omitempty"`
	Barcode       string     `json:"barcode,omitempty"`
	Description   string     `json:"description,omitempty"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty"`
	FiscalPrice   *float64   `json:"fiscal_price,omitempty"`
	InternalPrice *float64   `json:"internal_price,omitempty"`
	TaxRate       *float64   `json:"tax_rate,omitempty"`
}

type createCategoryRequest struct {
	Name     string     `json:"name"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

type productResponse struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Name          string     `json:"name"`
	SKU           string     `json:"sku,omitempty"`
	Barcode       string     `json:"barcode,omitempty"`
	Description   string     `json:"description,omitempty"`
	CategoryID    *uuid.UUID `json:"category_id,omitempty"`
	IsFiscal      bool       `json:"is_fiscal"`
	FiscalPrice   *float64   `json:"fiscal_price,omitempty"`
	InternalPrice *float64   `json:"internal_price,omitempty"`
	TaxRate       *float64   `json:"tax_rate,omitempty"`
	Active        bool       `json:"active"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
}

type productResultResponse struct {
	Product  productResponse `json:"product"`
	Warnings []string        `json:"warnings,omitempty"`
}

type categoryResponse struct {
	ID        string     `json:"id"`
	TenantID  string     `json:"tenant_id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt string     `json:"created_at"`
}

// ----- Helpers -----

func toProductResponse(p *product.Product) productResponse {
	return productResponse{
		ID:            p.ID.String(),
		TenantID:      p.TenantID.String(),
		Name:          p.Name,
		SKU:           p.SKU,
		Barcode:       p.Barcode,
		Description:   p.Description,
		CategoryID:    p.CategoryID,
		IsFiscal:      p.IsFiscal,
		FiscalPrice:   p.FiscalPrice,
		InternalPrice: p.InternalPrice,
		TaxRate:       p.TaxRate,
		Active:        p.Active,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}

func mapProductError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, product.ErrFiscalPriceMissing):
		return WriteProblem(c, http.StatusUnprocessableEntity, "fiscal-price-missing",
			"fiscal product must have a positive fiscal_price")
	case errors.Is(err, product.ErrInternalPriceMissing):
		return WriteProblem(c, http.StatusUnprocessableEntity, "internal-price-missing",
			"internal product must have a positive internal_price")
	case errors.Is(err, product.ErrTaxRateMissing):
		return WriteProblem(c, http.StatusUnprocessableEntity, "tax-rate-missing",
			"fiscal product must have a tax_rate")
	case errors.Is(err, product.ErrTaxRateOnInternal):
		return WriteProblem(c, http.StatusUnprocessableEntity, "tax-rate-on-internal",
			"internal product must not have a tax_rate")
	case errors.Is(err, product.ErrZeroPrice):
		return WriteProblem(c, http.StatusUnprocessableEntity, "zero-price", "price must be greater than zero")
	case errors.Is(err, product.ErrProductNotFound):
		return WriteProblem(c, http.StatusNotFound, "product-not-found", "product not found in this tenant")
	case errors.Is(err, product.ErrDuplicateSKU):
		return WriteProblem(c, http.StatusConflict, "duplicate-sku", "a product with this SKU already exists")
	default:
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "an unexpected error occurred")
	}
}

func getTenantID(c echo.Context) (uuid.UUID, error) {
	raw, _ := c.Get(string(middleware.ContextKeyTenantID)).(string)
	if raw == "" {
		return uuid.Nil, errors.New("tenant context missing")
	}
	return uuid.Parse(raw)
}

// ----- Handlers -----

// ListProducts handles GET /api/v1/products
// Query params: active, is_fiscal, category_id, q, page, limit
func (h *ProductHandler) ListProducts(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	filters := app.ProductFilters{}

	if v := c.QueryParam("active"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return WriteProblem(c, http.StatusBadRequest, "bad-request", "active must be true or false")
		}
		filters.Active = &b
	}
	if v := c.QueryParam("is_fiscal"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return WriteProblem(c, http.StatusBadRequest, "bad-request", "is_fiscal must be true or false")
		}
		filters.IsFiscal = &b
	}
	if v := c.QueryParam("category_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return WriteProblem(c, http.StatusBadRequest, "bad-request", "category_id must be a valid UUID")
		}
		filters.CategoryID = &id
	}
	filters.Query = c.QueryParam("q")

	if v := c.QueryParam("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filters.Page = n
		}
	}
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filters.Limit = n
		}
	}

	products, err := h.service.List(c.Request().Context(), tenantID, filters)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to list products")
	}

	resp := make([]productResponse, 0, len(products))
	for _, p := range products {
		resp = append(resp, toProductResponse(p))
	}
	return c.JSON(http.StatusOK, resp)
}

// GetProduct handles GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "id must be a valid UUID")
	}

	p, err := h.service.Get(c.Request().Context(), tenantID, id)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(http.StatusOK, toProductResponse(p))
}

// CreateProduct handles POST /api/v1/products
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	var req createProductRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	if req.Name == "" {
		return WriteProblem(c, http.StatusUnprocessableEntity, "validation-error", "name is required")
	}

	result, err := h.service.Create(c.Request().Context(), tenantID, app.CreateProductRequest{
		Name:          req.Name,
		SKU:           req.SKU,
		Barcode:       req.Barcode,
		Description:   req.Description,
		CategoryID:    req.CategoryID,
		IsFiscal:      req.IsFiscal,
		FiscalPrice:   req.FiscalPrice,
		InternalPrice: req.InternalPrice,
		TaxRate:       req.TaxRate,
	})
	if err != nil {
		return mapProductError(c, err)
	}

	return c.JSON(http.StatusCreated, productResultResponse{
		Product:  toProductResponse(result.Product),
		Warnings: result.Warnings,
	})
}

// UpdateProduct handles PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "id must be a valid UUID")
	}

	var req updateProductRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	if req.Name == "" {
		return WriteProblem(c, http.StatusUnprocessableEntity, "validation-error", "name is required")
	}

	p, err := h.service.Update(c.Request().Context(), tenantID, id, app.UpdateProductRequest{
		Name:          req.Name,
		SKU:           req.SKU,
		Barcode:       req.Barcode,
		Description:   req.Description,
		CategoryID:    req.CategoryID,
		FiscalPrice:   req.FiscalPrice,
		InternalPrice: req.InternalPrice,
		TaxRate:       req.TaxRate,
	})
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(http.StatusOK, toProductResponse(p))
}

// DeleteProduct handles DELETE /api/v1/products/:id (soft delete)
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "id must be a valid UUID")
	}

	p, err := h.service.Delete(c.Request().Context(), tenantID, id)
	if err != nil {
		return mapProductError(c, err)
	}
	return c.JSON(http.StatusOK, toProductResponse(p))
}

// ListCategories handles GET /api/v1/products/categories
func (h *ProductHandler) ListCategories(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	cats, err := h.service.ListCategories(c.Request().Context(), tenantID)
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to list categories")
	}

	resp := make([]categoryResponse, 0, len(cats))
	for _, cat := range cats {
		resp = append(resp, categoryResponse{
			ID:        cat.ID.String(),
			TenantID:  cat.TenantID.String(),
			Name:      cat.Name,
			ParentID:  cat.ParentID,
			CreatedAt: cat.CreatedAt.Format(time.RFC3339),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateCategory handles POST /api/v1/products/categories
func (h *ProductHandler) CreateCategory(c echo.Context) error {
	tenantID, err := getTenantID(c)
	if err != nil {
		return WriteProblem(c, http.StatusForbidden, "forbidden", "tenant context missing")
	}

	var req createCategoryRequest
	if err := c.Bind(&req); err != nil {
		return WriteProblem(c, http.StatusBadRequest, "bad-request", "invalid JSON body")
	}
	if req.Name == "" {
		return WriteProblem(c, http.StatusUnprocessableEntity, "validation-error", "name is required")
	}

	cat, err := h.service.CreateCategory(c.Request().Context(), tenantID, app.CreateCategoryRequest{
		Name:     req.Name,
		ParentID: req.ParentID,
	})
	if err != nil {
		return WriteProblem(c, http.StatusInternalServerError, "internal-error", "failed to create category")
	}

	return c.JSON(http.StatusCreated, categoryResponse{
		ID:        cat.ID.String(),
		TenantID:  cat.TenantID.String(),
		Name:      cat.Name,
		ParentID:  cat.ParentID,
		CreatedAt: cat.CreatedAt.Format(time.RFC3339),
	})
}
