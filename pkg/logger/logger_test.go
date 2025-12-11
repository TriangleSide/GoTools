package logger_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestFromContext_EmptyContext_ReturnsNewLogger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
}

func TestFromContext_CalledTwice_ReturnsSameLogger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, log1 := logger.FromContext(ctx)
	_, log2 := logger.FromContext(ctx)
	assert.Equals(t, log1, log2)
}

func TestFromContext_WithLogLevelError_ReturnsLoggerWithErrorLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "ERROR")
	ctx := context.Background()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelError))
	assert.False(t, log.Enabled(ctx, slog.LevelWarn))
}

func TestFromContext_WithLogLevelWarn_ReturnsLoggerWithWarnLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "WARN")
	ctx := context.Background()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelWarn))
	assert.False(t, log.Enabled(ctx, slog.LevelInfo))
}

func TestFromContext_WithLogLevelInfo_ReturnsLoggerWithInfoLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	ctx := context.Background()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelInfo))
	assert.False(t, log.Enabled(ctx, slog.LevelDebug))
}

func TestFromContext_WithLogLevelDebug_ReturnsLoggerWithDebugLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "DEBUG")
	ctx := context.Background()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelDebug))
}

func TestFromContext_WithLogLevelLowercase_ReturnsLoggerWithCorrectLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "error")
	ctx := context.Background()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelError))
	assert.False(t, log.Enabled(ctx, slog.LevelWarn))
}

func TestFromContext_WithInvalidLogLevel_ReturnsLoggerWithInfoLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INVALID")
	ctx := context.Background()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelInfo))
	assert.False(t, log.Enabled(ctx, slog.LevelDebug))
}

func TestWithAttrs_ExistingLogger_AddsAttrsToExistingLogger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx, originalLog := logger.FromContext(ctx)
	attr := slog.String("key", "value")
	resultCtx, newLog := logger.WithAttrs(ctx, attr)
	assert.NotNil(t, resultCtx)
	assert.NotEquals(t, originalLog, newLog)
}

func TestWithAttrs_NoAttrs_ReturnsLoggerWithNoAdditionalAttrs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	resultCtx, log := logger.WithAttrs(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
}

func TestFromContext_ConcurrentAccess_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100

	ctx := context.Background()
	ctx, _ = logger.FromContext(ctx)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				_, log := logger.FromContext(ctx)
				assert.NotNil(t, log, assert.Continue())
			}
		}()
	}

	wg.Wait()
}

func TestWithAttrs_ConcurrentAccess_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100

	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				localCtx, log := logger.WithAttrs(ctx, slog.Int("goroutine", id), slog.Int("iteration", j))
				assert.NotNil(t, localCtx, assert.Continue())
				assert.NotNil(t, log, assert.Continue())
			}
		}(i)
	}

	wg.Wait()
}
