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

func TestMustConfigure(t *testing.T) {
	t.Run("when the config provider succeeds it sets the logger level", func(t *testing.T) {
		t.Cleanup(func() {
			logger.SetLevel(logger.LevelInfo)
		})
		logger.MustConfigure(logger.WithConfigProvider(func() (*logger.Config, error) {
			return &logger.Config{
				LogLevel: "debug",
			}, nil
		}))
		assert.Equals(t, logger.GetLevel(), logger.LevelDebug)
	})

	t.Run("when the level is incorrect it should panic", func(t *testing.T) {
		assert.PanicExact(t, func() {
			logger.MustConfigure(logger.WithConfigProvider(func() (*logger.Config, error) {
				return &logger.Config{
					LogLevel: "incorrect",
				}, nil
			}))
		}, "Failed to parse the log level (invalid log level: incorrect).")
	})

	t.Run("when the config provider fails it should panic", func(t *testing.T) {
		assert.PanicExact(t, func() {
			logger.MustConfigure(logger.WithConfigProvider(func() (*logger.Config, error) {
				return nil, errors.New("config error")
			}))
		}, "Failed to get logger config (config error).")
	})

	t.Run("when the output provider succeeds does not panic", func(t *testing.T) {
		t.Cleanup(func() {
			logger.SetOutput(os.Stdout)
		})
		var outputBuffer bytes.Buffer
		logger.MustConfigure(logger.WithOutputProvider(func() (io.Writer, error) {
			return &outputBuffer, nil
		}))
		logger.Error(context.Background(), "test message")
		assert.Contains(t, outputBuffer.String(), "test message")
	})

	t.Run("when the output provider fails it should panic", func(t *testing.T) {
		assert.PanicExact(t, func() {
			logger.MustConfigure(logger.WithOutputProvider(func() (io.Writer, error) {
				return nil, errors.New("output error")
			}))
		}, "Failed to get logger output (output error).")
	})

	t.Run("when the defaults are used it should set the defaults", func(t *testing.T) {
		t.Cleanup(func() {
			logger.SetLevel(logger.LevelInfo)
		})
		logger.SetLevel(logger.LevelTrace)
		logger.MustConfigure()
		assert.Equals(t, logger.GetLevel(), logger.LevelInfo)
	})
}
