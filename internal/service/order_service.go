package service

import (
	"context"
	"fmt"

	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, items []domain.OrderItem, couponCode *string) (*domain.Order, []domain.Product, error)
}

type orderService struct {
	productRepo domain.ProductRepository
	orderRepo   domain.OrderRepository
}

func NewOrderService(orderRepo domain.OrderRepository, productRepo domain.ProductRepository) OrderService {
	return &orderService{
		productRepo: productRepo,
		orderRepo:   orderRepo,
	}
}

func (s *orderService) CreateOrder(
	ctx context.Context,
	items []domain.OrderItem,
	couponCode *string,
) (*domain.Order, []domain.Product, error) {
	// 1. Collect unique product IDs from the order items
	uniqueIDsMap := make(map[domain.ProductID]struct{})
	for _, item := range items {
		uniqueIDsMap[item.ProductID] = struct{}{}
	}

	uniqueIDs := make([]domain.ProductID, 0, len(uniqueIDsMap))
	for id := range uniqueIDsMap {
		uniqueIDs = append(uniqueIDs, id)
	}

	// 2. Batch fetch from product repo
	productsByID, err := s.productRepo.GetProductByIDs(ctx, uniqueIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("lookup products for order: %w", err)
	}

	// 3. Ensure all products exist
	for _, item := range items {
		if _, ok := productsByID[item.ProductID]; !ok {
			return nil, nil, fmt.Errorf("product %s does not exist: %w", item.ProductID, domain.ErrProductNotFound)
		}
	}

	// 4. Generate a new OrderID
	newOrderID := domain.OrderID(uuid.NewString())

	order, err := domain.NewOrder(newOrderID, items, couponCode)
	if err != nil {
		return nil, nil, err
	}

	// 5. Persist into DB
	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, nil, fmt.Errorf("persist order: %w", err)
	}

	// 6. Prepare the slice of products in a consistent order
	products := make([]domain.Product, 0, len(productsByID))
	for _, item := range items {
		// This preserves the order as used in the request
		if p, ok := productsByID[item.ProductID]; ok {
			products = append(products, p)
		}
	}

	return order, products, nil
}
