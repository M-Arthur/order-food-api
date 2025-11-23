package httpapi

import (
	"net/http"

	handlers "github.com/M-Arthur/kart-challenge/internal/httpapi/Handlers"
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

	// --- Route groups / endpoints ---
	registerHealthRoutes(r)

	return r
}

// registerHealthRoutes sets up health check endpoints
func registerHealthRoutes(r chi.Router) {
	r.Get("/health", handlers.Health)
}
