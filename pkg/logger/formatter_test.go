package logger_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestFormatter(t *testing.T) {
	recordOutput := func(t *testing.T) *bytes.Buffer {
		t.Helper()
		var output bytes.Buffer
		logger.SetOutput(&output)
		t.Cleanup(func() {
			logger.SetOutput(os.Stdout)
		})
		return &output
	}

	t.Run("when the fields are nil it should format without fields", func(t *testing.T) {
		buf := recordOutput(t)
		logger.Info("test message")
		assert.Contains(t, buf.String(), "test message")
	})

	t.Run("when fields are formatted it should include fields", func(t *testing.T) {
		buf := recordOutput(t)
		_, log := logger.AddFields(context.Background(), map[string]any{
			"key1": "value1",
			"key2": 2,
		})
		log.Info("test message")
		assert.Contains(t, buf.String(), "key1=value1")
		assert.Contains(t, buf.String(), "key2=2")
	})

	t.Run("when SetFormatter is called it should set custom formatter", func(t *testing.T) {
		t.Cleanup(func() {
			logger.SetFormatter(logger.DefaultFormatter)
		})
		logger.SetFormatter(func(fields map[string]any, msg string) string {
			return "custom: " + msg
		})
		buf := recordOutput(t)
		logger.Info("test message")
		assert.Contains(t, buf.String(), "custom: test message")
	})
}
