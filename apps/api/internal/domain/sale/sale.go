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
	ID               uuid.UUID
	TenantID         uuid.UUID
	Status           OrderStatus
	CustomerName     string
	CustomerIDType   string
	CustomerIDNumber string
	Notes            string
	Total            float64
	ConfirmedAt      *time.Time
	InvoicedAt       *time.Time
	InvoiceID        *uuid.UUID
	CreatedBy        uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Lines            []OrderLine
}

// OrderLine is a single product line inside a sale order.
type OrderLine struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	Description string
	Quantity    float64
	UnitPrice   float64
	Subtotal    float64
	SortOrder   int
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
