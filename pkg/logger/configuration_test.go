package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/config"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestMustConfigure(t *testing.T) {
	t.Cleanup(func() {
		SetOutput(os.Stdout)
		SetLevel(LevelInfo)
	})

	t.Run("when the config provider succeeds it sets the logger level", func(t *testing.T) {
		MustConfigure(WithConfigProvider(func() (*config.Logger, error) {
			return &config.Logger{
				LogLevel: "debug",
			}, nil
		}))
		assert.Equals(t, appLogLevel, LevelDebug)
	})

	t.Run("when the level is incorrect it should panic", func(t *testing.T) {
		assert.PanicExact(t, func() {
			MustConfigure(WithConfigProvider(func() (*config.Logger, error) {
				return &config.Logger{
					LogLevel: "incorrect",
				}, nil
			}))
		}, "Failed to parse the log level (invalid log level: incorrect).")
	})

	t.Run("when the config provider fails it should panic", func(t *testing.T) {
		assert.PanicExact(t, func() {
			MustConfigure(WithConfigProvider(func() (*config.Logger, error) {
				return nil, errors.New("config error")
			}))
		}, "Failed to get logger config (config error).")
	})

	t.Run("when the output provider succeeds does not panic", func(t *testing.T) {
		var outputBuffer bytes.Buffer
		MustConfigure(WithOutputProvider(func() (io.Writer, error) {
			return &outputBuffer, nil
		}))
		Error(context.Background(), "test message")
		assert.Contains(t, outputBuffer.String(), "test message")
	})

	t.Run("when the output provider fails it should panic", func(t *testing.T) {
		assert.PanicExact(t, func() {
			MustConfigure(WithOutputProvider(func() (io.Writer, error) {
				return nil, errors.New("output error")
			}))
		}, "Failed to get logger output (output error).")
	})

	t.Run("when the defaults are used it should set the defaults", func(t *testing.T) {
		SetLevel(LevelTrace)
		MustConfigure()
		assert.Equals(t, appLogLevel, LevelInfo)
	})
}
