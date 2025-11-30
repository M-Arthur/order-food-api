package middleware

import (
	"net/http"

	"github.com/M-Arthur/order-food-api/internal/httpapi/shared"
)

// APIKeyAuth returns a middleware that enforces the expected API key
// in the `api_key` header, as defined in the OpenAPI spec.
func APIKeyAuth(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("api_key")
			if apiKey == "" || apiKey != expectedKey {
				shared.WriteJSONError(w, r, http.StatusUnauthorized, "invalid or missing API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
