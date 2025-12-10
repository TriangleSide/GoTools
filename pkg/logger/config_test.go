package logger_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestMustConfigure_ConfigProviderSucceeds_SetsLoggerLevel(t *testing.T) {
	t.Cleanup(func() {
		logger.SetLevel(logger.LevelInfo)
	})
	logger.MustConfigure(logger.WithConfigProvider(func() (*logger.Config, error) {
		return &logger.Config{
			LogLevel: "debug",
		}, nil
	}))
	assert.Equals(t, logger.GetLevel(), logger.LevelDebug)
}

func TestMustConfigure_IncorrectLevel_Panics(t *testing.T) {
	assert.PanicExact(t, func() {
		logger.MustConfigure(logger.WithConfigProvider(func() (*logger.Config, error) {
			return &logger.Config{
				LogLevel: "incorrect",
			}, nil
		}))
	}, "Failed to parse the log level (invalid log level: incorrect).")
}

func TestMustConfigure_ConfigProviderFails_Panics(t *testing.T) {
	assert.PanicExact(t, func() {
		logger.MustConfigure(logger.WithConfigProvider(func() (*logger.Config, error) {
			return nil, errors.New("config error")
		}))
	}, "Failed to get logger config (config error).")
}

func TestMustConfigure_OutputProviderSucceeds_DoesNotPanic(t *testing.T) {
	t.Cleanup(func() {
		logger.SetOutput(os.Stdout)
	})
	var outputBuffer bytes.Buffer
	logger.MustConfigure(logger.WithOutputProvider(func() (io.Writer, error) {
		return &outputBuffer, nil
	}))
	logger.Error(context.Background(), "test message")
	assert.Contains(t, outputBuffer.String(), "test message")
}

func TestMustConfigure_OutputProviderFails_Panics(t *testing.T) {
	assert.PanicExact(t, func() {
		logger.MustConfigure(logger.WithOutputProvider(func() (io.Writer, error) {
			return nil, errors.New("output error")
		}))
	}, "Failed to get logger output (output error).")
}

func TestMustConfigure_DefaultsUsed_SetsDefaults(t *testing.T) {
	t.Cleanup(func() {
		logger.SetLevel(logger.LevelInfo)
	})
	logger.SetLevel(logger.LevelTrace)
	logger.MustConfigure()
	assert.Equals(t, logger.GetLevel(), logger.LevelInfo)
}
