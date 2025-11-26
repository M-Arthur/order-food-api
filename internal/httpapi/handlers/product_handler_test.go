package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/api"
	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/httpapi/handlers"
	"github.com/M-Arthur/order-food-api/internal/service"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

// stubProductService implemnets service.ProductService for tests
type stubProductService struct {
	products map[domain.ProductID]domain.Product
	err      error
}

func newStubProductService(seed []domain.Product, err error) *stubProductService {
	products := make(map[domain.ProductID]domain.Product)
	for _, p := range seed {
		products[p.ID] = p
	}
	return &stubProductService{
		products: products,
		err:      err,
	}
}

func (s *stubProductService) ListProducts(_ context.Context) ([]domain.Product, error) {
	out := make([]domain.Product, 0, len(s.products))
	for _, p := range s.products {
		out = append(out, p)
	}
	return out, s.err
}

func (s *stubProductService) GetProduct(_ context.Context, id domain.ProductID) (*domain.Product, error) {
	p, ok := s.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return &p, s.err
}

// Ensure stub implements the interface at complie time
var _ service.ProductService = (*stubProductService)(nil)

func TestProductHandler_ListProducts_Success(t *testing.T) {
	seed := []domain.Product{
		{ID: domain.ProductID("10"), Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		{ID: domain.ProductID("11"), Name: "Fries", Price: domain.NewMoneyFromFloat(5.5), Category: "Sides"},
	}

	svc := newStubProductService(seed, nil)
	h := handlers.NewProductHandler(svc)

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

	svc := newStubProductService(seed, errors.New("test error"))
	h := handlers.NewProductHandler(svc)

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

func TestProductHandler_GetProduct_Success(t *testing.T) {
	seed := []domain.Product{
		{ID: domain.ProductID("10"), Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		{ID: domain.ProductID("11"), Name: "Fries", Price: domain.NewMoneyFromFloat(5.5), Category: "Sides"},
	}

	svc := newStubProductService(seed, nil)
	h := handlers.NewProductHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/product/10", nil)
	rr := httptest.NewRecorder()

	// Simulate chi param extraction
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("productId", "10")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetProductByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var got api.ProductDTO
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if got.ID != "10" {
		t.Fatalf("got.ID = %s, want 10", got.ID)
	}
}

func TestProductHandler_GetProduct_Error(t *testing.T) {
	seed := []domain.Product{
		{ID: domain.ProductID("10"), Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
	}

	svc := newStubProductService(seed, nil)
	h := handlers.NewProductHandler(svc)

	tests := []struct {
		name string
		id   string
		code int
	}{
		{name: "invalid id format", id: "abc", code: http.StatusBadRequest},
		{name: "not found", id: "999", code: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/product/"+tt.id, nil)
			rr := httptest.NewRecorder()

			// Simulate chi param extraction
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("productId", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.GetProductByID(rr, req)

			if rr.Code != tt.code {
				t.Fatalf("status = %d, want %d", rr.Code, tt.code)
			}
		})
	}
}
