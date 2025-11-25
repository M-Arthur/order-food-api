package shared_test

import (
	"context"
	"io"
	"testing"

	"github.com/M-Arthur/order-food-api/internal/httpapi/shared"
	"github.com/rs/zerolog"
)

func TestWithLoggerAndLoggerFrom(t *testing.T) {
	base := zerolog.Nop()
	ctx := context.Background()

	// When no logger in context, it should return fallback
	got := shared.LoggerFrom(ctx, base)
	if got.GetLevel() != base.GetLevel() {
		t.Fatalf("expected fallback logger when none in context")
	}

	// Put a logger into context
	custom := zerolog.New(io.Discard).With().Str("component", "test").Logger()
	ctxWithLogger := shared.WithLogger(ctx, custom)

	got2 := shared.LoggerFrom(ctxWithLogger, base)
	// This is a value type, so you  can compare struct fields or just ensure it's not fallback
	if got2.GetLevel() != custom.GetLevel() {
		t.Fatalf("expected logger from context")
	}
}
