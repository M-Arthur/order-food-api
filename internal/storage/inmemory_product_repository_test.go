package storage_test

import (
	"context"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/storage"
)

func TestInMemoryProductRepository_GetProductByID(t *testing.T) {
	repo := storage.NewInMemoryProductRepository([]domain.Product{
		{ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
	})

	tests := []struct {
		name    string
		id      domain.ProductID
		wantErr bool
	}{
		{name: "found", id: "10", wantErr: false},
		{name: "not found", id: "999", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := repo.GetProductByID(context.Background(), tt.id)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if p != nil {
					t.Fatalf("expected nil product, got %+v", p)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p == nil {
				t.Fatalf("expected product, got nill")
			}
			if p.ID != "10" {
				t.Fatalf("expected ID=%s, got %s", "10", p.ID)
			}
		})
	}
}

func TestInMemoryProductRepository_ListProducts(t *testing.T) {
	repo := storage.NewInMemoryProductRepository([]domain.Product{
		{ID: "10", Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		{ID: "11", Name: "Vegan Waffle", Price: domain.NewMoneyFromFloat(11.0), Category: "Waffle"},
	})

	list, err := repo.ListProducts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(list); got != 2 {
		t.Fatalf("expected 2 products, got %d", got)
	}
}
