package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danzt/daas/api/internal/domain/inventory"
)

// MovementFilters holds optional filters for listing inventory movements.
type MovementFilters struct {
	ProductID *uuid.UUID
	Type      string // "entry" | "exit" | "adjustment" | ""
	From      *time.Time
	To        *time.Time
	Page      int
	Limit     int
}

// InventoryService handles all inventory control use cases.
type InventoryService struct {
	pool *pgxpool.Pool
}

// NewInventoryService creates an InventoryService backed by the given pool.
func NewInventoryService(pool *pgxpool.Pool) *InventoryService {
	return &InventoryService{pool: pool}
}

// ListStock returns the current stock for all products of a tenant.
func (s *InventoryService) ListStock(ctx context.Context, tenantID uuid.UUID) ([]*inventory.Stock, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ps.product_id, ps.tenant_id, ps.quantity_on_hand, ps.last_updated_at
		FROM product_stock ps
		WHERE ps.tenant_id = $1
		ORDER BY ps.last_updated_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list stock: %w", err)
	}
	defer rows.Close()

	var result []*inventory.Stock
	for rows.Next() {
		st, err := scanStock(rows)
		if err != nil {
			return nil, fmt.Errorf("scan stock: %w", err)
		}
		result = append(result, st)
	}
	return result, rows.Err()
}

