package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/danzt/daas/api/internal/domain/shop"
)

// ShopOrderService handles the public storefront order lifecycle.
// Stock is decremented atomically at checkout via Serializable transactions.
type ShopOrderService struct {
	pool     *pgxpool.Pool
	notifier *ShopOrderNotifier // optional — when nil, lifecycle emails are skipped
}

// NewShopOrderService creates a ShopOrderService backed by the given pool.
func NewShopOrderService(pool *pgxpool.Pool) *ShopOrderService {
	return &ShopOrderService{pool: pool}
}

// SetNotifier wires the lifecycle email notifier. Safe to call once at startup.
// When nil (or never called), all order state transitions skip the email send.
func (s *ShopOrderService) SetNotifier(n *ShopOrderNotifier) {
	s.notifier = n
}

// ═══ Public checkout ═════════════════════════════════════════════════════════

// Checkout creates a shop order from the storefront, decrements stock atomically,
// and returns the order with its one-time access token.
//
// The transaction uses Serializable isolation to protect against concurrent checkouts
// of the last available unit.
func (s *ShopOrderService) Checkout(ctx context.Context, tenantID uuid.UUID, req shop.CreateOrderRequest) (*shop.ShopOrder, error) {
	// 1. Validate request invariants.
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 2. Open Serializable transaction for atomic stock check + decrement.
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, fmt.Errorf("begin checkout tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// 3. For each line: lock product row, check active + stock, snapshot price/name.
	lines := make([]shop.OrderLine, 0, len(req.Lines))
	var subtotal float64

	for i, l := range req.Lines {
		var name string
		var internalPrice, fiscalPrice *float64
		var stockOnHand float64
		var isActive, isFiscal bool

		err := tx.QueryRow(ctx, `
			SELECT p.name,
			       p.internal_price,
			       p.fiscal_price,
			       COALESCE(ps.quantity_on_hand, 0),
			       p.active,
			       p.is_fiscal
			FROM products p
			LEFT JOIN product_stock ps ON ps.product_id = p.id
			WHERE p.id = $1 AND p.tenant_id = $2
			FOR UPDATE OF p`,
			l.ProductID, tenantID,
		).Scan(&name, &internalPrice, &fiscalPrice, &stockOnHand, &isActive, &isFiscal)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, shop.ErrProductNotPurchasable
			}
			return nil, fmt.Errorf("lock product %s: %w", l.ProductID, err)
		}

		if !isActive {
			return nil, shop.ErrProductNotPurchasable
		}
		if int(stockOnHand) < l.Quantity {
			return nil, shop.ErrInsufficientStock
		}

		// Resolve price: fiscal products use fiscal_price; non-fiscal use internal_price.
		var unitPrice float64
		if isFiscal && fiscalPrice != nil {
			unitPrice = *fiscalPrice
		} else if internalPrice != nil {
			unitPrice = *internalPrice
		}

		lineSubtotal := unitPrice * float64(l.Quantity)
		subtotal += lineSubtotal

		lines = append(lines, shop.OrderLine{
			ProductID: l.ProductID,
			Name:      name,
			UnitPrice: unitPrice,
			Quantity:  l.Quantity,
			Subtotal:  lineSubtotal,
			IsFiscal:  isFiscal,
			SortOrder: i,
		})

		// Decrement stock.
		if _, err := tx.Exec(ctx,
			`UPDATE product_stock
			 SET quantity_on_hand = quantity_on_hand - $1, last_updated_at = NOW()
			 WHERE product_id = $2`,
			l.Quantity, l.ProductID,
		); err != nil {
			return nil, fmt.Errorf("decrement stock for product %s: %w", l.ProductID, err)
		}
	}

	// 4. Compute shipping: flat $5 if subtotal < $100, else free.
	shippingCost := 0.0
	if subtotal < 100 {
		shippingCost = 5.0
	}
	total := subtotal + shippingCost

	// 5. Generate access token (crypto/rand hex 32 bytes → 64 hex chars).
	accessToken := generateAccessToken()

	// 6. Insert order.
	orderID := uuid.New()
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		INSERT INTO shop_orders (
			id, tenant_id,
			customer_name, customer_email, customer_phone,
			shipping_address, shipping_city, shipping_notes,
			subtotal, shipping_cost, total,
			status, notes, access_token,
			created_at, updated_at
		) VALUES (
			$1, $2,
			$3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			'pending', $12, $13,
			$14, $14
		)`,
		orderID, tenantID,
		req.CustomerName, req.CustomerEmail, req.CustomerPhone,
		req.ShippingAddress, req.ShippingCity, req.ShippingNotes,
		subtotal, shippingCost, total,
		req.Notes, accessToken,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert shop order: %w", err)
	}

	// 7. Insert order lines.
	for i := range lines {
		lines[i].ID = uuid.New()
		lines[i].OrderID = orderID
		_, err = tx.Exec(ctx, `
			INSERT INTO shop_order_lines (id, order_id, product_id, name, unit_price, quantity, subtotal, is_fiscal, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			lines[i].ID, orderID, lines[i].ProductID,
			lines[i].Name, lines[i].UnitPrice, lines[i].Quantity,
			lines[i].Subtotal, lines[i].IsFiscal, lines[i].SortOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("insert shop order line: %w", err)
		}
	}

	// 8. Commit.
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit checkout: %w", err)
	}

	return &shop.ShopOrder{
		ID:              orderID,
		TenantID:        tenantID,
		CustomerName:    req.CustomerName,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		ShippingAddress: req.ShippingAddress,
		ShippingCity:    req.ShippingCity,
		ShippingNotes:   req.ShippingNotes,
		Subtotal:        subtotal,
		ShippingCost:    shippingCost,
		Total:           total,
		Status:          shop.OrderStatusPending,
		Notes:           req.Notes,
		AccessToken:     accessToken,
		CreatedAt:       now,
		UpdatedAt:       now,
		Lines:           lines,
	}, nil
}

