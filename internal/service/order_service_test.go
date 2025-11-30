package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/service"
)

// stubProductRepoForOrder implements domain.ProductRepository for OrderService tests.
type stubProductRepoForOrder struct {
	productsByID map[domain.ProductID]domain.Product
	err          error
}

func (s *stubProductRepoForOrder) ListProducts(ctx context.Context) ([]domain.Product, error) {
	panic("ListProducts should not be called in OrderService tests")
}

func (s *stubProductRepoForOrder) GetProductByID(ctx context.Context, id domain.ProductID) (*domain.Product, error) {
	if s.err != nil {
		return nil, s.err
	}

	p, ok := s.productsByID[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return &p, nil
}

func (s *stubProductRepoForOrder) GetProductByIDs(ctx context.Context, ids []domain.ProductID) (map[domain.ProductID]domain.Product, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.productsByID == nil {
		return map[domain.ProductID]domain.Product{}, nil
	}
	// Simulate batch fetch: return only those known
	out := make(map[domain.ProductID]domain.Product, len(ids))
	for _, id := range ids {
		if p, ok := s.productsByID[id]; ok {
			out[id] = p
		}
	}
	return out, nil
}

// stubOrderRepo implements domain.OrderRepository for Orderservice tests
type stubOrderRepo struct {
	savedOrder *domain.Order
	saveErr    error
	saveCalls  int
}

func (s *stubOrderRepo) Save(ctx context.Context, order *domain.Order) error {
	s.saveCalls++
	s.savedOrder = order
	return s.saveErr
}

// complie-time checks
var (
	_ domain.ProductRepository = (*stubProductRepoForOrder)(nil)
	_ domain.OrderRepository   = (*stubOrderRepo)(nil)
)

func TestOrderService_CreateOrder_Success(t *testing.T) {
	ctx := context.Background()

	productsByID := map[domain.ProductID]domain.Product{
		"10": {ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		"11": {ID: "11", Name: "Fries", Price: domain.NewMoneyFromFloat(5.0), Category: "Sides"},
	}

	productRepo := &stubProductRepoForOrder{productsByID: productsByID}
	orderRepo := &stubOrderRepo{}

	svc := service.NewOrderService(orderRepo, productRepo)

	items := []domain.OrderItem{
		{ProductID: "10", Quantity: 2},
		{ProductID: "11", Quantity: 3},
	}
	coupon := ptr("PROMO10")

	order, products, err := svc.CreateOrder(ctx, items, coupon)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v, want nil", err)
	}

	if order == nil {
		t.Fatalf("CreateOrder() return nil order")
	}

	if order.ID == "" {
		t.Fatalf("order.ID is empty, want non-empty")
	}

	if len(order.Items) != len(items) {
		t.Fatalf("len(order.Items) = %d, want %d", len(order.Items), len(items))
	}
	for i, item := range order.Items {
		want := items[i]
		if item.ProductID != want.ProductID {
			t.Errorf("order.Items[%d].ProductID = %s, want %s", i, item.ProductID, want.ProductID)
		}
		if item.Quantity != want.Quantity {
			t.Errorf("order.Items[%d].Quantity = %d, want %d", i, item.Quantity, want.Quantity)
		}
	}

	if order.CouponCode == nil || *order.CouponCode != *coupon {
		t.Errorf("order.CouponCode = %v, want %s", order.CouponCode, *coupon)
	}

	// Products returned for response
	if len(products) != 2 {
		t.Fatalf("len(products) = %d, want %d", len(products), 2)
	}

	// Ensure order was saved exactly once
	if orderRepo.saveCalls != 1 {
		t.Fatalf("orderRepo.saveCalls = %d, want %d", orderRepo.saveCalls, 1)
	}
	if orderRepo.savedOrder == nil {
		t.Fatalf("orderRepo.savedOrder is nil, want non-nil")
	}
	if orderRepo.savedOrder.ID != order.ID {
		t.Fatalf("savedOrder.ID = %s, want %s", orderRepo.savedOrder.ID, order.ID)
	}
}

func TestOrderService_CreateOrder_ProductNotFound(t *testing.T) {
	ctx := context.Background()

	// Repo knows only product "10", but we request "11" as well.
	productsByID := map[domain.ProductID]domain.Product{
		"10": {ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
	}

	productRepo := &stubProductRepoForOrder{productsByID: productsByID}
	orderRepo := &stubOrderRepo{}

	svc := service.NewOrderService(orderRepo, productRepo)

	items := []domain.OrderItem{
		{ProductID: "10", Quantity: 1},
		{ProductID: "11", Quantity: 1}, // missing
	}

	_, _, err := svc.CreateOrder(ctx, items, nil)
	if err == nil {
		t.Fatalf("CreateOrder() error = nil, want non-nil")
	}
	if !errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("CreateOrder() error = %v, want to wrap %v", err, domain.ErrProductNotFound)
	}
	if orderRepo.saveCalls != 0 {
		t.Fatalf("orderRepo.saveCalls = %d, want 0 (order should not be saved)", orderRepo.saveCalls)
	}
}

func TestOrderService_CreateOrder_ProductRepoError(t *testing.T) {
	ctx := context.Background()

	repoErr := errors.New("db unavailable")
	productRepo := &stubProductRepoForOrder{
		err: repoErr,
	}
	orderRepo := &stubOrderRepo{}

	svc := service.NewOrderService(orderRepo, productRepo)

	items := []domain.OrderItem{
		{ProductID: "10", Quantity: 1},
	}

	_, _, err := svc.CreateOrder(ctx, items, nil)
	if err == nil {
		t.Fatalf("CreateOrder() error = nil, want non-nil")
	}
	if errors.Is(err, domain.ErrProductNotFound) {
		t.Fatalf("CreateOrder() error wrapped ErrProductNotFound unexpectedly: %v", err)
	}
	if orderRepo.saveCalls != 0 {
		t.Fatalf("orderRepo.saveCalls = %d, want 0", orderRepo.saveCalls)
	}
}

func TestOrderService_CreateOrder_OrderRepoError(t *testing.T) {
	ctx := context.Background()

	productsByID := map[domain.ProductID]domain.Product{
		"10": {ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
	}

	productRepo := &stubProductRepoForOrder{productsByID: productsByID}
	orderRepoErr := errors.New("insert failed")
	orderRepo := &stubOrderRepo{saveErr: orderRepoErr}

	svc := service.NewOrderService(orderRepo, productRepo)

	items := []domain.OrderItem{
		{ProductID: "10", Quantity: 1},
	}

	_, _, err := svc.CreateOrder(ctx, items, nil)
	if err == nil {
		t.Fatalf("CreateOrder() error = nil, want non-nil")
	}
	if !errors.Is(err, orderRepoErr) {
		t.Fatalf("CreateOrder() error = %v, want to wrap %v", err, orderRepoErr)
	}
	if orderRepo.saveCalls != 1 {
		t.Fatalf("orderRepo.saveCalls = %d, want 1", orderRepo.saveCalls)
	}
}

func TestOrderService_CreateOrder_InvalidQuantity(t *testing.T) {
	ctx := context.Background()

	productsByID := map[domain.ProductID]domain.Product{
		"10": {ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
	}

	productRepo := &stubProductRepoForOrder{productsByID: productsByID}
	orderRepo := &stubOrderRepo{}

	svc := service.NewOrderService(orderRepo, productRepo)

	items := []domain.OrderItem{
		{ProductID: "10", Quantity: 0}, // invalid
	}

	_, _, err := svc.CreateOrder(ctx, items, nil)
	if err == nil {
		t.Fatalf("CreateOrder() error = nil, want non-nil")
	}
	if !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Fatalf("CreateOrder() error = %v, want to wrap %v", err, domain.ErrInvalidQuantity)
	}
	if orderRepo.saveCalls != 0 {
		t.Fatalf("orderRepo.saveCalls = %d, want 0", orderRepo.saveCalls)
	}
}

func ptr[T any](v T) *T {
	return &v
}
