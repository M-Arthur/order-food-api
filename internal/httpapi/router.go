package httpapi

import (
	"net/http"

	"github.com/M-Arthur/order-food-api/internal/bootstrap"
	"github.com/M-Arthur/order-food-api/internal/httpapi/handlers"
	"github.com/M-Arthur/order-food-api/internal/httpapi/middleware"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

// RouterConfig centralises configration for the router
type RouterConfig struct {
	Logger zerolog.Logger
	Deps   *bootstrap.Dependencies
}

// NewRouter builds the HTTP router with middleware and routes
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// --- Global middlewares ---
	r.Use(
		middleware.Recover,
		middleware.JSONContentType,
		middleware.RequestID,
		middleware.RequestLogger,
	)

	// --- Route groups / endpoints ---
	r.Get("/health", handlers.Health)
	r.Get("/product", cfg.Deps.Handlers.Product.ListProducts)
	r.Get("/product/{productId}", cfg.Deps.Handlers.Product.GetProductByID)

	return r
}
