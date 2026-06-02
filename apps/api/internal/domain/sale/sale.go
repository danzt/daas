// Package sale contains domain entities and business rules for the sales order
// lifecycle. Zero external imports.
package sale

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ─── Status ───────────────────────────────────────────────────────────────────

type OrderStatus string

const (
	OrderStatusDraft     OrderStatus = "draft"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusInvoiced  OrderStatus = "invoiced"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// ─── Aggregate ────────────────────────────────────────────────────────────────

// SaleOrder is the root aggregate for a customer sale.
type SaleOrder struct {
	ID               uuid.UUID   `json:"id"`
	TenantID         uuid.UUID   `json:"tenant_id"`
	Status           OrderStatus `json:"status"`
	CustomerName     string      `json:"customer_name"`
	CustomerIDType   string      `json:"customer_id_type"`
	CustomerIDNumber string      `json:"customer_id_number"`
	Notes            string      `json:"notes"`
	Total            float64     `json:"total"`
	ConfirmedAt      *time.Time  `json:"confirmed_at"`
	InvoicedAt       *time.Time  `json:"invoiced_at"`
	InvoiceID        *uuid.UUID  `json:"invoice_id"`
	CreatedBy        uuid.UUID   `json:"created_by"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	Lines            []OrderLine `json:"lines"`
}

// OrderLine is a single product line inside a sale order.
type OrderLine struct {
	ID          uuid.UUID `json:"id"`
	OrderID     uuid.UUID `json:"order_id"`
	ProductID   uuid.UUID `json:"product_id"`
	Description string    `json:"description"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Subtotal    float64   `json:"subtotal"`
	SortOrder   int       `json:"sort_order"`
}

// ─── Request types ────────────────────────────────────────────────────────────

type CreateOrderLineRequest struct {
	ProductID   uuid.UUID
	Description string
	Quantity    float64
	UnitPrice   float64
}

type CreateOrderRequest struct {
	CustomerName     string
	CustomerIDType   string
	CustomerIDNumber string
	Notes            string
	Lines            []CreateOrderLineRequest
}

func (r *CreateOrderRequest) Validate() error {
	if len(r.Lines) == 0 {
		return ErrEmptyOrder
	}
	for _, l := range r.Lines {
		if l.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive")
		}
		if l.UnitPrice <= 0 {
			return fmt.Errorf("unit_price must be positive")
		}
	}
	return nil
}

// ─── Domain errors ────────────────────────────────────────────────────────────

var (
	ErrOrderNotFound         = errors.New("sale order not found")
	ErrEmptyOrder            = errors.New("sale order must have at least one line")
	ErrOrderAlreadyConfirmed = errors.New("sale order is already confirmed")
	ErrOrderAlreadyInvoiced  = errors.New("sale order is already invoiced")
	ErrOrderAlreadyCancelled = errors.New("sale order is already cancelled")
	ErrOrderNotConfirmed     = errors.New("sale order must be confirmed before invoicing")
	ErrOrderNotDraft         = errors.New("sale order is not in draft status")
	ErrInsufficientStock     = errors.New("insufficient stock for one or more products")
)
