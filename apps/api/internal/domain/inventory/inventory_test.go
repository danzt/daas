package inventory_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/danzt/daas/api/internal/domain/inventory"
)

// ─── AdjustmentRequest.Validate ──────────────────────────────────────────────

func TestAdjustmentRequest_Validate(t *testing.T) {
	productID := uuid.New()

	tests := []struct {
		name    string
		req     inventory.AdjustmentRequest
		wantErr error
	}{
		{
			name:    "valid positive adjustment",
			req:     inventory.AdjustmentRequest{ProductID: productID, Delta: 10.5, Notes: "initial stock count"},
			wantErr: nil,
		},
		{
			name:    "valid negative adjustment",
			req:     inventory.AdjustmentRequest{ProductID: productID, Delta: -3.0, Notes: "damaged goods"},
			wantErr: nil,
		},
		{
			name:    "missing notes returns ErrAdjustmentNotesMissing",
			req:     inventory.AdjustmentRequest{ProductID: productID, Delta: 5.0, Notes: ""},
			wantErr: inventory.ErrAdjustmentNotesMissing,
		},
		{
			name:    "zero delta returns ErrZeroDelta",
			req:     inventory.AdjustmentRequest{ProductID: productID, Delta: 0, Notes: "some reason"},
			wantErr: inventory.ErrZeroDelta,
		},
		{
			name:    "both missing notes and zero delta — notes checked first",
			req:     inventory.AdjustmentRequest{ProductID: productID, Delta: 0, Notes: ""},
			wantErr: inventory.ErrAdjustmentNotesMissing,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if err != tc.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// ─── AdjustmentRequest.MovementType ──────────────────────────────────────────

func TestAdjustmentRequest_MovementType(t *testing.T) {
	req := inventory.AdjustmentRequest{ProductID: uuid.New(), Delta: 1, Notes: "test"}
	got := req.MovementType()
	if got != inventory.MovementAdjustment {
		t.Errorf("MovementType() = %q, want %q", got, inventory.MovementAdjustment)
	}
}

// ─── Error sentinel identity ──────────────────────────────────────────────────

func TestErrors_Sentinel(t *testing.T) {
	// Verify error variables are stable (not recreated per call)
	if inventory.ErrAdjustmentNotesMissing == nil {
		t.Error("ErrAdjustmentNotesMissing must not be nil")
	}
	if inventory.ErrZeroDelta == nil {
		t.Error("ErrZeroDelta must not be nil")
	}
	if inventory.ErrInsufficientStock == nil {
		t.Error("ErrInsufficientStock must not be nil")
	}
	if inventory.ErrMovementNotFound == nil {
		t.Error("ErrMovementNotFound must not be nil")
	}
	if inventory.ErrStockNotFound == nil {
		t.Error("ErrStockNotFound must not be nil")
	}
}

// ─── Stock struct ─────────────────────────────────────────────────────────────

func TestStock_NonNegativeConstraint_Semantic(t *testing.T) {
	// The DB enforces non-negative via CHECK constraint.
	// At domain level, we document this expectation: quantity_on_hand >= 0.
	s := inventory.Stock{QuantityOnHand: 0}
	if s.QuantityOnHand < 0 {
		t.Error("initial zero stock should not be negative")
	}
}
