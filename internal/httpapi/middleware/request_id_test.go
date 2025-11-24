package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware_SetHeaderAndContext(t *testing.T) {
	var (
		capturedID        string
		capturedFromCtx   string
		capturedFromCtxOk bool
	)

	// Handler to inspect context and header
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := w.Header().Get("X-Request-ID")
		if id == "" {
			t.Fatalf("expected X-Request-ID header to be set")
		}
		capturedID = id

		if got, ok := RequestIDFrom(r.Context()); !ok {
			t.Fatalf("expected request ID in context")
		} else {
			capturedFromCtx = got
			capturedFromCtxOk = ok
		}

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	mw := RequestID(handler)
	mw.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if !capturedFromCtxOk {
		t.Fatalf("expected request ID in context")
	}
	if capturedID == "" || capturedFromCtx == "" {
		t.Fatalf("expected non-empty request IDs")
	}
	if capturedID != capturedFromCtx {
		t.Fatalf("heder and context request IDs differ: header=%s ctx=%s", capturedID, capturedFromCtx)
	}
}

func TestWithRequestIDAndRequestIDFrom(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-id")

	id, ok := RequestIDFrom(ctx)
	if !ok {
		t.Fatalf("expected request ID in context")
	}
	if id != "test-id" {
		t.Fatalf("expected request ID 'test-id', got '%s'", id)
	}
}
