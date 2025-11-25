package shared

import (
	"context"

	"github.com/rs/zerolog"
)

type ctxLoggerKey struct{}

// loggerKey is a unique, unexported key used to store the logger in context.
var loggerKey ctxLoggerKey

// WithLogger returns a new context that carries the given logger.
func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFrom returns the logger stored in ctx, or fallback if none is found.
func LoggerFrom(ctx context.Context, fallback zerolog.Logger) zerolog.Logger {
	if l, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return l
	}
	return fallback
}
