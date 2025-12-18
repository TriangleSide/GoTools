package logger_test

import (
	"log/slog"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestFromContext_EmptyContext_ReturnsNewLogger(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
}

func TestFromContext_CalledTwice_ReturnsSameLogger(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, log1 := logger.FromContext(ctx)
	_, log2 := logger.FromContext(ctx)
	assert.Equals(t, log1, log2)
}

func TestFromContext_WithLogLevelError_ReturnsLoggerWithErrorLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "ERROR")
	ctx := t.Context()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelError))
	assert.False(t, log.Enabled(ctx, slog.LevelWarn))
}

func TestFromContext_WithLogLevelWarn_ReturnsLoggerWithWarnLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "WARN")
	ctx := t.Context()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelWarn))
	assert.False(t, log.Enabled(ctx, slog.LevelInfo))
}

func TestFromContext_WithLogLevelInfo_ReturnsLoggerWithInfoLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INFO")
	ctx := t.Context()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelInfo))
	assert.False(t, log.Enabled(ctx, slog.LevelDebug))
}

func TestFromContext_WithLogLevelDebug_ReturnsLoggerWithDebugLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "DEBUG")
	ctx := t.Context()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelDebug))
}

func TestFromContext_WithLogLevelLowercase_ReturnsLoggerWithCorrectLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "error")
	ctx := t.Context()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelError))
	assert.False(t, log.Enabled(ctx, slog.LevelWarn))
}

func TestFromContext_WithInvalidLogLevel_ReturnsLoggerWithInfoLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "INVALID")
	ctx := t.Context()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelInfo))
	assert.False(t, log.Enabled(ctx, slog.LevelDebug))
}

func TestFromContext_WithNoLogLevel_ReturnsLoggerWithInfoLevel(t *testing.T) {
	t.Setenv("NOT_LOG_LEVEL", "SOMETHING_ELSE")
	ctx := t.Context()
	resultCtx, log := logger.FromContext(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
	assert.True(t, log.Enabled(ctx, slog.LevelInfo))
	assert.False(t, log.Enabled(ctx, slog.LevelDebug))
}

func TestWithAttrs_ExistingLogger_AddsAttrsToExistingLogger(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctx, originalLog := logger.FromContext(ctx)
	attr := slog.String("key", "value")
	resultCtx, newLog := logger.WithAttrs(ctx, attr)
	assert.NotNil(t, resultCtx)
	assert.NotEquals(t, originalLog, newLog)
}

func TestWithAttrs_NoAttrs_ReturnsLoggerWithNoAdditionalAttrs(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	resultCtx, log := logger.WithAttrs(ctx)
	assert.NotNil(t, resultCtx)
	assert.NotNil(t, log)
}

func TestFromContext_ConcurrentAccess_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100

	ctx := t.Context()
	ctx, _ = logger.FromContext(ctx)

	var waitGroup sync.WaitGroup

	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				_, log := logger.FromContext(ctx)
				assert.NotNil(t, log, assert.Continue())
			}
		})
	}

	waitGroup.Wait()
}

func TestWithAttrs_ConcurrentAccess_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100

	ctx := t.Context()

	var waitGroup sync.WaitGroup

	for goroutineIdx := range goroutines {
		waitGroup.Go(func() {
			for j := range iterations {
				localCtx, log := logger.WithAttrs(ctx, slog.Int("goroutine", goroutineIdx), slog.Int("iteration", j))
				assert.NotNil(t, localCtx, assert.Continue())
				assert.NotNil(t, log, assert.Continue())
			}
		})
	}

	waitGroup.Wait()
}
