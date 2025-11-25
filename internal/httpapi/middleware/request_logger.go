package middleware

import (
	"net/http"
	"time"

	"github.com/M-Arthur/order-food-api/internal/httpapi/shared"
	"github.com/rs/zerolog"
)

// responseWRiter wraps http.ResponseWriter to capture status code & bytes written.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// RequestLogger creates a middleware that:
//  1. Derives a reqeust-scoped logger from base logger
//  2. Stores it in context via WithLogger
//  3. Logs the request after it completes
func RequestLogger(baseLogger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Start from base logger or any logger already in context
			l := shared.LoggerFrom(r.Context(), baseLogger).With().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent())

			if id, ok := RequestIDFrom(r.Context()); ok {
				l = l.Str("request_id", id)
			}
			reqLogger := l.Logger()
			// Store in context for handlers to use
			ctx := shared.WithLogger(r.Context(), reqLogger)

			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(rw, r.WithContext(ctx))

			reqLogger.Info().
				Int("status", rw.status).
				Int("bytes", rw.bytes).
				Dur("duration", time.Since(start)).
				Msg("http_request")
		})
	}
}
