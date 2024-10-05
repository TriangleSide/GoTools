package logger

import (
	"context"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestFormatter(t *testing.T) {
	t.Run("when context is nil it should format without fields", func(t *testing.T) {
		msg := formatLog(nil, "test message")
		assert.Contains(t, msg, "test message")
	})

	t.Run("when context has fields it should include fields", func(t *testing.T) {
		ctx := WithFields(context.Background(), map[string]any{
			"key1": "value1",
			"key2": 2,
		})
		msg := formatLog(ctx, "test message")
		assert.Contains(t, msg, "key1=value1")
		assert.Contains(t, msg, "key2=2")
	})

	t.Run("when fields are not map[string]any it should panic", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), contextKey, "invalid")
		assert.PanicPart(t, func() {
			formatLog(ctx, "test message")
		}, "logger context fields is not the correct type")
	})

	t.Run("when SetFormatter is called it should set custom formatter", func(t *testing.T) {
		t.Cleanup(func() {
			SetFormatter(defaultLogFormatter)
		})
		SetFormatter(func(fields map[string]any, msg string) string {
			return "custom: " + msg
		})
		msg := formatLog(nil, "test message")
		assert.Contains(t, msg, "custom: test message")
	})
}
