package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/rs/zerolog"
)

type logEntry map[string]any

type jsonCapturingWriter struct {
	mu      sync.Mutex
	entries []logEntry
}

func newJSONCapturingWriter() *jsonCapturingWriter {
	return &jsonCapturingWriter{
		entries: make([]logEntry, 0),
	}
}

func (w *jsonCapturingWriter) Write(p []byte) (int, error) {
	var e logEntry

	if err := json.Unmarshal(p, &e); err != nil {
		return 0, err
	}

	w.mu.Lock()
	w.entries = append(w.entries, e)
	w.mu.Unlock()

	return len(p), nil
}

func TestRequestLogger_LogRequest(t *testing.T) {
	writer := newJSONCapturingWriter()
	baseLogger := zerolog.New(writer).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &baseLogger

	handlerCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte(`{"ok":true}`))
		if err != nil {
			t.Fatalf("failed to write response body: %v", err)
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rr := httptest.NewRecorder()

	mw := RequestLogger(handler)
	mw.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Fatalf("expected handler to be called")
	}

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected satus 201, got %d", rr.Code)
	}

	if len(writer.entries) == 0 {
		t.Fatalf("expected log output, got empty")
	}

	entries := writer.entries[0]
	checks := map[string]any{
		"path":   "/test",
		"method": "POST",
		"status": float64(201),
	}

	for k, want := range checks {
		if got := entries[k]; got != want {
			t.Fatalf("expected %s '%v(type: %T)', got '%v(type: %T)'", k, want, want, got, got)
		}
	}
}
