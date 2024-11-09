package logger_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
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

func TestSetOutput(t *testing.T) {
	var output bytes.Buffer
	logger.SetOutput(&output)
	t.Cleanup(func() {
		logger.SetOutput(os.Stdout)
	})
	logger.Error(context.Background(), "test message")
	assert.Contains(t, output.String(), "test message")
}

func TestLogger(t *testing.T) {
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
	})

	t.Run("when logging at Panic level it should panic with the messages", func(t *testing.T) {
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
		testCases := []struct {
			level    logger.LogLevel
			expected string
		}{
			{logger.LevelError, "EEE"},
			{logger.LevelWarn, "EEEWWW"},
			{logger.LevelInfo, "EEEWWWIII"},
			{logger.LevelDebug, "EEEWWWIIIDDD"},
			{logger.LevelTrace, "EEEWWWIIIDDDTTT"},
		}
		for _, tc := range testCases {
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
		}
	})
}

func testFatalScenario(t *testing.T, uniqueEnvName string, testName string, fatalCallback func()) {
	t.Helper()
	t.Parallel()

	if os.Getenv(uniqueEnvName) == "1" {
		fatalCallback()
		os.Exit(0) // Shouldn't reach here.
	}

	cmd := exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = append(os.Environ(), uniqueEnvName+"=1")
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	t.Logf("Captured output: '%s'.", output.String())
	var exitError *exec.ExitError
	if err == nil || (errors.As(err, &exitError) && exitError.ExitCode() != 1) {
		t.Fatalf("Process ran with err '%v' but want exit status 1.", err)
	}
}

func TestFatal(t *testing.T) {
	testFatalScenario(t, "TEST_FATAL", "TestFatal", func() {
		logger.Fatal("Should call os.Exit(1).")
	})
}

func TestFatalf(t *testing.T) {
	testFatalScenario(t, "TEST_FATALF", "TestFatalf", func() {
		logger.Fatalf("Should call %s.", "os.Exit(1)")
	})
}

func TestFatalFn(t *testing.T) {
	testFatalScenario(t, "TEST_FATAL_FN", "TestFatalFn", func() {
		logger.FatalFn(func() []any {
			return []any{"Should call os.Exit(1)."}
		})
	})
}

func TestLoggerConcurrency(t *testing.T) {
	t.Cleanup(func() {
		logger.SetOutput(os.Stdout)
		logger.SetLevel(logger.LevelInfo)
		logger.SetFormatter(logger.DefaultFormatter)
	})

	const threadCount = 4
	const opsPerThread = 1000

	wg := &sync.WaitGroup{}
	waitChan := make(chan struct{})

	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-waitChan
			for k := 0; k < opsPerThread; k++ {
				logger.SetOutput(&testWriter{})
				logger.SetFormatter(logger.DefaultFormatter)
				for _, level := range []logger.LogLevel{logger.LevelError, logger.LevelWarn, logger.LevelInfo, logger.LevelDebug, logger.LevelTrace} {
					logger.SetLevel(level)
					ctx := context.Background()
					logEntry := logger.AddField(&ctx, "test", "test")
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
		}()
	}

	close(waitChan)
	wg.Wait()
}
