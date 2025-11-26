package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, []domain.Product, error)
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
	input *domain.Order,
) (*domain.Order, []domain.Product, error) {
	// 1. Validate product existence and collect domain.Product
	var products []domain.Product
	for _, item := range input.Items {
		p, err := s.productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, domain.ErrProductNotFound) {
				return nil, nil, fmt.Errorf("product %s does not exist: %w", item.ProductID, domain.ErrProductNotFound)
			}

			return nil, nil, fmt.Errorf("lookup product %s failed: %w", item.ProductID, err)
		}

		products = append(products, *p)
	}

	// 2. Create new OrderID
	newOrderID := domain.OrderID(uuid.NewString())

	order, err := domain.NewOrder(newOrderID, input.Items, input.CouponCode)
	if err != nil {
		return nil, nil, err
	}

	// 3. Persist into DB
	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, nil, fmt.Errorf("persist order: %w", err)
	}

	return order, products, nil
}
