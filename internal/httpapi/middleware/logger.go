package middleware

import (
	"net/http"

	"github.com/rs/zerolog"
)

func LoggerMiddleware(base zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := base.WithContext(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
