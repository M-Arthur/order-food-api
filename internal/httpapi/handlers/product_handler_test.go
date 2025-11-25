package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/api"
	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/httpapi/handlers"
	"github.com/M-Arthur/order-food-api/internal/service"
	"github.com/rs/zerolog"
)

// stubProductService implemnets service.ProductService for tests
type stubProductService struct {
	products []domain.Product
	err      error
}

func (s *stubProductService) ListProducts(_ context.Context) ([]domain.Product, error) {
	return s.products, s.err
}

// Ensure stub implements the interface at complie time
var _ service.ProductService = (*stubProductService)(nil)

func TestProductHandler_ListProducts_Success(t *testing.T) {
	seed := []domain.Product{
		{ID: domain.ProductID("10"), Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		{ID: domain.ProductID("11"), Name: "Fries", Price: domain.NewMoneyFromFloat(5.5), Category: "Sides"},
	}
	baseLogger := zerolog.New(io.Discard).With().Timestamp().Logger()

	svc := &stubProductService{products: seed}
	h := handlers.NewProductHandler(svc, baseLogger)

	req := httptest.NewRequest(http.MethodGet, "/product", nil)
	rr := httptest.NewRecorder()

	h.ListProducts(rr, req)

	res := rr.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}

	ct := res.Header.Get("Content-Type")
	if ct == "" || ct[:16] != "application/json" {
		t.Fatalf("Content-Type =%q, want application/json", ct)
	}

	var got []api.ProductDTO
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(got) != len(seed) {
		t.Fatalf("len(products) = %d, want %d", len(got), len(seed))
	}

	for i, dto := range got {
		want := seed[i]
		if dto.ID != string(want.ID) {
			t.Errorf("products[%d].ID = %s, want %s", i, dto.ID, want.ID)
		}
		if dto.Name != want.Name {
			t.Errorf("products[%d].Name = %s, want %s", i, dto.Name, want.Name)
		}
		if dto.Category != want.Category {
			t.Errorf("products[%d].Category = %s, want %s", i, dto.Category, want.Category)
		}
		if dto.Price != want.Price.ToFloat() {
			t.Errorf("products[%d].Price = %v, want %v", i, dto.Price, want.Price.ToFloat())
		}
	}
}

func TestProductHandler_ListProducts_Error(t *testing.T) {
	seed := []domain.Product{
		{ID: domain.ProductID("10"), Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
	}
	var buf bytes.Buffer
	baseLogger := zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &baseLogger

	svc := &stubProductService{products: seed, err: errors.New("test error")}
	h := handlers.NewProductHandler(svc, baseLogger)

	req := httptest.NewRequest(http.MethodGet, "/product", nil)
	rr := httptest.NewRecorder()

	h.ListProducts(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "failed to list products") {
		t.Fatalf("expected 'failed to list products' in log output, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "test error") {
		t.Fatalf("expected 'test error' in log output, got: %s", logOutput)
	}
}
