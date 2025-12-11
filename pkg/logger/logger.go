package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/TriangleSide/GoTools/pkg/config"
)

// ctxKey is the type used for storing the logger in context.
type ctxKey struct{}

// ctxKeyInstance is the context key for the logger.
var ctxKeyInstance = ctxKey{}

// getLogLevelFromEnv retrieves the log level from environment variables.
func getLogLevelFromEnv() slog.Level {
	type Config struct {
		LogLevel string `config:"ENV" config_default:"INFO" validate:"required,oneof=ERROR WARN INFO DEBUG error warn info debug"`
	}
	cfg, err := config.ProcessAndValidate[Config]()
	if err != nil {
		slog.Error(err.Error())
		return slog.LevelInfo
	}
	strLevel := strings.ToLower(cfg.LogLevel)
	var level = slog.LevelInfo
	switch strLevel {
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	}
	return level
}

// new creates a new slog.Logger with a JSON handler writing to stdout.
func newLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevelFromEnv(),
	})
	return slog.New(handler)
}

// FromContext retrieves the logger from the context.
// If no logger is found, it returns a new default logger.
func FromContext(ctx context.Context) (context.Context, *slog.Logger) {
	loggerUncast := ctx.Value(ctxKeyInstance)
	if loggerUncast != nil {
		logger := loggerUncast.(*slog.Logger)
		return ctx, logger
	}
	logger := newLogger()
	ctx = context.WithValue(ctx, ctxKeyInstance, logger)
	return ctx, logger
}

// WithAttrs returns a new context with a logger that has the given attributes.
// If no logger exists in the context, a new one is created.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	ctx, logger := FromContext(ctx)
	anySlice := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		anySlice = append(anySlice, attr)
	}
	return context.WithValue(ctx, ctxKeyInstance, logger.With(anySlice...))
}
