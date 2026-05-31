package sale_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/danzt/daas/api/internal/domain/sale"
)

// ─── CreateOrderRequest.Validate ─────────────────────────────────────────────

func TestCreateOrderRequest_Validate(t *testing.T) {
	productID := uuid.New()

	validLine := sale.CreateOrderLineRequest{
		ProductID:   productID,
		Description: "Producto A",
		Quantity:    2,
		UnitPrice:   15,
	}

	tests := []struct {
		name    string
		req     sale.CreateOrderRequest
		wantErr bool
	}{
		{
			name:    "valid order with one line",
			req:     sale.CreateOrderRequest{Lines: []sale.CreateOrderLineRequest{validLine}},
			wantErr: false,
		},
		{
			name:    "empty lines returns ErrEmptyOrder",
			req:     sale.CreateOrderRequest{Lines: nil},
			wantErr: true,
		},
		{
			name: "zero quantity in line",
			req: sale.CreateOrderRequest{
				Lines: []sale.CreateOrderLineRequest{
					{ProductID: productID, Description: "test", Quantity: 0, UnitPrice: 10},
				},
			},
			wantErr: true,
		},
		{
			name: "zero unit_price in line",
			req: sale.CreateOrderRequest{
				Lines: []sale.CreateOrderLineRequest{
					{ProductID: productID, Description: "test", Quantity: 1, UnitPrice: 0},
				},
			},
			wantErr: true,
		},
		{
			name: "multiple valid lines",
			req: sale.CreateOrderRequest{
				CustomerName: "Cliente XYZ",
				Lines: []sale.CreateOrderLineRequest{
					validLine,
					{ProductID: uuid.New(), Description: "Producto B", Quantity: 5, UnitPrice: 3},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

// ─── OrderStatus constants ────────────────────────────────────────────────────

func TestOrderStatus_Values(t *testing.T) {
	statuses := []sale.OrderStatus{
		sale.OrderStatusDraft,
		sale.OrderStatusConfirmed,
		sale.OrderStatusInvoiced,
		sale.OrderStatusCancelled,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("OrderStatus constant must not be empty")
		}
	}
}

// ─── Error sentinel identity ──────────────────────────────────────────────────

func TestErrors_Sentinel(t *testing.T) {
	sentinels := []error{
		sale.ErrOrderNotFound,
		sale.ErrEmptyOrder,
		sale.ErrOrderAlreadyConfirmed,
		sale.ErrOrderAlreadyInvoiced,
		sale.ErrOrderAlreadyCancelled,
		sale.ErrOrderNotConfirmed,
		sale.ErrOrderNotDraft,
		sale.ErrInsufficientStock,
	}
	for _, err := range sentinels {
		if err == nil {
			t.Errorf("sentinel error must not be nil")
		}
	}
}
