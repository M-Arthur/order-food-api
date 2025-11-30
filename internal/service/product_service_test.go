package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/service"
)

// stubProductRepo implements domain.ProductRepository for ProductService tests.
type stubProductRepo struct {
	products []domain.Product
	err      error
}

func (s *stubProductRepo) ListProducts(ctx context.Context) ([]domain.Product, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.products, nil
}

func (s *stubProductRepo) GetProductByID(ctx context.Context, id domain.ProductID) (*domain.Product, error) {
	panic("GetProductByID should not be called in ProductService tests")
}

func (s *stubProductRepo) GetProductByIDs(ctx context.Context, ids []domain.ProductID) (map[domain.ProductID]domain.Product, error) {
	panic("GetProductByIDs should not be called in ProductService tests")
}

// compile-time check
var _ domain.ProductRepository = (*stubProductRepo)(nil)

func TestProductService_ListProducts_Success(t *testing.T) {
	ctx := context.Background()

	products := []domain.Product{
		{ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		{ID: "11", Name: "Fries", Price: domain.NewMoneyFromFloat(5.0), Category: "Sides"},
	}

	repo := &stubProductRepo{products: products}
	svc := service.NewProductService(repo)

	got, err := svc.ListProducts(ctx)
	if err != nil {
		t.Fatalf("ListProducts() error = %v, want nil", err)
	}

	if len(got) != len(products) {
		t.Fatalf("len(products) = %d, want %d", len(got), len(products))
	}

	for i, p := range got {
		want := products[i]
		if p.ID != want.ID {
			t.Errorf("products[%d].ID = %s, want %s", i, p.ID, want.ID)
		}
		if p.Name != want.Name {
			t.Errorf("products[%d].Name = %s, want %s", i, p.Name, want.Name)
		}
		if p.Category != want.Category {
			t.Errorf("products[%d].Category = %s, want %s", i, p.Category, want.Category)
		}
		if p.Price != want.Price {
			t.Errorf("products[%d].Price = %d, want %d", i, p.Price, want.Price)
		}
	}
}

func TestProductService_ListProducts_RepoError(t *testing.T) {
	ctx := context.Background()

	repoErr := errors.New("db error")
	repo := &stubProductRepo{err: repoErr}
	svc := service.NewProductService(repo)

	_, err := svc.ListProducts(ctx)
	if err == nil {
		t.Fatalf("ListProducts() error = nil, want non-nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("ListProducts() error = %v, want %v", err, repoErr)
	}
}
