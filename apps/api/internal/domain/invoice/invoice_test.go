package invoice_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/danzt/daas/api/internal/domain/invoice"
)

// ─── CreateRequest.Validate ───────────────────────────────────────────────────

func TestCreateRequest_Validate(t *testing.T) {
	productID := uuid.New()

	tests := []struct {
		name    string
		req     invoice.CreateRequest
		wantErr error
	}{
		{
			name: "valid single non-fiscal line",
			req: invoice.CreateRequest{
				Lines: []invoice.CreateLineRequest{
					{ProductID: productID, Quantity: 2, UnitPrice: 10, IsFiscal: false},
				},
			},
			wantErr: nil,
		},
		{
			name:    "empty lines returns ErrEmptyInvoice",
			req:     invoice.CreateRequest{Lines: nil},
			wantErr: invoice.ErrEmptyInvoice,
		},
		{
			name: "fiscal product returns ErrFiscalProductForbidden",
			req: invoice.CreateRequest{
				Lines: []invoice.CreateLineRequest{
					{ProductID: productID, Quantity: 1, UnitPrice: 5, IsFiscal: true},
				},
			},
			wantErr: invoice.ErrFiscalProductForbidden,
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

// ─── FormatCorrelative ────────────────────────────────────────────────────────

func TestFormatCorrelative(t *testing.T) {
	tests := []struct {
		year, seq int
		want      string
	}{
		{2026, 1, "INT-2026-00001"},
		{2026, 999, "INT-2026-00999"},
		{2026, 10000, "INT-2026-10000"},
		{2027, 42, "INT-2027-00042"},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d-%d", tc.year, tc.seq), func(t *testing.T) {
			got := invoice.FormatCorrelative(tc.year, tc.seq)
			if got != tc.want {
				t.Errorf("FormatCorrelative(%d, %d) = %q, want %q", tc.year, tc.seq, got, tc.want)
			}
		})
	}
}

// ─── Error sentinel identity ──────────────────────────────────────────────────

func TestErrors_Sentinel(t *testing.T) {
	sentinels := []error{
		invoice.ErrInvoiceNotFound,
		invoice.ErrInvoiceAlreadyIssued,
		invoice.ErrInvoiceAlreadyCancelled,
		invoice.ErrFiscalProductForbidden,
		invoice.ErrEmptyInvoice,
		invoice.ErrInvoiceNotDraft,
	}
	for _, err := range sentinels {
		if err == nil {
			t.Errorf("sentinel error must not be nil")
		}
	}
}

// ─── CustomerIDType constants ─────────────────────────────────────────────────

func TestCustomerIDType_Values(t *testing.T) {
	types := []invoice.CustomerIDType{
		invoice.CustomerIDTypeCedula,
		invoice.CustomerIDTypeRIF,
		invoice.CustomerIDTypePassport,
		invoice.CustomerIDTypeAnonymous,
	}
	for _, ct := range types {
		if ct == "" {
			t.Error("CustomerIDType constant must not be empty")
		}
	}
}

// ─── Status constants ─────────────────────────────────────────────────────────

func TestStatus_Values(t *testing.T) {
	statuses := []invoice.Status{
		invoice.StatusDraft,
		invoice.StatusIssued,
		invoice.StatusCancelled,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("Status constant must not be empty")
		}
	}
}
