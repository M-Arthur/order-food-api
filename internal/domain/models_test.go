package domain_test

import (
	"math"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/domain"
)

func TestNewMoney_FromFloatAndToFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected domain.Money
	}{
		{
			name:     "simple amount",
			input:    12.34,
			expected: domain.Money(1234),
		},
		{
			name:     "rounding_half_away_from_zero",
			input:    1.005,
			expected: domain.Money(100), // banker's round
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := domain.NewMoneyFromFloat(tt.input)
			if m != tt.expected {
				t.Fatalf("NewMoneyFromFloat(%v) = %d, want %d", tt.input, m, tt.expected)
			}

			back := m.ToFloat()
			// Allow for tiny float errors on the round-trip.
			if diff := math.Abs(back - float64(m)/100); diff > 1e-9 {
				t.Fatalf("ToFloat round-trip mismatch: got %v, want %v (diff=%v)", back, float64(m)/100, diff)
			}
		})
	}
}

func TestNewOrder_Success(t *testing.T) {
	orderID := domain.OrderID("order-123")
	items := []domain.OrderItem{
		{
			ProductID: domain.ProductID("p1"),
			Quantity:  2,
		},
		{
			ProductID: domain.ProductID("p2"),
			Quantity:  1,
		},
	}
	coupon := "PROMO10"

	order, err := domain.NewOrder(orderID, items, &coupon)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if order.ID != orderID {
		t.Fatalf("order.ID = %s, want %s", order.ID, orderID)
	}
	if len(order.Items) != len(items) {
		t.Fatalf("order.Items len = %d, want %d", len(order.Items), len(items))
	}
	if order.CouponCode == nil || *order.CouponCode != coupon {
		t.Fatalf("order.CouponCode = %v, want %s", order.CouponCode, coupon)
	}

	// Ensure NewOrder defensively copies the items slice.
	items[0].Quantity = 999
	if order.Items[0].Quantity == 999 {
		t.Fatalf("order.Items was mutated after NewOrder; expected defensive copy")
	}
}

func TestNewOrder_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		orderID   domain.OrderID
		items     []domain.OrderItem
		coupon    *string
		wantError bool
	}{
		{
			name:    "empty order id",
			orderID: "",
			items: []domain.OrderItem{
				{ProductID: "p1", Quantity: 1},
			},
			wantError: true,
		},
		{
			name:      "no items",
			orderID:   "order-1",
			items:     nil,
			wantError: true,
		},
		{
			name:    "item with empty product id",
			orderID: "order-1",
			items: []domain.OrderItem{
				{ProductID: "", Quantity: 1},
			},
			wantError: true,
		},
		{
			name:    "item with invalid quantity",
			orderID: "order-1",
			items: []domain.OrderItem{
				{ProductID: "p1", Quantity: 0},
			},
			wantError: true,
		},
		{
			name:    "empty coupon string",
			orderID: "order-2",
			items: []domain.OrderItem{
				{ProductID: "p1", Quantity: 2},
			},
			coupon:    ptr(""),
			wantError: true,
		},
		{
			name:    "valid no coupon",
			orderID: "order-3",
			items: []domain.OrderItem{
				{ProductID: "p5", Quantity: 5},
			},
			coupon:    nil,
			wantError: false,
		},
	}

	for _, ts := range tests {
		tt := ts
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewOrder(tt.orderID, tt.items, tt.coupon)
			if (err != nil) != tt.wantError {
				t.Fatalf("NewOrder() error = %v, wantError = %v", err, tt.wantError)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
