package storage

import (
	"context"

	"github.com/M-Arthur/order-food-api/internal/domain"
)

// InMemoryProductRepository is an adapter implementing domain.ProductRepository.
//
// It is read-only after construction and is safe for concurrent access.
type InMemoryProductRepository struct {
	products map[domain.ProductID]domain.Product
}

func NewInMemoryProductRepository(seed []domain.Product) domain.ProductRepository {
	m := make(map[domain.ProductID]domain.Product)
	for _, p := range seed {
		m[p.ID] = p
	}
	return &InMemoryProductRepository{
		products: m,
	}
}

func (r *InMemoryProductRepository) ListProducts(_ context.Context) ([]domain.Product, error) {
	out := make([]domain.Product, 0, len(r.products))
	for _, p := range r.products {
		out = append(out, p)
	}
	return out, nil
}

func (r *InMemoryProductRepository) GetProductByID(_ context.Context, id domain.ProductID) (*domain.Product, error) {
	p, ok := r.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}

	cp := p // copy to avoid mutation of internal state
	return &cp, nil
}
