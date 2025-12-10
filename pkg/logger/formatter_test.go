package logger_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/logger"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func formatterRecordOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var output bytes.Buffer
	logger.SetOutput(&output)
	t.Cleanup(func() {
		logger.SetOutput(os.Stdout)
	})
	return &output
}

func TestDefaultFormatter_NilFields_FormatsWithoutFields(t *testing.T) {
	buf := formatterRecordOutput(t)
	logger.Info("test message")
	assert.Contains(t, buf.String(), "test message")
}

func TestDefaultFormatter_WithFields_IncludesFields(t *testing.T) {
	buf := formatterRecordOutput(t)
	_, log := logger.AddFields(context.Background(), map[string]any{
		"key1": "value1",
		"key2": 2,
	})
	log.Info("test message")
	assert.Contains(t, buf.String(), "key1=value1")
	assert.Contains(t, buf.String(), "key2=2")
}

func TestSetFormatter_CustomFormatter_SetsCustomFormatter(t *testing.T) {
	t.Cleanup(func() {
		logger.SetFormatter(logger.DefaultFormatter)
	})
	logger.SetFormatter(func(fields map[string]any, msg string) string {
		return "custom: " + msg
	})
	buf := formatterRecordOutput(t)
	logger.Info("test message")
	assert.Contains(t, buf.String(), "custom: test message")
}

func TestDefaultFormatter_EmptyMap_FormatsWithoutFields(t *testing.T) {
	buf := formatterRecordOutput(t)
	_, log := logger.AddFields(context.Background(), map[string]any{})
	log.Info("test message")
	assert.Contains(t, buf.String(), "test message")
}

func TestDefaultFormatter_EmptyMessage_FormatsWithEmptyMessage(t *testing.T) {
	buf := formatterRecordOutput(t)
	logger.Info("")
	output := buf.String()
	assert.NotEquals(t, output, "")
}

func TestDefaultFormatter_VariousFieldTypes_FormatsAllTypes(t *testing.T) {
	buf := formatterRecordOutput(t)
	_, log := logger.AddFields(context.Background(), map[string]any{
		"bool":   true,
		"float":  3.14,
		"nil":    nil,
		"slice":  []int{1, 2, 3},
		"struct": struct{ Name string }{Name: "test"},
	})
	log.Info("test message")
	output := buf.String()
	assert.Contains(t, output, "bool=true")
	assert.Contains(t, output, "float=3.14")
	assert.Contains(t, output, "nil=<nil>")
	assert.Contains(t, output, "slice=[1 2 3]")
	assert.Contains(t, output, "struct={test}")
}

func TestDefaultFormatter_SingleField_FormatsWithSingleField(t *testing.T) {
	buf := formatterRecordOutput(t)
	_, log := logger.AddFields(context.Background(), map[string]any{
		"only": "one",
	})
	log.Info("test message")
	output := buf.String()
	assert.Contains(t, output, "only=one")
	assert.Contains(t, output, "test message")
}
