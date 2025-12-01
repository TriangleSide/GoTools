package logger_test

import (
	"bytes"
	"context"
	"maps"
	"os"
	"strings"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestLoggerContext(t *testing.T) {
	setAndRecordOutput := func(t *testing.T) (*bytes.Buffer, map[string]any) {
		t.Helper()
		var output bytes.Buffer
		logger.SetOutput(&output)
		t.Cleanup(func() {
			logger.SetOutput(os.Stdout)
		})
		fieldsMap := make(map[string]any)
		logger.SetFormatter(func(fields map[string]any, msg string) string {
			maps.Copy(fieldsMap, fields)
			return msg
		})
		t.Cleanup(func() {
			logger.SetFormatter(logger.DefaultFormatter)
		})
		return &output, fieldsMap
	}

	t.Run("when no fields are added context should be empty", func(t *testing.T) {
		output, fieldsMap := setAndRecordOutput(t)
		ctx := context.Background()
		testLogger := logger.FromCtx(ctx)
		testLogger.Error("msg")
		assert.Equals(t, len(fieldsMap), 0)
		assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
	})

	t.Run("when adding a field to a context it should be included in the formatter", func(t *testing.T) {
		output, fieldsMap := setAndRecordOutput(t)
		_, testLogger := logger.AddField(context.Background(), "key", "value")
		testLogger.Error("msg")
		assert.Equals(t, len(fieldsMap), 1)
		value, ok := fieldsMap["key"]
		assert.True(t, ok)
		assert.Equals(t, value, "value")
		assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
	})

	t.Run("when adding multiple fields to a context it should be retrievable", func(t *testing.T) {
		output, fieldsMap := setAndRecordOutput(t)
		_, testLogger := logger.AddFields(context.Background(), map[string]any{"key1": "value1", "key2": 2})
		testLogger.Error("msg")
		assert.Equals(t, len(fieldsMap), 2)
		value1, ok1 := fieldsMap["key1"]
		assert.True(t, ok1)
		assert.Equals(t, value1, "value1")
		value2, ok2 := fieldsMap["key2"]
		assert.True(t, ok2)
		assert.Equals(t, value2, 2)
		assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
	})

	t.Run("when adding a field to a context with existing fields it should add correctly", func(t *testing.T) {
		output, fieldsMap := setAndRecordOutput(t)
		ctx := context.Background()
		ctx, _ = logger.AddFields(ctx, map[string]any{"key1": "value1"})
		ctx, _ = logger.AddField(ctx, "key2", 2)
		testLogger := logger.FromCtx(ctx)
		testLogger.Error("msg")
		assert.Equals(t, len(fieldsMap), 2)
		value1, ok1 := fieldsMap["key1"]
		assert.True(t, ok1)
		assert.Equals(t, value1, "value1")
		value2, ok2 := fieldsMap["key2"]
		assert.True(t, ok2)
		assert.Equals(t, value2, 2)
		assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
	})

	t.Run("when adding a field to a context with the same key it should overwrite the value", func(t *testing.T) {
		output, fieldsMap := setAndRecordOutput(t)
		ctx := context.Background()
		ctx, _ = logger.AddField(ctx, "key", "value")
		ctx, _ = logger.AddField(ctx, "key", "new_value")
		testLogger := logger.FromCtx(ctx)
		testLogger.Error("msg")
		assert.Equals(t, len(fieldsMap), 1)
		value1, ok1 := fieldsMap["key"]
		assert.True(t, ok1)
		assert.Equals(t, value1, "new_value")
		assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
	})

	t.Run("when adding fields to a context with overlapping keys it should overwrite the values", func(t *testing.T) {
		output, fieldsMap := setAndRecordOutput(t)
		ctx := context.Background()
		ctx, _ = logger.AddFields(ctx, map[string]any{"key1": "value1", "key2": "value2"})
		ctx, _ = logger.AddFields(ctx, map[string]any{"key2": "new_value2", "key3": "value3"})
		testLogger := logger.FromCtx(ctx)
		testLogger.Error("msg")
		assert.Equals(t, len(fieldsMap), 3)
		value1, ok1 := fieldsMap["key1"]
		assert.True(t, ok1)
		assert.Equals(t, value1, "value1")
		value2, ok2 := fieldsMap["key2"]
		assert.True(t, ok2)
		assert.Equals(t, value2, "new_value2")
		value3, ok3 := fieldsMap["key3"]
		assert.True(t, ok3)
		assert.Equals(t, value3, "value3")
		assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
	})
}
