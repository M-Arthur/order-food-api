package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/M-Arthur/order-food-api/internal/bootstrap"
	"github.com/M-Arthur/order-food-api/internal/config"
	"github.com/M-Arthur/order-food-api/internal/httpapi"
	"github.com/M-Arthur/order-food-api/internal/logger"
	"github.com/M-Arthur/order-food-api/internal/server"
)

func main() {
	// 1) Init logger
	cfg := config.Load()
	appLogger := logger.New(cfg.AppEnv)
	appLogger.Info().Msg("initiating server")

	deps, err := bootstrap.BuildDependencies(cfg)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("loading dependencies unsuccessfully")
	}

	appLogger.Info().Msg("starting server")

	// 2) Build router with configured routes & middleware
	r := httpapi.NewRouter(httpapi.RouterConfig{
		Logger: appLogger,
		Deps:   deps,
	})

	// 3) Server config
	addr := ":" + cfg.Port
	srv := server.New(addr, r)

	// 4) Graceful shutdown wiring
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		appLogger.Info().Str("addr", addr).Msg("server listening")
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Wait for signal
	<-stop
	appLogger.Info().Msg("shutdown signal received, shutting down server")
	// Graceful shutdown context (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal().Err(err).Msg("server forced to shutdown")
	}

	appLogger.Info().Msg("server exited gracefully")
}
