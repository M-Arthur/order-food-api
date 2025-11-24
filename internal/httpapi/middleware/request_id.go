package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxRequestIDKey struct{}

var requestIDKey ctxRequestIDKey

// WithRequestID attaches a request ID to the context
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFrom retrieves the request ID from the context, if present.
func RequestIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

// RequestID middleware generates a requuest ID for each request, stores it in
// the context, and wrtes it to the X-Request-ID response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()

		// Set header so clients can see/correlate it
		w.Header().Set("X-Request-ID", id)

		ctx := WithRequestID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
