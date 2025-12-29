package attribute_test

import (
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/trace/attribute"
)

func TestKey_StringAttribute_ReturnsKey(t *testing.T) {
	t.Parallel()
	attr := attribute.String("myKey", "value")
	assert.Equals(t, "myKey", attr.Key())
}

func TestType_StringAttribute_ReturnsTypeString(t *testing.T) {
	t.Parallel()
	attr := attribute.String("key", "value")
	assert.Equals(t, attribute.TypeString, attr.Type())
}

func TestType_IntAttribute_ReturnsTypeInt(t *testing.T) {
	t.Parallel()
	attr := attribute.Int("key", 42)
	assert.Equals(t, attribute.TypeInt, attr.Type())
}

func TestType_FloatAttribute_ReturnsTypeFloat(t *testing.T) {
	t.Parallel()
	attr := attribute.Float("key", 3.14)
	assert.Equals(t, attribute.TypeFloat, attr.Type())
}

func TestType_BoolAttribute_ReturnsTypeBool(t *testing.T) {
	t.Parallel()
	attr := attribute.Bool("key", true)
	assert.Equals(t, attribute.TypeBool, attr.Type())
}

func TestStringValue_StringAttribute_ReturnsValue(t *testing.T) {
	t.Parallel()
	attr := attribute.String("key", "testValue")
	assert.Equals(t, "testValue", attr.StringValue())
}

func TestIntValue_IntAttribute_ReturnsValue(t *testing.T) {
	t.Parallel()
	attr := attribute.Int("key", 12345)
	assert.Equals(t, int64(12345), attr.IntValue())
}

func TestFloatValue_FloatAttribute_ReturnsValue(t *testing.T) {
	t.Parallel()
	attr := attribute.Float("key", 2.71828)
	assert.FloatEquals(t, 2.71828, attr.FloatValue(), 0.00001)
}

func TestBoolValue_BoolAttribute_ReturnsValue(t *testing.T) {
	t.Parallel()
	attr := attribute.Bool("key", true)
	assert.Equals(t, true, attr.BoolValue())
}

func TestAsString_IntType_ReturnsFormattedInt(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		value    int64
		expected string
	}{
		{
			name:     "positive",
			value:    12345,
			expected: "12345",
		},
		{
			name:     "negative",
			value:    -67890,
			expected: "-67890",
		},
		{
			name:     "zero",
			value:    0,
			expected: "0",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			attr := attribute.Int("key", tc.value)
			assert.Equals(t, tc.expected, attr.AsString())
		})
	}
}

func TestAsString_FloatType_ReturnsFormattedFloat(t *testing.T) {
	t.Parallel()
	attr := attribute.Float("key", 3.14)
	result := attr.AsString()
	assert.Contains(t, result, "3.14")
}

func TestAsString_BoolType_ReturnsFormattedBool(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		value    bool
		expected string
	}{
		{
			name:     "true",
			value:    true,
			expected: "true",
		},
		{
			name:     "false",
			value:    false,
			expected: "false",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			attr := attribute.Bool("key", tc.value)
			assert.Equals(t, tc.expected, attr.AsString())
		})
	}
}

func TestAsString_StringType_ReturnsStringValue(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "non-empty",
			value:    "hello world",
			expected: "hello world",
		},
		{
			name:     "empty",
			value:    "",
			expected: "",
		},
		{
			name:     "special-characters",
			value:    "test\n\t\"quoted\"",
			expected: "test\n\t\"quoted\"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			attr := attribute.String("key", tc.value)
			assert.Equals(t, tc.expected, attr.AsString())
		})
	}
}

func TestAttribute_ConcurrentReads_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 5000
	attr := attribute.String("key", "value")
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Go(func() {
			for range iterations {
				_ = attr.Key()
				_ = attr.Type()
				_ = attr.StringValue()
				_ = attr.AsString()
			}
		})
	}
	waitGroup.Wait()
}

func TestType_String_ReturnsCorrectString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		attrType attribute.Type
		expected string
	}{
		{
			name:     "String",
			attrType: attribute.TypeString,
			expected: "String",
		},
		{
			name:     "Int",
			attrType: attribute.TypeInt,
			expected: "Int",
		},
		{
			name:     "Float",
			attrType: attribute.TypeFloat,
			expected: "Float",
		},
		{
			name:     "Bool",
			attrType: attribute.TypeBool,
			expected: "Bool",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equals(t, tc.expected, tc.attrType.String())
		})
	}
}

func TestType_String_UnknownType_ReturnsUnknown(t *testing.T) {
	t.Parallel()
	unknownType := attribute.Type(999)
	assert.Equals(t, "Unknown", unknownType.String())
}
