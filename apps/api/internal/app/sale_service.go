package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danzt/daas/api/internal/domain/invoice"
	"github.com/danzt/daas/api/internal/domain/sale"
)

// SaleService handles the sales order lifecycle.
type SaleService struct {
	pool *pgxpool.Pool
}

func NewSaleService(pool *pgxpool.Pool) *SaleService {
	return &SaleService{pool: pool}
}

// ═══ CRUD ════════════════════════════════════════════════════════════════════

func (s *SaleService) CreateOrder(ctx context.Context, tenantID uuid.UUID, createdByUID string, req sale.CreateOrderRequest) (*sale.SaleOrder, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	createdBy, err := s.resolveUserID(ctx, tenantID, createdByUID)
	if err != nil {
		return nil, err
	}

	// Validate products and compute total.
	var total float64
	for i, l := range req.Lines {
		var name string
		err := s.pool.QueryRow(ctx,
			`SELECT name FROM products WHERE id=$1 AND tenant_id=$2 AND active=true`,
			l.ProductID, tenantID,
		).Scan(&name)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("product %s not found or inactive", l.ProductID)
			}
			return nil, fmt.Errorf("lookup product: %w", err)
		}
		if req.Lines[i].Description == "" {
			req.Lines[i].Description = name
		}
		total += l.Quantity * l.UnitPrice
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	idType := req.CustomerIDType
	if idType == "" {
		idType = "anonymous"
	}

	order := &sale.SaleOrder{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Status:           sale.OrderStatusDraft,
		CustomerName:     req.CustomerName,
		CustomerIDType:   idType,
		CustomerIDNumber: req.CustomerIDNumber,
		Notes:            req.Notes,
		Total:            total,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO sales_orders
		    (id, tenant_id, status, customer_name, customer_id_type, customer_id_number, notes, total, created_by)
		 VALUES ($1,$2,'draft',$3,$4,$5,$6,$7,$8)`,
		order.ID, order.TenantID, order.CustomerName, order.CustomerIDType,
		order.CustomerIDNumber, order.Notes, order.Total, order.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	for i, l := range req.Lines {
		lineID := uuid.New()
		sub := l.Quantity * l.UnitPrice
		_, err = tx.Exec(ctx,
			`INSERT INTO sales_order_lines
			    (id, order_id, product_id, description, quantity, unit_price, subtotal, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			lineID, order.ID, l.ProductID, req.Lines[i].Description, l.Quantity, l.UnitPrice, sub, i,
		)
		if err != nil {
			return nil, fmt.Errorf("insert line: %w", err)
		}
		order.Lines = append(order.Lines, sale.OrderLine{
			ID: lineID, OrderID: order.ID, ProductID: l.ProductID,
			Description: req.Lines[i].Description, Quantity: l.Quantity,
			UnitPrice: l.UnitPrice, Subtotal: sub, SortOrder: i,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return order, nil
}

func (s *SaleService) ListOrders(ctx context.Context, tenantID uuid.UUID, status *sale.OrderStatus) ([]*sale.SaleOrder, error) {
	query := `SELECT id, tenant_id, status, COALESCE(customer_name,''),
	                 customer_id_type, COALESCE(customer_id_number,''),
	                 COALESCE(notes,''), total, confirmed_at, invoiced_at, invoice_id,
	                 created_by, created_at, updated_at
	          FROM sales_orders WHERE tenant_id=$1`
	args := []any{tenantID}
	if status != nil {
		query += " AND status=$2"
		args = append(args, *status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()
	var result []*sale.SaleOrder
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

func (s *SaleService) GetOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*sale.SaleOrder, error) {
	order, err := scanOrder(s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, status, COALESCE(customer_name,''),
		        customer_id_type, COALESCE(customer_id_number,''),
		        COALESCE(notes,''), total, confirmed_at, invoiced_at, invoice_id,
		        created_by, created_at, updated_at
		 FROM sales_orders WHERE id=$1 AND tenant_id=$2`,
		orderID, tenantID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sale.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	lines, err := s.loadLines(ctx, orderID)
	if err != nil {
		return nil, err
	}
	order.Lines = lines
	return order, nil
}

// ═══ Lifecycle ═══════════════════════════════════════════════════════════════

// ConfirmOrder transitions draft → confirmed and deducts stock.
func (s *SaleService) ConfirmOrder(ctx context.Context, tenantID, orderID uuid.UUID, supabaseUID string) (*sale.SaleOrder, error) {
	userID, err := s.resolveUserID(ctx, tenantID, supabaseUID)
	if err != nil {
		return nil, err
	}
	order, err := s.GetOrder(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	switch order.Status {
	case sale.OrderStatusConfirmed:
		return nil, sale.ErrOrderAlreadyConfirmed
	case sale.OrderStatusInvoiced:
		return nil, sale.ErrOrderAlreadyInvoiced
	case sale.OrderStatusCancelled:
		return nil, sale.ErrOrderAlreadyCancelled
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	now := time.Now().UTC()

	// Deduct stock per line.
	for _, line := range order.Lines {
		var onHand float64
		err := tx.QueryRow(ctx,
			`SELECT quantity_on_hand FROM product_stock WHERE product_id=$1 AND tenant_id=$2 FOR UPDATE`,
			line.ProductID, tenantID,
		).Scan(&onHand)
		if err != nil {
			// If no stock row exists, treat as zero stock.
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("lock stock: %w", err)
			}
		}
		if onHand < line.Quantity {
			return nil, sale.ErrInsufficientStock
		}

		_, err = tx.Exec(ctx,
			`UPDATE product_stock
			 SET quantity_on_hand = quantity_on_hand - $1, last_updated_at = NOW()
			 WHERE product_id = $2 AND tenant_id = $3`,
			line.Quantity, line.ProductID, tenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("deduct stock: %w", err)
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO inventory_movements
			    (id, tenant_id, product_id, type, quantity, reference_type, reference_id, notes, created_by)
			 VALUES ($1,$2,$3,'exit',$4,'sale',$5,'Orden de venta confirmada',$6)`,
			uuid.New(), tenantID, line.ProductID, line.Quantity, orderID, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("insert movement: %w", err)
		}
	}

	_, err = tx.Exec(ctx,
		`UPDATE sales_orders SET status='confirmed', confirmed_at=$1, updated_at=NOW() WHERE id=$2`,
		now, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("update order status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.GetOrder(ctx, tenantID, orderID)
}

// InvoiceOrder transitions confirmed → invoiced and creates an internal invoice.
// Stock is already deducted at Confirm time, so no inventory movement is created here.
func (s *SaleService) InvoiceOrder(ctx context.Context, tenantID, orderID uuid.UUID, supabaseUID string) (*sale.SaleOrder, error) {
	createdBy, err := s.resolveUserID(ctx, tenantID, supabaseUID)
	if err != nil {
		return nil, err
	}
	order, err := s.GetOrder(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != sale.OrderStatusConfirmed {
		switch order.Status {
		case sale.OrderStatusInvoiced:
			return nil, sale.ErrOrderAlreadyInvoiced
		case sale.OrderStatusCancelled:
			return nil, sale.ErrOrderAlreadyCancelled
		default:
			return nil, sale.ErrOrderNotConfirmed
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Generate correlative atomically.
	year := time.Now().Year()
	_, err = tx.Exec(ctx,
		`INSERT INTO invoice_correlative_seq (tenant_id, year, last_seq)
		 VALUES ($1, $2, 0)
		 ON CONFLICT (tenant_id, year) DO NOTHING`,
		tenantID, year,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert correlative: %w", err)
	}
	var seq int
	err = tx.QueryRow(ctx,
		`UPDATE invoice_correlative_seq
		 SET last_seq = last_seq + 1
		 WHERE tenant_id=$1 AND year=$2
		 RETURNING last_seq`,
		tenantID, year,
	).Scan(&seq)
	if err != nil {
		return nil, fmt.Errorf("increment correlative: %w", err)
	}
	correlative := invoice.FormatCorrelative(year, seq)
	now := time.Now().UTC()

	idType := order.CustomerIDType
	if idType == "" {
		idType = "anonymous"
	}

	// Create invoice directly as issued (stock already deducted at confirm).
	invoiceID := uuid.New()
	_, err = tx.Exec(ctx,
		`INSERT INTO internal_invoices
		    (id, tenant_id, correlative, customer_name, customer_id_type, customer_id_number,
		     status, subtotal, total, notes, issued_at, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,'issued',$7,$8,$9,$10,$11)`,
		invoiceID, tenantID, correlative,
		order.CustomerName, idType, order.CustomerIDNumber,
		order.Total, order.Total,
		fmt.Sprintf("Orden de venta %s", order.ID),
		now, createdBy,
	)
	if err != nil {
		return nil, fmt.Errorf("insert invoice: %w", err)
	}

	for i, l := range order.Lines {
		_, err = tx.Exec(ctx,
			`INSERT INTO internal_invoice_lines
			    (id, invoice_id, product_id, description, quantity, unit_price, subtotal, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			uuid.New(), invoiceID, l.ProductID, l.Description, l.Quantity, l.UnitPrice, l.Subtotal, i,
		)
		if err != nil {
			return nil, fmt.Errorf("insert invoice line: %w", err)
		}
	}

	_, err = tx.Exec(ctx,
		`UPDATE sales_orders SET status='invoiced', invoice_id=$1, invoiced_at=$2, updated_at=NOW() WHERE id=$3`,
		invoiceID, now, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.GetOrder(ctx, tenantID, orderID)
}

// CancelOrder transitions any non-final state to cancelled.
// If the order was confirmed, stock is restored.
func (s *SaleService) CancelOrder(ctx context.Context, tenantID, orderID uuid.UUID, supabaseUID string) (*sale.SaleOrder, error) {
	userID, err := s.resolveUserID(ctx, tenantID, supabaseUID)
	if err != nil {
		return nil, err
	}
	order, err := s.GetOrder(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	switch order.Status {
	case sale.OrderStatusCancelled:
		return nil, sale.ErrOrderAlreadyCancelled
	case sale.OrderStatusInvoiced:
		return nil, sale.ErrOrderAlreadyInvoiced
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Restore stock if order was already confirmed (stock was deducted).
	if order.Status == sale.OrderStatusConfirmed {
		for _, line := range order.Lines {
			_, err = tx.Exec(ctx,
				`UPDATE product_stock
				 SET quantity_on_hand = quantity_on_hand + $1, last_updated_at = NOW()
				 WHERE product_id = $2 AND tenant_id = $3`,
				line.Quantity, line.ProductID, tenantID,
			)
			if err != nil {
				return nil, fmt.Errorf("restore stock: %w", err)
			}
			_, err = tx.Exec(ctx,
				`INSERT INTO inventory_movements
				    (id, tenant_id, product_id, type, quantity, reference_type, reference_id, notes, created_by)
				 VALUES ($1,$2,$3,'adjustment',$4,'sale',$5,'Reversa por cancelación de orden',$6)`,
				uuid.New(), tenantID, line.ProductID, line.Quantity, orderID, userID,
			)
			if err != nil {
				return nil, fmt.Errorf("insert reversal movement: %w", err)
			}
		}
	}

	_, err = tx.Exec(ctx,
		`UPDATE sales_orders SET status='cancelled', updated_at=NOW() WHERE id=$1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("cancel order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.GetOrder(ctx, tenantID, orderID)
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (s *SaleService) resolveUserID(ctx context.Context, tenantID uuid.UUID, supabaseUID string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM tenant_users WHERE tenant_id=$1 AND supabase_uid=$2`,
		tenantID, supabaseUID,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve user: %w", err)
	}
	return id, nil
}

func (s *SaleService) loadLines(ctx context.Context, orderID uuid.UUID) ([]sale.OrderLine, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, order_id, product_id, description, quantity, unit_price, subtotal, sort_order
		 FROM sales_order_lines WHERE order_id=$1 ORDER BY sort_order`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("load lines: %w", err)
	}
	defer rows.Close()
	var lines []sale.OrderLine
	for rows.Next() {
		var l sale.OrderLine
		if err := rows.Scan(&l.ID, &l.OrderID, &l.ProductID, &l.Description,
			&l.Quantity, &l.UnitPrice, &l.Subtotal, &l.SortOrder); err != nil {
			return nil, fmt.Errorf("scan line: %w", err)
		}
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

type orderScanner interface {
	Scan(dest ...any) error
}

func scanOrder(row orderScanner) (*sale.SaleOrder, error) {
	var o sale.SaleOrder
	if err := row.Scan(
		&o.ID, &o.TenantID, &o.Status,
		&o.CustomerName, &o.CustomerIDType, &o.CustomerIDNumber,
		&o.Notes, &o.Total,
		&o.ConfirmedAt, &o.InvoicedAt, &o.InvoiceID,
		&o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &o, nil
}
