package logger_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type testWriter struct{}

func (w *testWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestLogger(t *testing.T) {
	t.Run("when setting output it should write to specified writer", func(t *testing.T) {
		var output bytes.Buffer
		logger.SetOutput(&output)
		t.Cleanup(func() {
			logger.SetOutput(os.Stdout)
		})

		logger.Error(context.Background(), "test message")

		assert.Contains(t, output.String(), "test message")
	})

	t.Run("when logging at panic level it should panic with messages", func(t *testing.T) {
		assert.PanicPart(t, func() {
			logger.Panic("Panic Message")
		}, "Panic Message")
		assert.PanicPart(t, func() {
			logger.Panicf("Panic %s", "Message F")
		}, "Panic Message F")
		assert.PanicPart(t, func() {
			logger.PanicFn(func() []any {
				return []any{"Panic Message Fn"}
			})
		}, "Panic Message Fn")
	})

	t.Run("when logging at different levels it should log appropriate messages", func(t *testing.T) {
		setAndRecordOutput := func() *bytes.Buffer {
			var output bytes.Buffer
			logger.SetOutput(&output)
			logger.SetFormatter(func(fields map[string]any, msg string) string {
				return msg
			})
			return &output
		}

		t.Cleanup(func() {
			logger.SetOutput(os.Stdout)
			logger.SetLevel(logger.LevelInfo)
			logger.SetFormatter(logger.DefaultFormatter)
		})

		testCases := []struct {
			name     string
			level    logger.LogLevel
			expected string
		}{
			{name: "error level", level: logger.LevelError, expected: "EEE"},
			{name: "warn level", level: logger.LevelWarn, expected: "EEEWWW"},
			{name: "info level", level: logger.LevelInfo, expected: "EEEWWWIII"},
			{name: "debug level", level: logger.LevelDebug, expected: "EEEWWWIIIDDD"},
			{name: "trace level", level: logger.LevelTrace, expected: "EEEWWWIIIDDDTTT"},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("when log level is %s it should write expected messages", tc.level), func(t *testing.T) {
				logger.SetLevel(tc.level)
				assert.Equals(t, logger.GetLevel(), tc.level)

				output := setAndRecordOutput()
				assert.Equals(t, len(output.Bytes()), 0)

				logger.Error("E")
				logger.Errorf("E")
				logger.ErrorFn(func() []any { return []any{"E"} })
				logger.Warn("W")
				logger.Warnf("W")
				logger.WarnFn(func() []any { return []any{"W"} })
				logger.Info("I")
				logger.Infof("I")
				logger.InfoFn(func() []any { return []any{"I"} })
				logger.Debug("D")
				logger.Debugf("D")
				logger.DebugFn(func() []any { return []any{"D"} })
				logger.Trace("T")
				logger.Tracef("T")
				logger.TraceFn(func() []any { return []any{"T"} })

				assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), tc.expected)
			})
		}
	})

	t.Run("when logging concurrently it should handle output changes", func(t *testing.T) {
		t.Cleanup(func() {
			logger.SetOutput(os.Stdout)
			logger.SetLevel(logger.LevelInfo)
			logger.SetFormatter(logger.DefaultFormatter)
		})

		const threadCount = 4
		const opsPerThread = 1000

		wg := &sync.WaitGroup{}
		waitChan := make(chan struct{})

		for range threadCount {
			wg.Go(func() {
				<-waitChan
				for range opsPerThread {
					logger.SetOutput(&testWriter{})
					logger.SetFormatter(logger.DefaultFormatter)
					for _, level := range []logger.LogLevel{logger.LevelError, logger.LevelWarn, logger.LevelInfo, logger.LevelDebug, logger.LevelTrace} {
						logger.SetLevel(level)
						_, logEntry := logger.AddField(context.Background(), "test", "test")
						logEntry.Error("test")
						logEntry.Errorf("test %s", "test")
						logEntry.ErrorFn(func() []any { return []any{"test"} })
						logEntry.Warn("test")
						logEntry.Warnf("test %s", "test")
						logEntry.WarnFn(func() []any { return []any{"test"} })
						logEntry.Info("test")
						logEntry.Infof("test %s", "test")
						logEntry.InfoFn(func() []any { return []any{"test"} })
						logEntry.Debug("test")
						logEntry.Debugf("test %s", "test")
						logEntry.DebugFn(func() []any { return []any{"test"} })
						logEntry.Trace("test")
						logEntry.Tracef("test %s", "test")
						logEntry.TraceFn(func() []any { return []any{"test"} })
					}
				}
			})
		}

		close(waitChan)
		wg.Wait()
	})

	t.Run("when logging while formatters change it should log all messages", func(t *testing.T) {
		t.Cleanup(func() {
			logger.SetOutput(os.Stdout)
			logger.SetLevel(logger.LevelInfo)
			logger.SetFormatter(logger.DefaultFormatter)
		})

		const iterations = 5000

		output := &bytes.Buffer{}
		logger.SetOutput(output)
		logger.SetLevel(logger.LevelTrace)

		formatterDone := make(chan struct{})
		go func() {
			for i := range iterations/6 {
				logger.SetFormatter(func(fields map[string]any, msg string) string {
					return fmt.Sprintf("%d %s", i, msg)
				})
			}
			close(formatterDone)
		}()

		for range iterations {
			assert.PanicPart(t, func() {
				logger.Panic("P")
			}, "P")
			logger.Error("E")
			logger.Warn("W")
			logger.Info("I")
			logger.Debug("D")
			logger.Trace("T")
		}

		<-formatterDone
		assert.Equals(t, strings.Count(output.String(), "P"), iterations)
		assert.Equals(t, strings.Count(output.String(), "E"), iterations)
		assert.Equals(t, strings.Count(output.String(), "W"), iterations)
		assert.Equals(t, strings.Count(output.String(), "I"), iterations)
		assert.Equals(t, strings.Count(output.String(), "D"), iterations)
		assert.Equals(t, strings.Count(output.String(), "T"), iterations)
	})
}
