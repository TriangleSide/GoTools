package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/TriangleSide/GoTools/pkg/config"
)

// ctxKey is the type used for storing the logger in context.
type ctxKey struct{}

// ctxKeyInstance is the context key for the logger.
var ctxKeyInstance ctxKey

// getLogLevelFromEnv retrieves the log level from environment variables.
func getLogLevelFromEnv() slog.Level {
	type Config struct {
		LogLevel string `config:"ENV"`
	}

	cfg, err := config.Process[Config]()
	if err != nil {
		return slog.LevelInfo
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return slog.LevelInfo
	}
	return level
}

// newLogger creates a new slog.Logger with a JSON handler writing to stdout.
func newLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevelFromEnv(),
	})
	return slog.New(handler)
}

// FromContext retrieves the logger from the context.
// If no logger is found, it returns a new default logger.
func FromContext(ctx context.Context) (context.Context, *slog.Logger) {
	if logger, ok := ctx.Value(ctxKeyInstance).(*slog.Logger); ok {
		return ctx, logger
	}
	logger := newLogger()
	ctx = context.WithValue(ctx, ctxKeyInstance, logger)
	return ctx, logger
}

// WithAttrs returns a new context with a logger that has the given attributes.
// If no logger exists in the context, a new one is created.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) (context.Context, *slog.Logger) {
	ctx, logger := FromContext(ctx)
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	loggerWithAttrs := logger.With(args...)
	return context.WithValue(ctx, ctxKeyInstance, loggerWithAttrs), loggerWithAttrs
}
