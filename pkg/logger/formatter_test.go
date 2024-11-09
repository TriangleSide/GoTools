package logger

import (
	"context"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestFormatter(t *testing.T) {
	t.Run("when the fields are nil it should format without fields", func(t *testing.T) {
		msg := formatLog(nil, "test message")
		assert.Contains(t, msg, "test message")
	})

	t.Run("when fields are formatted it should include fields", func(t *testing.T) {
		ctx := context.Background()
		testLogger := AddFields(&ctx, map[string]any{
			"key1": "value1",
			"key2": 2,
		})
		testEntry := testLogger.(*entry)
		msg := formatLog(testEntry.fields, "test message")
		assert.Contains(t, msg, "key1=value1")
		assert.Contains(t, msg, "key2=2")
	})

	t.Run("when SetFormatter is called it should set custom formatter", func(t *testing.T) {
		t.Cleanup(func() {
			SetFormatter(DefaultFormatter)
		})
		SetFormatter(func(fields map[string]any, msg string) string {
			return "custom: " + msg
		})
		msg := formatLog(nil, "test message")
		assert.Contains(t, msg, "custom: test message")
	})
}
