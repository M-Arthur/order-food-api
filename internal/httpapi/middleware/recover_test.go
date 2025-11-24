package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestRecoverMiddleware_CatchesPanic(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := zerolog.New(&buf).With().Timestamp().Logger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	mw := Recover(baseLogger)(handler)
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	body := rr.Body.String()
	if body == "" {
		t.Fatalf("expected non-empty body")
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got '%s'", rr.Header().Get("Content-Type"))
	}

	logOutput := buf.String()
	println(logOutput)
	if !strings.Contains(logOutput, "panic") {
		t.Fatalf("expected panic details in log output, got: %s", logOutput)
	}
}
