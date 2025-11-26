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
}

func NewOrderService(pr domain.ProductRepository) OrderService {
	return &orderService{
		productRepo: pr,
	}
}

func (s *orderService) CreateOrder(
	ctx context.Context,
	order *domain.Order,
) (*domain.Order, []domain.Product, error) {
	// 1. Validate tat all product IDs exist
	var products []domain.Product
	for _, item := range order.Items {
		p, err := s.productRepo.GetProductByID(item.ProductID)
		if err != nil {
			if errors.Is(err, domain.ErrProductNotFound) {
				return nil, nil, fmt.Errorf("product %s does not exist: %w", item.ProductID, domain.ErrProductNotFound)
			}

			return nil, nil, fmt.Errorf("lookup product %s failed: %w", item.ProductID, err)
		}

		products = append(products, *p)
	}

	// 2.Create new OrderID
	newOrderID := domain.OrderID(uuid.NewString())

	order, err := domain.NewOrder(newOrderID, order.Items, order.CouponCode)
	if err != nil {
		return nil, nil, err
	}

	return order, products, nil
}
