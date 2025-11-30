package httpapi

import (
	"net/http"

	"github.com/M-Arthur/order-food-api/internal/bootstrap"
	"github.com/M-Arthur/order-food-api/internal/httpapi/handlers"
	"github.com/M-Arthur/order-food-api/internal/httpapi/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// RouterConfig centralises configration for the router
type RouterConfig struct {
	Logger zerolog.Logger
	Deps   *bootstrap.Dependencies
	APIKey string
}

// NewRouter builds the HTTP router with middleware and routes
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// --- Global middlewares ---
	r.Use(
		chimiddleware.StripSlashes,
		chimiddleware.SupressNotFound(r),
		chimiddleware.RequestID,
		chimiddleware.RealIP,
		middleware.LoggerMiddleware(cfg.Logger),
		middleware.Recover,
		middleware.JSONContentType,
		middleware.RequestLogger,
	)

	// --- Route groups / endpoints ---
	r.Get("/health", handlers.Health)

	r.Route("/api", func(api chi.Router) {
		api.Get("/product", cfg.Deps.Handlers.Product.ListProducts)
		api.Get("/product/{productId}", cfg.Deps.Handlers.Product.GetProductByID)
		api.With(middleware.APIKeyAuth(cfg.APIKey)).Post("/order", cfg.Deps.Handlers.Order.PlaceOrder)
	})

	return r
}
