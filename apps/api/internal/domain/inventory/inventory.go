// Package inventory contains the domain entities and business rules for
// inventory control. Zero imports from Echo, pgx, or any external framework.
package inventory

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// MovementType classifies the direction and intent of a stock change.
type MovementType string

const (
	MovementEntry      MovementType = "entry"
	MovementExit       MovementType = "exit"
	MovementAdjustment MovementType = "adjustment"
)

// ReferenceType identifies what triggered the movement.
type ReferenceType string

const (
	RefPurchaseOrder ReferenceType = "purchase_order"
	RefSale          ReferenceType = "sale"
	RefManualAdjust  ReferenceType = "manual_adjustment"
	RefOpening       ReferenceType = "opening"
)

// Movement is an immutable record of a stock change.
type Movement struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	ProductID     uuid.UUID
	Type          MovementType
	Quantity      float64 // always the absolute delta; direction comes from Type
	UnitCost      *float64
	ReferenceType ReferenceType
	ReferenceID   *uuid.UUID
	Notes         string
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
}

// Stock is the current on-hand quantity for a product.
type Stock struct {
	ProductID      uuid.UUID
	TenantID       uuid.UUID
	QuantityOnHand float64
	LastUpdatedAt  time.Time
}

// AdjustmentRequest is the input for a manual stock adjustment.
type AdjustmentRequest struct {
	ProductID uuid.UUID
	Delta     float64 // positive = increase, negative = decrease
	Notes     string
}

// Validate checks business rules for a manual adjustment.
func (r *AdjustmentRequest) Validate() error {
	if r.Notes == "" {
		return ErrAdjustmentNotesMissing
	}
	if r.Delta == 0 {
		return ErrZeroDelta
	}
	return nil
}

// MovementType for a given delta — positive delta is adjustment (increase treated
// as entry, decrease as exit internally, but we always persist as "adjustment").
func (r *AdjustmentRequest) MovementType() MovementType {
	return MovementAdjustment
}

// Errors
var (
	ErrAdjustmentNotesMissing = errors.New("adjustment notes are required")
	ErrZeroDelta              = errors.New("adjustment delta must not be zero")
	ErrInsufficientStock      = errors.New("insufficient stock")
	ErrMovementNotFound       = errors.New("movement not found")
	ErrStockNotFound          = errors.New("stock record not found")
)
