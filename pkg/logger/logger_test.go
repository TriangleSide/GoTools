package logger_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/logger"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestSetOutput(t *testing.T) {
	var output bytes.Buffer
	logger.SetOutput(&output)
	t.Cleanup(func() {
		logger.SetOutput(os.Stdout)
	})
	logger.Error(nil, "test message")
	assert.Contains(t, string(output.Bytes()), "test message")
}

func TestLogger(t *testing.T) {
	ctx := context.Background()
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
			logger.Panic(ctx, "Panic Message")
		}, "Panic Message")
		assert.PanicPart(t, func() {
			logger.Panicf(ctx, "Panic %s", "Message F")
		}, "Panic Message F")
		assert.PanicPart(t, func() {
			logger.PanicFn(ctx, func() []any {
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
			output := setAndRecordOutput()
			assert.Equals(t, len(output.Bytes()), 0)
			logger.Error(ctx, "E")
			logger.Errorf(ctx, "E")
			logger.ErrorFn(ctx, func() []any { return []any{"E"} })
			logger.Warn(ctx, "W")
			logger.Warnf(ctx, "W")
			logger.WarnFn(ctx, func() []any { return []any{"W"} })
			logger.Info(ctx, "I")
			logger.Infof(ctx, "I")
			logger.InfoFn(ctx, func() []any { return []any{"I"} })
			logger.Debug(ctx, "D")
			logger.Debugf(ctx, "D")
			logger.DebugFn(ctx, func() []any { return []any{"D"} })
			logger.Trace(ctx, "T")
			logger.Tracef(ctx, "T")
			logger.TraceFn(ctx, func() []any { return []any{"T"} })
			assert.Equals(t, strings.ReplaceAll(string(output.Bytes()), "\n", ""), tc.expected)
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
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", testName))
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=1", uniqueEnvName))
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
		logger.Fatal(context.Background(), "Should call os.Exit(1).")
	})
}

func TestFatalf(t *testing.T) {
	testFatalScenario(t, "TEST_FATALF", "TestFatalf", func() {
		logger.Fatalf(context.Background(), "Should call %s.", "os.Exit(1)")
	})
}

func TestFatalFn(t *testing.T) {
	testFatalScenario(t, "TEST_FATAL_FN", "TestFatalFn", func() {
		logger.FatalFn(context.Background(), func() []any {
			return []any{"Should call os.Exit(1)."}
		})
	})
}
