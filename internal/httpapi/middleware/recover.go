package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/M-Arthur/order-food-api/internal/httpapi/shared"
	"github.com/rs/zerolog"
)

// Recover creates a middlware that recovers from panics in handlers,
// logs the panic + stack trace, and returns a 500 error
func Recover(baseLogger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger := shared.LoggerFrom(r.Context(), baseLogger)
					logger.Error().
						Interface("panic", rec).
						Bytes("stack", debug.Stack()).
						Msg("panic recovered")

					// Best effort to return 500 response; if write fails, nothing else we can do.
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal_server_error"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
