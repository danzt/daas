// Package shop contains public-facing read-model types for the storefront catalog.
// These are DTOs (not aggregates) — no domain methods, no mutations in S6.
package shop

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrProductNotFound is returned when a product cannot be found for the given
// tenant and product ID, or when the product is inactive.
var ErrProductNotFound = errors.New("shop product not found")

// ShopProduct is the public-safe projection of a product.
// Sensitive fields (cost_price, supplier_id, fiscal_id, created_by) MUST NOT
// be included — they are intentionally absent from this struct.
type ShopProduct struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	FiscalPrice *float64  `json:"fiscal_price,omitempty"`
	Category    *string   `json:"category,omitempty"`
	StockQty    int       `json:"stock_qty"`
	IsFiscal    bool      `json:"is_fiscal"`
	ImageURL    *string   `json:"image_url,omitempty"`
}

// ProductFilter holds query parameters for ListPublicProducts.
type ProductFilter struct {
	Page     int
	PerPage  int
	Category *string
	Q        *string
}

// ShopProductRepository is the port for reading the public catalog.
// Implementations must enforce RLS-level tenant isolation.
type ShopProductRepository interface {
	ListPublicProducts(ctx context.Context, tenantID uuid.UUID, filter ProductFilter) (products []ShopProduct, total int, err error)
	GetPublicProduct(ctx context.Context, tenantID, productID uuid.UUID) (*ShopProduct, error)
}
