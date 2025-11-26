package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestLoggerMiddleware_AddsLoggerToContext(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := zerolog.New(&buf).With().Timestamp().Logger()

	var (
		called    bool
		ctxLogger zerolog.Logger
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		ctxLogger = *zerolog.Ctx(r.Context())
		ctxLogger.Info().Msg("test message")
	})

	mw := LoggerMiddleware(baseLogger)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	mw(next).ServeHTTP(rr, req)

	if !called {
		t.Fatalf("next handler was not called")
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "test message") {
		t.Fatalf("expected test message details in log output, got: %s", logOutput)
	}
}
