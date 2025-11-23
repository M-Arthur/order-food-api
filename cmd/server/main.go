package main

import (
	"log"
	"net/http"

	"github.com/M-Arthur/kart-challenge/internal/config"
	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()

	// Temporary health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	cfg := config.Load()
	port := cfg.Port

	log.Printf("Server starrting on :%s...\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
