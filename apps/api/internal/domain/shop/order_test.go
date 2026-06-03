package shop_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/danzt/daas/api/internal/domain/shop"
)

// ─── CreateOrderRequest.Validate ─────────────────────────────────────────────

func TestCreateOrderRequest_Validate_HappyPath(t *testing.T) {
	req := shop.CreateOrderRequest{
		CustomerName:    "Ana Pérez",
		CustomerEmail:   "ana@example.com",
		ShippingAddress: "Av. Principal 123",
		Lines: []shop.CreateOrderLineRequest{
			{ProductID: uuid.New(), Quantity: 2},
		},
	}
	if err := req.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestCreateOrderRequest_Validate_MissingCustomerName(t *testing.T) {
	req := shop.CreateOrderRequest{
		CustomerName:    "",
		CustomerEmail:   "ana@example.com",
		ShippingAddress: "Av. Principal 123",
		Lines:           []shop.CreateOrderLineRequest{{ProductID: uuid.New(), Quantity: 1}},
	}
	if err := req.Validate(); err == nil {
		t.Error("Validate() expected error for missing customer name, got nil")
	}
}

func TestCreateOrderRequest_Validate_MissingEmail(t *testing.T) {
	req := shop.CreateOrderRequest{
		CustomerName:    "Ana Pérez",
		CustomerEmail:   "",
		ShippingAddress: "Av. Principal 123",
		Lines:           []shop.CreateOrderLineRequest{{ProductID: uuid.New(), Quantity: 1}},
	}
	if err := req.Validate(); err == nil {
		t.Error("Validate() expected error for missing email, got nil")
	}
}

func TestCreateOrderRequest_Validate_MissingAddress(t *testing.T) {
	req := shop.CreateOrderRequest{
		CustomerName:    "Ana Pérez",
		CustomerEmail:   "ana@example.com",
		ShippingAddress: "",
		Lines:           []shop.CreateOrderLineRequest{{ProductID: uuid.New(), Quantity: 1}},
	}
	if err := req.Validate(); err == nil {
		t.Error("Validate() expected error for missing shipping address, got nil")
	}
}

func TestCreateOrderRequest_Validate_EmptyLines(t *testing.T) {
	req := shop.CreateOrderRequest{
		CustomerName:    "Ana Pérez",
		CustomerEmail:   "ana@example.com",
		ShippingAddress: "Av. Principal 123",
		Lines:           nil,
	}
	if err := req.Validate(); err == nil {
		t.Error("Validate() expected error for empty lines, got nil")
	}
}

func TestCreateOrderRequest_Validate_NegativeQuantity(t *testing.T) {
	req := shop.CreateOrderRequest{
		CustomerName:    "Ana Pérez",
		CustomerEmail:   "ana@example.com",
		ShippingAddress: "Av. Principal 123",
		Lines: []shop.CreateOrderLineRequest{
			{ProductID: uuid.New(), Quantity: -1},
		},
	}
	if err := req.Validate(); err == nil {
		t.Error("Validate() expected error for negative quantity, got nil")
	}
}

// ─── ShopOrder.CanTransitionTo ────────────────────────────────────────────────

func TestShopOrder_CanTransitionTo_PendingToPaid(t *testing.T) {
	o := shop.ShopOrder{Status: shop.OrderStatusPending}
	if err := o.CanTransitionTo(shop.OrderStatusPaid); err != nil {
		t.Errorf("pending → paid should be allowed, got: %v", err)
	}
}

func TestShopOrder_CanTransitionTo_PendingToCancelled(t *testing.T) {
	o := shop.ShopOrder{Status: shop.OrderStatusPending}
	if err := o.CanTransitionTo(shop.OrderStatusCancelled); err != nil {
		t.Errorf("pending → cancelled should be allowed, got: %v", err)
	}
}

func TestShopOrder_CanTransitionTo_PendingToDelivered(t *testing.T) {
	o := shop.ShopOrder{Status: shop.OrderStatusPending}
	if err := o.CanTransitionTo(shop.OrderStatusDelivered); err == nil {
		t.Error("pending → delivered should NOT be allowed, got nil error")
	}
}

func TestShopOrder_CanTransitionTo_PaidToFulfilled(t *testing.T) {
	o := shop.ShopOrder{Status: shop.OrderStatusPaid}
	if err := o.CanTransitionTo(shop.OrderStatusFulfilled); err != nil {
		t.Errorf("paid → fulfilled should be allowed, got: %v", err)
	}
}

func TestShopOrder_CanTransitionTo_FulfilledToDelivered(t *testing.T) {
	o := shop.ShopOrder{Status: shop.OrderStatusFulfilled}
	if err := o.CanTransitionTo(shop.OrderStatusDelivered); err != nil {
		t.Errorf("fulfilled → delivered should be allowed, got: %v", err)
	}
}

func TestShopOrder_CanTransitionTo_DeliveredToAnything(t *testing.T) {
	o := shop.ShopOrder{Status: shop.OrderStatusDelivered}
	targets := []shop.OrderStatus{
		shop.OrderStatusPending,
		shop.OrderStatusPaid,
		shop.OrderStatusFulfilled,
		shop.OrderStatusCancelled,
	}
	for _, target := range targets {
		if err := o.CanTransitionTo(target); err == nil {
			t.Errorf("delivered → %s should NOT be allowed (terminal state), got nil error", target)
		}
	}
}

func TestShopOrder_CanTransitionTo_CancelledToAnything(t *testing.T) {
	o := shop.ShopOrder{Status: shop.OrderStatusCancelled}
	targets := []shop.OrderStatus{
		shop.OrderStatusPending,
		shop.OrderStatusPaid,
		shop.OrderStatusFulfilled,
		shop.OrderStatusDelivered,
	}
	for _, target := range targets {
		if err := o.CanTransitionTo(target); err == nil {
			t.Errorf("cancelled → %s should NOT be allowed (terminal state), got nil error", target)
		}
	}
}
