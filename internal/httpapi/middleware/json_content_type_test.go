package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONContentTypeMiddleware_SetsDefaultHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler does not set Content-Type
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	mw := JSONContentType(handler)
	mw.ServeHTTP(rr, req)

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected Content-Type to be application/json, got %s", got)
	}
}

func TestJSONContentTypeMiddleware_DoesNotOverrideExistingHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	mw := JSONContentType(handler)
	mw.ServeHTTP(rr, req)

	if got := rr.Header().Get("Content-Type"); got != "text/plain" {
		t.Errorf("expected Content-Type to remain 'text/plain', got '%s'", got)
	}
}
