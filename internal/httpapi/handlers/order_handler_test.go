package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/api"
	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/httpapi/handlers"
	"github.com/M-Arthur/order-food-api/internal/service"
)

type stubOrderService struct {
	order    *domain.Order
	products []domain.Product
	err      error
}

func (s *stubOrderService) CreateOrder(_ context.Context, _ []domain.OrderItem, _ *string) (*domain.Order, []domain.Product, error) {
	if s.err != nil {
		return nil, nil, s.err
	}

	return s.order, s.products, nil
}

// complie-time safety
var _ service.OrderService = (*stubOrderService)(nil)

func TestOrderHandler_PlaceOrder_Success(t *testing.T) {
	// The order + products the service will return
	order := &domain.Order{
		ID: "order-123",
		Items: []domain.OrderItem{
			{ProductID: "10", Quantity: 2},
			{ProductID: "11", Quantity: 3},
		},
	}
	products := []domain.Product{
		{ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		{ID: "11", Name: "Fries", Price: domain.NewMoneyFromFloat(5.0), Category: "sides"},
	}

	svc := &stubOrderService{
		order:    order,
		products: products,
	}
	h := handlers.NewOrderHandler(svc)

	reqDTO := api.OrderReqDTO{
		CouponCode: ptr("PROMO10"),
		Items: []api.OrderItemDTO{
			{ProductID: "10", Quantity: 2},
			{ProductID: "11", Quantity: 3},
		},
	}
	body, err := json.Marshal(reqDTO)
	if err != nil {
		t.Fatalf("failed to marshal request dto: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.PlaceOrder(rr, req)

	res := rr.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want %d, body=%q", res.StatusCode, http.StatusOK, string(b))
	}

	var got api.OrderDTO
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if got.ID != string(order.ID) {
		t.Fatalf("order.ID =%s, want %s", got.ID, order.ID)
	}

	if len(got.Items) != len(order.Items) {
		t.Fatalf("len(got.Items) = %d, want %d", len(got.Items), len(order.Items))
	}
	for i, p := range got.Products {
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
			t.Errorf("products[%d].ID = %v, want %v", i, p.Price, want.Price.ToFloat())
		}
	}
}

func TestOrderHandler_PlaceOrder_InvalidJSON(t *testing.T) {
	svc := &stubOrderService{}
	h := handlers.NewOrderHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewBufferString("{invalid-json"))
	rr := httptest.NewRecorder()

	h.PlaceOrder(rr, req)

	res := rr.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want %d, body=%q", res.StatusCode, http.StatusBadRequest, string(b))
	}
}

func TestOrderHandler_PlaceOrder_ValidationError(t *testing.T) {
	svc := &stubOrderService{}
	h := handlers.NewOrderHandler(svc)

	// Quantity = 0 should trigger validationError from mapper (items[0].quantity).
	reqDTO := api.OrderReqDTO{
		Items: []api.OrderItemDTO{
			{ProductID: "10", Quantity: 0},
		},
	}
	body, err := json.Marshal(reqDTO)
	if err != nil {
		t.Fatalf("failed to marshal request dto: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.PlaceOrder(rr, req)

	res := rr.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusUnprocessableEntity {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want %d, body=%q", res.StatusCode, http.StatusUnprocessableEntity, string(b))
	}
}

func TestOrderHandler_PlaceOrder_ProductNotFound(t *testing.T) {
	svc := &stubOrderService{
		err: domain.ErrProductNotFound,
	}
	h := handlers.NewOrderHandler(svc)

	reqDTO := api.OrderReqDTO{
		Items: []api.OrderItemDTO{
			{ProductID: "unknown", Quantity: 1},
		},
	}
	body, err := json.Marshal(reqDTO)
	if err != nil {
		t.Fatalf("failed to marshal request dto: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.PlaceOrder(rr, req)

	res := rr.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want %d, body=%q", res.StatusCode, http.StatusBadRequest, string(b))
	}
}

func TestOrderHandler_PlaceOrder_InternalError(t *testing.T) {
	svc := &stubOrderService{
		err: errors.New("db connection failed"),
	}
	h := handlers.NewOrderHandler(svc)

	reqDTO := api.OrderReqDTO{
		Items: []api.OrderItemDTO{
			{ProductID: "10", Quantity: 1},
		},
	}
	body, err := json.Marshal(reqDTO)
	if err != nil {
		t.Fatalf("failed to marshal reqeust dto: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/order", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.PlaceOrder(rr, req)

	res := rr.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusInternalServerError {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status = %d, want %d, body=%q", res.StatusCode, http.StatusInternalServerError, string(b))
	}
}

func ptr[T any](v T) *T {
	return &v
}