// ═══ Public order tracking ════════════════════════════════════════════════════

// GetPublicOrder returns an order for the customer, validating the access token.
// Returns ErrShopOrderNotFound if the order does not exist or the token does not match.
// The returned ShopOrder has AccessToken cleared (not exposed on tracking view).
func (s *ShopOrderService) GetPublicOrder(ctx context.Context, tenantID, orderID uuid.UUID, accessToken string) (*shop.ShopOrder, error) {
	order, err := s.getOrderByIDWithToken(ctx, tenantID, orderID, accessToken)
	if err != nil {
		return nil, err
	}
	// Never expose the token after creation.
	order.AccessToken = ""
	return order, nil
}

// ═══ Tenant admin ═════════════════════════════════════════════════════════════

// GetAdminOrder returns an order for the tenant admin (no token required).
func (s *ShopOrderService) GetAdminOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
	return s.getOrderByID(ctx, tenantID, orderID)
}

// ListTenantOrders returns a paginated list of orders for the tenant admin.
func (s *ShopOrderService) ListTenantOrders(ctx context.Context, tenantID uuid.UUID, status *shop.OrderStatus, limit, offset int) ([]*shop.ShopOrder, int, error) {
	// Count.
	var total int
	countQ := `SELECT COUNT(*) FROM shop_orders WHERE tenant_id = $1`
	countArgs := []any{tenantID}
	if status != nil {
		countQ += ` AND status = $2`
		countArgs = append(countArgs, *status)
	}
	if err := s.pool.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count shop orders: %w", err)
	}

	// Data.
	dataQ := `
		SELECT id, tenant_id,
		       customer_name, customer_email, COALESCE(customer_phone, ''),
		       shipping_address, COALESCE(shipping_city, ''), COALESCE(shipping_notes, ''),
		       subtotal, shipping_cost, total,
		       status, COALESCE(notes, ''),
		       paid_at, fulfilled_at, delivered_at, cancelled_at,
		       created_at, updated_at
		FROM shop_orders
		WHERE tenant_id = $1`
	dataArgs := []any{tenantID}
	if status != nil {
		dataQ += ` AND status = $2`
		dataArgs = append(dataArgs, *status)
		dataQ += ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`
		dataArgs = append(dataArgs, limit, offset)
	} else {
		dataQ += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		dataArgs = append(dataArgs, limit, offset)
	}

	rows, err := s.pool.Query(ctx, dataQ, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list shop orders: %w", err)
	}
	defer rows.Close()

	var orders []*shop.ShopOrder
	for rows.Next() {
		o, err := scanShopOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan shop order: %w", err)
		}
		// Don't expose access token in list view.
		o.AccessToken = ""
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("shop orders rows: %w", err)
	}
	if orders == nil {
		orders = []*shop.ShopOrder{}
	}
	return orders, total, nil
}

// MarkPaid transitions pending → paid.
func (s *ShopOrderService) MarkPaid(ctx context.Context, tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
	o, err := s.transition(ctx, tenantID, orderID, shop.OrderStatusPaid, "paid_at")
	if err == nil {
		s.notifier.Notify(ctx, tenantID, o)
	}
	return o, err
}

// MarkFulfilled transitions paid → fulfilled.
func (s *ShopOrderService) MarkFulfilled(ctx context.Context, tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
	o, err := s.transition(ctx, tenantID, orderID, shop.OrderStatusFulfilled, "fulfilled_at")
	if err == nil {
		s.notifier.Notify(ctx, tenantID, o)
	}
	return o, err
}

// MarkDelivered transitions fulfilled → delivered.
func (s *ShopOrderService) MarkDelivered(ctx context.Context, tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
	o, err := s.transition(ctx, tenantID, orderID, shop.OrderStatusDelivered, "delivered_at")
	if err == nil {
		s.notifier.Notify(ctx, tenantID, o)
	}
	return o, err
}

// Cancel transitions pending|paid → cancelled (admin version, no token required).
func (s *ShopOrderService) Cancel(ctx context.Context, tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
	o, err := s.transition(ctx, tenantID, orderID, shop.OrderStatusCancelled, "cancelled_at")
	if err == nil {
		s.notifier.Notify(ctx, tenantID, o)
	}
	return o, err
}

// CustomerCancel transitions pending → cancelled using the access token.
func (s *ShopOrderService) CustomerCancel(ctx context.Context, tenantID, orderID uuid.UUID, accessToken string) (*shop.ShopOrder, error) {
	order, err := s.getOrderByIDWithToken(ctx, tenantID, orderID, accessToken)
	if err != nil {
		return nil, err
	}
	if err := order.CanTransitionTo(shop.OrderStatusCancelled); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_, err = s.pool.Exec(ctx, `
		UPDATE shop_orders
		SET status = 'cancelled', cancelled_at = $1, updated_at = $1
		WHERE id = $2 AND tenant_id = $3`,
		now, orderID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("cancel shop order: %w", err)
	}
	o, err := s.getOrderByID(ctx, tenantID, orderID)
	if err == nil {
		s.notifier.Notify(ctx, tenantID, o)
	}
	return o, err
}

// ─── Private helpers ──────────────────────────────────────────────────────────

// transition runs a status transition for admin operations (no token check).
func (s *ShopOrderService) transition(ctx context.Context, tenantID, orderID uuid.UUID, next shop.OrderStatus, tsField string) (*shop.ShopOrder, error) {
	order, err := s.getOrderByID(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	if err := order.CanTransitionTo(next); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	q := fmt.Sprintf(`
		UPDATE shop_orders
		SET status = $1, %s = $2, updated_at = $2
		WHERE id = $3 AND tenant_id = $4`, tsField)
	if _, err := s.pool.Exec(ctx, q, next, now, orderID, tenantID); err != nil {
		return nil, fmt.Errorf("update shop order status: %w", err)
	}
	return s.getOrderByID(ctx, tenantID, orderID)
}

// getOrderByID fetches a full order (lines included) by tenant + order ID.
func (s *ShopOrderService) getOrderByID(ctx context.Context, tenantID, orderID uuid.UUID) (*shop.ShopOrder, error) {
	o, err := scanShopOrder(s.pool.QueryRow(ctx, `
		SELECT id, tenant_id,
		       customer_name, customer_email, COALESCE(customer_phone, ''),
		       shipping_address, COALESCE(shipping_city, ''), COALESCE(shipping_notes, ''),
		       subtotal, shipping_cost, total,
		       status, COALESCE(notes, ''),
		       paid_at, fulfilled_at, delivered_at, cancelled_at,
		       created_at, updated_at
		FROM shop_orders
		WHERE id = $1 AND tenant_id = $2`,
		orderID, tenantID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shop.ErrShopOrderNotFound
		}
		return nil, fmt.Errorf("get shop order: %w", err)
	}
	lines, err := s.loadLines(ctx, orderID)
	if err != nil {
		return nil, err
	}
	o.Lines = lines
	return o, nil
}

// getOrderByIDWithToken fetches an order and validates the access token.
func (s *ShopOrderService) getOrderByIDWithToken(ctx context.Context, tenantID, orderID uuid.UUID, accessToken string) (*shop.ShopOrder, error) {
	o, err := scanShopOrder(s.pool.QueryRow(ctx, `
		SELECT id, tenant_id,
		       customer_name, customer_email, COALESCE(customer_phone, ''),
		       shipping_address, COALESCE(shipping_city, ''), COALESCE(shipping_notes, ''),
		       subtotal, shipping_cost, total,
		       status, COALESCE(notes, ''),
		       paid_at, fulfilled_at, delivered_at, cancelled_at,
		       created_at, updated_at
		FROM shop_orders
		WHERE id = $1 AND tenant_id = $2 AND access_token = $3`,
		orderID, tenantID, accessToken,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Intentionally opaque: whether the order doesn't exist or the token is wrong,
			// we return the same error to prevent token enumeration.
			return nil, shop.ErrShopOrderNotFound
		}
		return nil, fmt.Errorf("get shop order (public): %w", err)
	}
	lines, err := s.loadLines(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Lines = lines
	return o, nil
}

// loadLines returns all lines for a given order sorted by sort_order.
func (s *ShopOrderService) loadLines(ctx context.Context, orderID uuid.UUID) ([]shop.OrderLine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, order_id, product_id, name, unit_price, quantity, subtotal, is_fiscal, sort_order
		FROM shop_order_lines
		WHERE order_id = $1
		ORDER BY sort_order`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("load shop order lines: %w", err)
	}
	defer rows.Close()

	var lines []shop.OrderLine
	for rows.Next() {
		var l shop.OrderLine
		if err := rows.Scan(
			&l.ID, &l.OrderID, &l.ProductID,
			&l.Name, &l.UnitPrice, &l.Quantity,
			&l.Subtotal, &l.IsFiscal, &l.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("scan order line: %w", err)
		}
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

type shopOrderScanner interface {
	Scan(dest ...any) error
}

func scanShopOrder(row shopOrderScanner) (*shop.ShopOrder, error) {
	var o shop.ShopOrder
	if err := row.Scan(
		&o.ID, &o.TenantID,
		&o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
		&o.ShippingAddress, &o.ShippingCity, &o.ShippingNotes,
		&o.Subtotal, &o.ShippingCost, &o.Total,
		&o.Status, &o.Notes,
		&o.PaidAt, &o.FulfilledAt, &o.DeliveredAt, &o.CancelledAt,
		&o.CreatedAt, &o.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &o, nil
}

// generateAccessToken returns a 64-char hex string from crypto/rand.
func generateAccessToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely; fall back to timestamp-based to avoid panic in handlers.
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
