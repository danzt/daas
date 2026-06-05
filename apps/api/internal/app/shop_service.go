package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danzt/daas/api/internal/domain/shop"
)

// ShopService provides public-facing catalog read operations.
// It implements the shop.ShopProductRepository port directly, using the pool
// to query the RLS-activated connection provided by PublicTenantMiddleware.
//
// IMPORTANT: SQL queries in this service MUST select only public-safe columns:
//   - id, tenant_id, name, description, price (COALESCE), fiscal_price, category, stock_qty, is_fiscal
//   - NEVER: internal_price, cost_price, supplier_id, fiscal_id, created_by, sku, barcode
type ShopService struct {
	pool *pgxpool.Pool
}

// NewShopService creates a ShopService backed by the given connection pool.
func NewShopService(pool *pgxpool.Pool) *ShopService {
	return &ShopService{pool: pool}
}

// ListPublicProducts returns a paginated list of active products for the given
// tenant. Category and full-text (ILIKE) filters are applied when set.
// RLS must already be active on the connection for this query to be tenant-scoped.
func (s *ShopService) ListPublicProducts(
	ctx context.Context,
	tenantID uuid.UUID,
	filter shop.ProductFilter,
) ([]shop.ShopProduct, int, error) {
	// Count query — mirrors the WHERE clause of the data query.
	var total int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		WHERE p.tenant_id = $1
		  AND p.active = TRUE
		  AND ($2::text IS NULL OR pc.name = $2)
		  AND ($3::text IS NULL OR p.name ILIKE '%' || $3 || '%' OR p.description ILIKE '%' || $3 || '%')`,
		tenantID, filter.Category, filter.Q,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count public products: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage

	// Data query — selects ONLY public-safe columns.
	// COALESCE(fiscal_price, internal_price) is exposed as "price" to the public.
	// fiscal_price is also returned as the optional second price for fiscal products.
	// internal_price is intentionally excluded from SELECT list (STORE-CATALOG-7).
	rows, err := s.pool.Query(ctx, `
		SELECT p.id,
		       p.tenant_id,
		       p.name,
		       COALESCE(p.description, '') AS description,
		       COALESCE(p.fiscal_price, p.internal_price, 0) AS price,
		       p.fiscal_price,
		       pc.name AS category,
		       COALESCE(ps.quantity_on_hand::integer, 0) AS stock_qty,
		       p.is_fiscal,
		       p.image_url
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		LEFT JOIN product_stock ps ON ps.product_id = p.id
		WHERE p.tenant_id = $1
		  AND p.active = TRUE
		  AND ($2::text IS NULL OR pc.name = $2)
		  AND ($3::text IS NULL OR p.name ILIKE '%' || $3 || '%' OR p.description ILIKE '%' || $3 || '%')
		ORDER BY p.name ASC
		LIMIT $4 OFFSET $5`,
		tenantID, filter.Category, filter.Q, filter.PerPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list public products: %w", err)
	}
	defer rows.Close()

	var products []shop.ShopProduct
	for rows.Next() {
		p, err := scanShopProduct(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan shop product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list public products rows: %w", err)
	}
	if products == nil {
		products = []shop.ShopProduct{}
	}
	return products, total, nil
}

// GetPublicProduct returns a single active product by ID for the given tenant.
// Returns shop.ErrProductNotFound if the product does not exist, belongs to a
// different tenant, or is inactive.
func (s *ShopService) GetPublicProduct(
	ctx context.Context,
	tenantID, productID uuid.UUID,
) (*shop.ShopProduct, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT p.id,
		       p.tenant_id,
		       p.name,
		       COALESCE(p.description, '') AS description,
		       COALESCE(p.fiscal_price, p.internal_price, 0) AS price,
		       p.fiscal_price,
		       pc.name AS category,
		       COALESCE(ps.quantity_on_hand::integer, 0) AS stock_qty,
		       p.is_fiscal,
		       p.image_url
		FROM products p
		LEFT JOIN product_categories pc ON pc.id = p.category_id
		LEFT JOIN product_stock ps ON ps.product_id = p.id
		WHERE p.id = $1
		  AND p.tenant_id = $2
		  AND p.active = TRUE`,
		productID, tenantID,
	)

	p, err := scanShopProduct(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shop.ErrProductNotFound
		}
		return nil, fmt.Errorf("get public product: %w", err)
	}
	return &p, nil
}

// scanShopProduct scans a single row into a ShopProduct.
// Works with both pgx.Row and pgx.Rows.
func scanShopProduct(row interface {
	Scan(dest ...any) error
}) (shop.ShopProduct, error) {
	var p shop.ShopProduct
	var category *string
	var fiscalPrice *float64
	var imageURL *string

	err := row.Scan(
		&p.ID,
		&p.TenantID,
		&p.Name,
		&p.Description,
		&p.Price,
		&fiscalPrice,
		&category,
		&p.StockQty,
		&p.IsFiscal,
		&imageURL,
	)
	if err != nil {
		return shop.ShopProduct{}, err
	}
	p.FiscalPrice = fiscalPrice
	p.Category = category
	p.ImageURL = imageURL
	return p, nil
}
