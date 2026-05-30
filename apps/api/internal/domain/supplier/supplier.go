// Package supplier contains the domain entities and business rules for
// supplier management and purchase orders. Zero external imports.
package supplier

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ─── Supplier ────────────────────────────────────────────────────────────────

// Supplier is an external vendor from which the tenant purchases products.
type Supplier struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	RIF         string
	ContactName string
	Email       string
	Phone       string
	Address     string
	Notes       string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateSupplierRequest is the input for creating a supplier.
type CreateSupplierRequest struct {
	Name        string
	RIF         string
	ContactName string
	Email       string
	Phone       string
	Address     string
	Notes       string
}

// UpdateSupplierRequest carries the fields that can be changed.
type UpdateSupplierRequest struct {
	Name        *string
	RIF         *string
	ContactName *string
	Email       *string
	Phone       *string
	Address     *string
	Notes       *string
	Active      *bool
}

func (r *CreateSupplierRequest) Validate() error {
	if r.Name == "" {
		return ErrSupplierNameRequired
	}
	return nil
}

// ─── Purchase Order ───────────────────────────────────────────────────────────

// POStatus represents the lifecycle of a purchase order.
type POStatus string

const (
	POStatusDraft     POStatus = "draft"
	POStatusOrdered   POStatus = "ordered"
	POStatusReceived  POStatus = "received"
	POStatusCancelled POStatus = "cancelled"
)

// PurchaseOrder is an aggregate root representing a purchase from a supplier.
type PurchaseOrder struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	SupplierID uuid.UUID
	Status     POStatus
	Notes      string
	Total      float64
	OrderedAt  *time.Time
	ReceivedAt *time.Time
	CreatedBy  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Lines      []POLine
	Supplier   *Supplier // optional, populated on demand
}

// POLine is a single product line in a purchase order.
type POLine struct {
	ID              uuid.UUID
	POID            uuid.UUID
	ProductID       uuid.UUID
	Description     string
	QuantityOrdered float64
	UnitCost        float64
	Subtotal        float64
	SortOrder       int
}

// CreatePOLineRequest is the input for one purchase order line.
type CreatePOLineRequest struct {
	ProductID   uuid.UUID
	Description string
	Quantity    float64
	UnitCost    float64
}

// CreatePORequest is the full purchase order creation payload.
type CreatePORequest struct {
	SupplierID uuid.UUID
	Notes      string
	Lines      []CreatePOLineRequest
}

func (r *CreatePORequest) Validate() error {
	if r.SupplierID == uuid.Nil {
		return fmt.Errorf("supplier_id is required")
	}
	if len(r.Lines) == 0 {
		return ErrEmptyPO
	}
	for _, l := range r.Lines {
		if l.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive")
		}
		if l.UnitCost <= 0 {
			return fmt.Errorf("unit_cost must be positive")
		}
	}
	return nil
}

// ─── Domain errors ────────────────────────────────────────────────────────────

var (
	ErrSupplierNotFound     = errors.New("supplier not found")
	ErrSupplierNameRequired = errors.New("supplier name is required")
	ErrPONotFound           = errors.New("purchase order not found")
	ErrEmptyPO              = errors.New("purchase order must have at least one line")
	ErrPOAlreadyOrdered     = errors.New("purchase order is already ordered")
	ErrPOAlreadyReceived    = errors.New("purchase order is already received")
	ErrPOAlreadyCancelled   = errors.New("purchase order is already cancelled")
	ErrPONotOrdered         = errors.New("purchase order is not in ordered status")
	ErrPONotDraft           = errors.New("purchase order is not in draft status")
)
