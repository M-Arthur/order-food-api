package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/M-Arthur/order-food-api/internal/httpapi/middleware"
	"github.com/M-Arthur/order-food-api/internal/httpapi/shared"
)

func TestAPIKeyAuth_ValidKey_AllowsRequest(t *testing.T) {
	const expectedKey = "apitest"

	r := chi.NewRouter()
	// chi RequestID so we can inspect it if needed
	r.Use(chimiddleware.RequestID)
	r.With(middleware.APIKeyAuth(expectedKey)).Post("/order", func(w http.ResponseWriter, r *http.Request) {
		shared.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/order", nil)
	req.Header.Set("api_key", expectedKey)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	res := rr.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
}

func TestAPIKeyAuth_MissingOrInvalidKey_Returns401(t *testing.T) {
	const expectedKey = "apitest"

	tests := []struct {
		name      string
		headerKey string
	}{
		{name: "missing", headerKey: ""},
		{name: "invalid", headerKey: "wrong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(chimiddleware.RequestID)
			r.With(middleware.APIKeyAuth(expectedKey)).Post("/order", func(w http.ResponseWriter, r *http.Request) {
				shared.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
			})

			req := httptest.NewRequest(http.MethodPost, "/order", nil)
			if tt.headerKey != "" {
				req.Header.Set("api_key", tt.headerKey)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			res := rr.Result()
			defer func() {
				_ = res.Body.Close()
			}()

			if res.StatusCode != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusUnauthorized)
			}
		})
	}
}
