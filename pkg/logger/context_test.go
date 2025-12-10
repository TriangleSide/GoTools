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

func setAndRecordOutput(t *testing.T) (*bytes.Buffer, map[string]any) {
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

func TestFromCtx_NoFieldsAdded_ContextShouldBeEmpty(t *testing.T) {
	output, fieldsMap := setAndRecordOutput(t)
	ctx := context.Background()
	testLogger := logger.FromCtx(ctx)
	testLogger.Error("msg")
	assert.Equals(t, len(fieldsMap), 0)
	assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
}

func TestAddField_SingleField_IncludedInFormatter(t *testing.T) {
	output, fieldsMap := setAndRecordOutput(t)
	_, testLogger := logger.AddField(context.Background(), "key", "value")
	testLogger.Error("msg")
	assert.Equals(t, len(fieldsMap), 1)
	value, ok := fieldsMap["key"]
	assert.True(t, ok)
	assert.Equals(t, value, "value")
	assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
}

func TestAddFields_MultipleFields_AllRetrievable(t *testing.T) {
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
}

func TestAddField_ContextWithExistingFields_AddsCorrectly(t *testing.T) {
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
}

func TestAddField_SameKey_OverwritesValue(t *testing.T) {
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
}

func TestAddFields_OverlappingKeys_OverwritesValues(t *testing.T) {
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
}

func TestAddField_EmptyStringKey_AddsFieldWithEmptyKey(t *testing.T) {
	output, fieldsMap := setAndRecordOutput(t)
	_, testLogger := logger.AddField(context.Background(), "", "value")
	testLogger.Error("msg")
	assert.Equals(t, len(fieldsMap), 1)
	value, ok := fieldsMap[""]
	assert.True(t, ok)
	assert.Equals(t, value, "value")
	assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
}

func TestAddField_NilValue_AddsFieldWithNilValue(t *testing.T) {
	output, fieldsMap := setAndRecordOutput(t)
	_, testLogger := logger.AddField(context.Background(), "key", nil)
	testLogger.Error("msg")
	assert.Equals(t, len(fieldsMap), 1)
	value, ok := fieldsMap["key"]
	assert.True(t, ok)
	assert.Equals(t, value, nil)
	assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
}

func TestAddFields_EmptyMap_NoFieldsAdded(t *testing.T) {
	output, fieldsMap := setAndRecordOutput(t)
	_, testLogger := logger.AddFields(context.Background(), map[string]any{})
	testLogger.Error("msg")
	assert.Equals(t, len(fieldsMap), 0)
	assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
}

func TestAddFields_NilMap_NoFieldsAdded(t *testing.T) {
	output, fieldsMap := setAndRecordOutput(t)
	_, testLogger := logger.AddFields(context.Background(), nil)
	testLogger.Error("msg")
	assert.Equals(t, len(fieldsMap), 0)
	assert.Equals(t, strings.ReplaceAll(output.String(), "\n", ""), "msg")
}
