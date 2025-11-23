package server_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/M-Arthur/kart-challenge/internal/server"
	"github.com/go-chi/chi"
)

func findFreePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	defer func() {
		_ = l.Close()
	}()
	return l.Addr().String()
}

func TestServer_GracefulShutdown(t *testing.T) {
	// Arrange: router with /health
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	addr := findFreePort(t)
	srv := server.New(addr, r)

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Timeout: 2 * time.Second}

	// 1) Server responds before shutdown
	resp, err := client.Get("http://" + addr + "/health")
	if err != nil {
		t.Fatalf("expected /health to be reachable, got error: %v", err)
	}

	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	// 2) Trigger graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		t.Fatalf("expected graceful shutdown without error, got: %v", err)
	}

	// 3) Ensrue Start() returned (either nil or ErrServerClosed)
	select {
	case err := <-errCh:
		// In a normal graceful shutdown, http.Server.ListenAndServe
		// returns http.ErrServerClosed, which we treat as "ok".
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("expected server to shut down without error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("server did not exit after shutdown")
	}

	// 4) After shutdown, new requests should fail
	_, err = client.Get("http://" + addr + "/health")
	if err == nil {
		t.Fatalf("expected /health to be unreachable after shutdown, but request succeeded")
	}
}
