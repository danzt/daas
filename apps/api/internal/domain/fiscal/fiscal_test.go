package fiscal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/danzt/daas/api/internal/domain/fiscal"
)

// ─── CreateRequest.Validate ───────────────────────────────────────────────────

func TestCreateRequest_Validate(t *testing.T) {
	productID := uuid.New()

	tests := []struct {
		name    string
		req     fiscal.CreateRequest
		wantErr error
	}{
		{
			name: "valid single fiscal line",
			req: fiscal.CreateRequest{
				Lines: []fiscal.CreateLineRequest{
					{ProductID: productID, Quantity: 1, UnitPrice: 10, TaxRate: 0.16, IsFiscal: true},
				},
			},
			wantErr: nil,
		},
		{
			name:    "empty lines returns ErrEmptyInvoice",
			req:     fiscal.CreateRequest{Lines: nil},
			wantErr: fiscal.ErrEmptyInvoice,
		},
		{
			name: "non-fiscal product returns ErrNonFiscalProductForbidden",
			req: fiscal.CreateRequest{
				Lines: []fiscal.CreateLineRequest{
					{ProductID: productID, Quantity: 1, UnitPrice: 5, TaxRate: 0, IsFiscal: false},
				},
			},
			wantErr: fiscal.ErrNonFiscalProductForbidden,
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

// ─── ComputeTotals ────────────────────────────────────────────────────────────

func TestComputeTotals(t *testing.T) {
	lines := []fiscal.CreateLineRequest{
		{Quantity: 2, UnitPrice: 100, TaxRate: 0.16}, // base=200, tax=32
		{Quantity: 1, UnitPrice: 50, TaxRate: 0.16},  // base=50,  tax=8
	}

	base, tax, total := fiscal.ComputeTotals(lines)

	if base != 250 {
		t.Errorf("subtotal_base = %.2f, want 250.00", base)
	}
	if tax != 40 {
		t.Errorf("tax_amount = %.2f, want 40.00", tax)
	}
	if total != 290 {
		t.Errorf("total = %.2f, want 290.00", total)
	}
}

func TestComputeTotals_ZeroTax(t *testing.T) {
	lines := []fiscal.CreateLineRequest{
		{Quantity: 3, UnitPrice: 10, TaxRate: 0},
	}
	base, tax, total := fiscal.ComputeTotals(lines)
	if base != 30 || tax != 0 || total != 30 {
		t.Errorf("ComputeTotals with 0 tax: got base=%.2f tax=%.2f total=%.2f, want 30/0/30", base, tax, total)
	}
}

// ─── Status constants ─────────────────────────────────────────────────────────

func TestStatus_Values(t *testing.T) {
	statuses := []fiscal.Status{
		fiscal.StatusDraft,
		fiscal.StatusPendingFiscal,
		fiscal.StatusIssued,
		fiscal.StatusFailed,
		fiscal.StatusCancelled,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("Status constant must not be empty")
		}
	}
}

// ─── Error sentinel identity ──────────────────────────────────────────────────

func TestErrors_Sentinel(t *testing.T) {
	sentinels := []error{
		fiscal.ErrInvoiceNotFound,
		fiscal.ErrInvoiceAlreadyIssued,
		fiscal.ErrInvoiceAlreadyCancelled,
		fiscal.ErrNonFiscalProductForbidden,
		fiscal.ErrEmptyInvoice,
		fiscal.ErrInvoiceNotDraft,
		fiscal.ErrInvoiceCannotRetry,
	}
	for _, err := range sentinels {
		if err == nil {
			t.Errorf("sentinel error must not be nil")
		}
	}
}
