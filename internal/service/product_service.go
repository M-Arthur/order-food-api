package service

import (
	"context"

	"github.com/M-Arthur/order-food-api/internal/domain"
)

// ProductService defines application-level operations for products.
type ProductService interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
}

type productService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) ProductService {
	return &productService{
		repo: repo,
	}
}

func (s *productService) ListProducts(ctx context.Context) ([]domain.Product, error) {
	// Business logic would go here in future (filtering, sorting, etc.)
	return s.repo.ListProducts()
}
