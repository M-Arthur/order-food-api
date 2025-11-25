package httpapi

import (
	"net/http"

	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/httpapi/handlers"
	"github.com/M-Arthur/order-food-api/internal/httpapi/middleware"
	"github.com/M-Arthur/order-food-api/internal/service"
	"github.com/M-Arthur/order-food-api/internal/storage"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

// RouterConfig centralises configration for the router
type RouterConfig struct {
	Logger zerolog.Logger
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
	registerHealthRoutes(r)
	registerProductRoutes(r, cfg.Logger)

	return r
}

// registerHealthRoutes sets up health check endpoints
func registerHealthRoutes(r chi.Router) {
	r.Get("/health", handlers.Health)
}

func registerProductRoutes(r chi.Router, l zerolog.Logger) {
	seedProducts := []domain.Product{
		{ID: domain.ProductID("10"), Name: "Chicken Waffle", Price: domain.NewMoneyFromFloat(12.5), Category: "Waffle"},
		{ID: domain.ProductID("11"), Name: "Fries", Price: domain.NewMoneyFromFloat(5.5), Category: "Sides"},
	}
	productRepo := storage.NewInMemoryProductRepository(seedProducts)
	productSvc := service.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productSvc, l)

	r.Get("/product", productHandler.ListProducts)
}
