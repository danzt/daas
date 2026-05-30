-- name: ListProducts :many
SELECT * FROM products
WHERE tenant_id = $1
  AND ($2::boolean IS NULL OR active = $2)
  AND ($3::boolean IS NULL OR is_fiscal = $3)
  AND ($4::uuid IS NULL OR category_id = $4)
  AND ($5::text IS NULL OR name ILIKE '%' || $5 || '%' OR sku ILIKE '%' || $5 || '%' OR barcode ILIKE '%' || $5 || '%')
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1 AND tenant_id = $2;

-- name: CreateProduct :one
INSERT INTO products (tenant_id, name, sku, barcode, description, category_id, is_fiscal, fiscal_price, internal_price, tax_rate)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET name = $3, sku = $4, barcode = $5, description = $6, category_id = $7,
    fiscal_price = $8, internal_price = $9, tax_rate = $10, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: SoftDeleteProduct :one
UPDATE products SET active = FALSE, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: ListCategories :many
SELECT * FROM product_categories WHERE tenant_id = $1 ORDER BY name;

-- name: CreateCategory :one
INSERT INTO product_categories (tenant_id, name, parent_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CheckBarcodeDuplicate :many
SELECT id, name FROM products
WHERE tenant_id = $1 AND barcode = $2 AND active = TRUE;