// GetProductStock returns the current stock for a single product.
func (s *InventoryService) GetProductStock(ctx context.Context, tenantID, productID uuid.UUID) (*inventory.Stock, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT product_id, tenant_id, quantity_on_hand, last_updated_at
		FROM product_stock
		WHERE product_id = $1 AND tenant_id = $2`,
		productID, tenantID,
	)
	st, err := scanStock(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, inventory.ErrStockNotFound
		}
		return nil, fmt.Errorf("get product stock: %w", err)
	}
	return st, nil
}

// ListMovements returns a paginated list of inventory movements for the tenant.
func (s *InventoryService) ListMovements(ctx context.Context, tenantID uuid.UUID, f MovementFilters) ([]*inventory.Movement, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, product_id, type, quantity, unit_cost,
		       reference_type, reference_id, notes, created_by, created_at
		FROM inventory_movements
		WHERE tenant_id = $1
		  AND ($2::uuid IS NULL OR product_id = $2)
		  AND ($3::varchar IS NULL OR type = $3)
		  AND ($4::timestamptz IS NULL OR created_at >= $4)
		  AND ($5::timestamptz IS NULL OR created_at <= $5)
		ORDER BY created_at DESC
		LIMIT $6 OFFSET $7`,
		tenantID, f.ProductID, nullableString(f.Type), f.From, f.To, f.Limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list movements: %w", err)
	}
	defer rows.Close()

	var result []*inventory.Movement
	for rows.Next() {
		m, err := scanMovement(rows)
		if err != nil {
			return nil, fmt.Errorf("scan movement: %w", err)
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// GetMovement returns a single movement by ID.
func (s *InventoryService) GetMovement(ctx context.Context, tenantID, id uuid.UUID) (*inventory.Movement, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, product_id, type, quantity, unit_cost,
		       reference_type, reference_id, notes, created_by, created_at
		FROM inventory_movements
		WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	m, err := scanMovement(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, inventory.ErrMovementNotFound
		}
		return nil, fmt.Errorf("get movement: %w", err)
	}
	return m, nil
}

// CreateAdjustment records a manual stock adjustment. The delta may be positive
// (increase) or negative (decrease). Negative adjustments that would result in
// negative stock are rejected with ErrInsufficientStock.
// createdBySupabaseUID is the Supabase sub claim; it is resolved to a tenant_users.id.
func (s *InventoryService) CreateAdjustment(
	ctx context.Context,
	tenantID uuid.UUID,
	createdBySupabaseUID string,
	req inventory.AdjustmentRequest,
) (*inventory.Movement, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Resolve supabase_uid → tenant_users.id
	var createdByID uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM tenant_users WHERE tenant_id = $1 AND supabase_uid = $2`,
		tenantID, createdBySupabaseUID,
	).Scan(&createdByID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	// Transactional: lock stock row, validate, apply delta, record movement.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Lock the stock row for update.
	var current float64
	err = tx.QueryRow(ctx,
		`SELECT quantity_on_hand FROM product_stock
		 WHERE product_id = $1 AND tenant_id = $2
		 FOR UPDATE`,
		req.ProductID, tenantID,
	).Scan(&current)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, inventory.ErrStockNotFound
		}
		return nil, fmt.Errorf("lock stock: %w", err)
	}

	newQty := current + req.Delta
	if newQty < 0 {
		return nil, inventory.ErrInsufficientStock
	}

	// Update stock snapshot.
	_, err = tx.Exec(ctx,
		`UPDATE product_stock
		 SET quantity_on_hand = $1, last_updated_at = NOW()
		 WHERE product_id = $2 AND tenant_id = $3`,
		newQty, req.ProductID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("update stock: %w", err)
	}

	// Record the movement (quantity stored as absolute value).
	absQty := req.Delta
	if absQty < 0 {
		absQty = -absQty
	}

	var m inventory.Movement
	err = tx.QueryRow(ctx, `
		INSERT INTO inventory_movements
		  (tenant_id, product_id, type, quantity, reference_type, notes, created_by)
		VALUES ($1, $2, 'adjustment', $3, 'manual_adjustment', $4, $5)
		RETURNING id, tenant_id, product_id, type, quantity, unit_cost,
		          reference_type, reference_id, notes, created_by, created_at`,
		tenantID, req.ProductID, absQty, req.Notes, createdByID,
	).Scan(
		&m.ID, &m.TenantID, &m.ProductID, &m.Type, &m.Quantity, &m.UnitCost,
		&m.ReferenceType, &m.ReferenceID, &m.Notes, &m.CreatedBy, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &m, nil
}

// RegisterEntry records a stock entry (e.g. from purchase order receipt).
// This is called internally by the purchase order service in Sprint 7.
func (s *InventoryService) RegisterEntry(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, productID, createdByID uuid.UUID,
	quantity float64,
	unitCost *float64,
	refType inventory.ReferenceType,
	refID *uuid.UUID,
	notes string,
) (*inventory.Movement, error) {
	// Update stock (no lock needed — caller holds the transaction).
	_, err := tx.Exec(ctx,
		`UPDATE product_stock
		 SET quantity_on_hand = quantity_on_hand + $1, last_updated_at = NOW()
		 WHERE product_id = $2 AND tenant_id = $3`,
		quantity, productID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("update stock for entry: %w", err)
	}

	var m inventory.Movement
	err = tx.QueryRow(ctx, `
		INSERT INTO inventory_movements
		  (tenant_id, product_id, type, quantity, unit_cost, reference_type, reference_id, notes, created_by)
		VALUES ($1, $2, 'entry', $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, product_id, type, quantity, unit_cost,
		          reference_type, reference_id, notes, created_by, created_at`,
		tenantID, productID, quantity, unitCost, string(refType), refID, notes, createdByID,
	).Scan(
		&m.ID, &m.TenantID, &m.ProductID, &m.Type, &m.Quantity, &m.UnitCost,
		&m.ReferenceType, &m.ReferenceID, &m.Notes, &m.CreatedBy, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert entry movement: %w", err)
	}
	return &m, nil
}

// RegisterExit records a stock exit (e.g. from a sale). Returns
// ErrInsufficientStock if the exit would drive quantity below zero.
// Caller must hold an open transaction and lock the stock row.
func (s *InventoryService) RegisterExit(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, productID, createdByID uuid.UUID,
	quantity float64,
	refType inventory.ReferenceType,
	refID *uuid.UUID,
	notes string,
) (*inventory.Movement, error) {
	var current float64
	err := tx.QueryRow(ctx,
		`SELECT quantity_on_hand FROM product_stock
		 WHERE product_id = $1 AND tenant_id = $2 FOR UPDATE`,
		productID, tenantID,
	).Scan(&current)
	if err != nil {
		return nil, fmt.Errorf("lock stock for exit: %w", err)
	}
	if current-quantity < 0 {
		return nil, inventory.ErrInsufficientStock
	}

	_, err = tx.Exec(ctx,
		`UPDATE product_stock
		 SET quantity_on_hand = quantity_on_hand - $1, last_updated_at = NOW()
		 WHERE product_id = $2 AND tenant_id = $3`,
		quantity, productID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("update stock for exit: %w", err)
	}

	var m inventory.Movement
	err = tx.QueryRow(ctx, `
		INSERT INTO inventory_movements
		  (tenant_id, product_id, type, quantity, reference_type, reference_id, notes, created_by)
		VALUES ($1, $2, 'exit', $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, product_id, type, quantity, unit_cost,
		          reference_type, reference_id, notes, created_by, created_at`,
		tenantID, productID, quantity, string(refType), refID, notes, createdByID,
	).Scan(
		&m.ID, &m.TenantID, &m.ProductID, &m.Type, &m.Quantity, &m.UnitCost,
		&m.ReferenceType, &m.ReferenceID, &m.Notes, &m.CreatedBy, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert exit movement: %w", err)
	}
	return &m, nil
}

// ─── Scan helpers ────────────────────────────────────────────────────────────

func scanStock(row interface{ Scan(dest ...any) error }) (*inventory.Stock, error) {
	var st inventory.Stock
	err := row.Scan(&st.ProductID, &st.TenantID, &st.QuantityOnHand, &st.LastUpdatedAt)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func scanMovement(row interface{ Scan(dest ...any) error }) (*inventory.Movement, error) {
	var m inventory.Movement
	var movType, refType string
	err := row.Scan(
		&m.ID, &m.TenantID, &m.ProductID,
		&movType, &m.Quantity, &m.UnitCost,
		&refType, &m.ReferenceID,
		&m.Notes, &m.CreatedBy, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.Type = inventory.MovementType(movType)
	m.ReferenceType = inventory.ReferenceType(refType)
	return &m, nil
}
