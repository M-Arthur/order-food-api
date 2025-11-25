package api_test

import (
	"errors"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/api"
	"github.com/M-Arthur/order-food-api/internal/domain"
)

func TestMapOrderReqToDomain_Success(t *testing.T) {
	req := api.OrderReqDTO{
		CouponCode: ptr("PROMO10"),
		Items: []api.OrderItemDTO{
			{ProductID: "p1", Quantity: 2},
			{ProductID: "p2", Quantity: 1},
		},
	}

	orderID := domain.OrderID("order-123")

	order, err := api.MapOrderReqToDomain(orderID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if orderID != order.ID {
		t.Fatalf("order.ID = %s, want %s", order.ID, orderID)
	}

	if len(order.Items) != len(req.Items) {
		t.Fatalf("order.Items len = %d, want %d", len(order.Items), len(req.Items))
	}

	if order.CouponCode == nil || *order.CouponCode != *req.CouponCode {
		t.Fatalf("order.CouponCode = %v, want %s", order.CouponCode, *req.CouponCode)
	}

	for i, item := range order.Items {
		want := req.Items[i]
		if string(item.ProductID) != want.ProductID {
			t.Errorf("item[%d].ProductID = %s, want %s", i, item.ProductID, want.ProductID)
		}
		if item.Quantity != want.Quantity {
			t.Errorf("item[%d].Quantity = %d, want %d", i, item.Quantity, want.Quantity)
		}
	}
}

func TestMapOrderReqToDomain_ValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		req           api.OrderReqDTO
		wantErr       bool
		wantField     string
		wantMessage   string
		wantDomainErr error
	}{
		{
			name: "missing product id",
			req: api.OrderReqDTO{
				Items: []api.OrderItemDTO{
					{ProductID: "", Quantity: 1},
				},
			},
			wantErr:     true,
			wantField:   "items[0].productId",
			wantMessage: "required",
		},
		{
			name: "invalid quantity",
			req: api.OrderReqDTO{
				Items: []api.OrderItemDTO{
					{ProductID: "p1", Quantity: 0},
				},
			},
			wantErr:     true,
			wantField:   "items[0].quantity",
			wantMessage: "must be >= 1",
		},
		{
			name: "empty items - domain will reject",
			req: api.OrderReqDTO{
				Items: nil,
			},
			wantErr:       true,
			wantDomainErr: domain.ErrEmptyOrderItems,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := api.MapOrderReqToDomain(domain.OrderID("order-1"), tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MapOrderReqToDomain() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("did not expect error, got %v", err)
				}
				return
			}

			// If we expect a ValidationError (field + message).
			if tt.wantField != "" {
				var ve *api.ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("expected ValidationError, got %T (%v)", err, err)
				}
				if ve.Field != tt.wantField {
					t.Errorf("ValidationError.Field = %q, want %q", ve.Field, tt.wantField)
				}
				if ve.Message != tt.wantMessage {
					t.Errorf("ValidationError.Message = %q, want %q", ve.Message, tt.wantMessage)
				}
				return
			}

			// If we expect a specific domain error.
			if tt.wantDomainErr != nil && !errors.Is(err, tt.wantDomainErr) {
				t.Fatalf("expected error %v, got %v", tt.wantDomainErr, err)
			}
		})
	}
}

func TestMapDomainProductToDTO(t *testing.T) {
	p := domain.Product{
		ID:       domain.ProductID("10"),
		Name:     "Chicken Waffle",
		Price:    domain.NewMoneyFromFloat(12.5),
		Category: "Waffle",
	}

	dto := api.MapDomainProductToDTO(p)

	if dto.ID != string(p.ID) {
		t.Errorf("dto.ID = %s, want %s", dto.ID, p.ID)
	}
	if dto.Name != p.Name {
		t.Errorf("dto.Name = %s, want %s", dto.Name, p.Name)
	}
	if dto.Category != p.Category {
		t.Errorf("dto.Category = %s, want %s", dto.Category, p.Category)
	}
	if dto.Price != p.Price.ToFloat() {
		t.Errorf("dto.Price = %v, want %v", dto.Price, p.Price.ToFloat())
	}
}

func TestMapDomainOrderToDTO(t *testing.T) {
	order := &domain.Order{
		ID: "order-123",
		Items: []domain.OrderItem{
			{ProductID: "10", Quantity: 2},
			{ProductID: "11", Quantity: 3},
		},
	}

	products := []domain.Product{
		{ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.0), Category: "Waffle"},
		{ID: "11", Name: "Fries", Price: domain.NewMoneyFromFloat(5.5), Category: "Sides"},
	}

	dto := api.MapDomainOrderToDTO(order, products)

	if dto.ID != string(order.ID) {
		t.Errorf("dto.ID = %s, want %s", dto.ID, order.ID)
	}

	if len(dto.Items) != len(order.Items) {
		t.Fatalf("dto.Items len = %d, want %d", len(dto.Items), len(order.Items))
	}
	for i, item := range dto.Items {
		want := order.Items[i]
		if item.ProductID != string(want.ProductID) {
			t.Errorf("item[%d].ProductID = %s, want %s", i, item.ProductID, want.ProductID)
		}
		if item.Quantity != want.Quantity {
			t.Errorf("item[%d].Quantity = %d, want %d", i, item.Quantity, want.Quantity)
		}
	}

	if len(dto.Products) != len(products) {
		t.Errorf("dto.Products len = %d, want %d", len(dto.Products), len(products))
	}
	for i, p := range dto.Products {
		want := products[i]
		if p.ID != string(want.ID) {
			t.Errorf("products[%d].ID = %s, want %s", i, p.ID, want.ID)
		}
		if p.Name != want.Name {
			t.Errorf("products[%d].Name = %s, want %s", i, p.Name, want.Name)
		}
		if p.Category != want.Category {
			t.Errorf("products[%d].Category = %s, want %s", i, p.Category, want.Category)
		}
		if p.Price != want.Price.ToFloat() {
			t.Errorf("products[%d].Price = %v, want %v", i, p.Price, want.Price.ToFloat())
		}
	}
}

func ptr[T any](v T) *T {
	return &v
}
