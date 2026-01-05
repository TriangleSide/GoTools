package attribute_test

import (
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/telemetry/trace/attribute"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestString_ValidInput_CreatesStringAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.String("testKey", "testValue")
	assert.NotNil(t, attr)
	assert.Equals(t, "testKey", attr.Key())
	assert.Equals(t, attribute.TypeString, attr.Type())
	assert.Equals(t, "testValue", attr.StringValue())
}

func TestString_EmptyValues_CreatesAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.String("", "")
	assert.NotNil(t, attr)
	assert.Equals(t, "", attr.Key())
	assert.Equals(t, "", attr.StringValue())
}

func TestString_SpecialCharacters_PreservesValue(t *testing.T) {
	t.Parallel()
	attr := attribute.String("key", "test\n\t\"quoted\"")
	assert.NotNil(t, attr)
	assert.Equals(t, "test\n\t\"quoted\"", attr.StringValue())
}

func TestBool_TrueValue_CreatesBoolAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Bool("enabled", true)
	assert.NotNil(t, attr)
	assert.Equals(t, "enabled", attr.Key())
	assert.Equals(t, attribute.TypeBool, attr.Type())
	assert.Equals(t, true, attr.BoolValue())
}

func TestBool_FalseValue_CreatesBoolAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Bool("disabled", false)
	assert.NotNil(t, attr)
	assert.Equals(t, "disabled", attr.Key())
	assert.Equals(t, attribute.TypeBool, attr.Type())
	assert.Equals(t, false, attr.BoolValue())
}

func TestInt_PositiveValue_CreatesIntAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Int("count", 42)
	assert.NotNil(t, attr)
	assert.Equals(t, "count", attr.Key())
	assert.Equals(t, attribute.TypeInt, attr.Type())
	assert.Equals(t, int64(42), attr.IntValue())
}

func TestInt_NegativeValue_CreatesIntAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Int("temperature", -273)
	assert.NotNil(t, attr)
	assert.Equals(t, "temperature", attr.Key())
	assert.Equals(t, int64(-273), attr.IntValue())
}

func TestInt_ZeroValue_CreatesIntAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Int("zero", 0)
	assert.NotNil(t, attr)
	assert.Equals(t, int64(0), attr.IntValue())
}

func TestInt_MaxValue_CreatesIntAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Int("max", 9223372036854775807)
	assert.NotNil(t, attr)
	assert.Equals(t, int64(9223372036854775807), attr.IntValue())
}

func TestInt_MinValue_CreatesIntAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Int("min", -9223372036854775808)
	assert.NotNil(t, attr)
	assert.Equals(t, int64(-9223372036854775808), attr.IntValue())
}

func TestFloat_PositiveValue_CreatesFloatAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Float("pi", 3.14159)
	assert.NotNil(t, attr)
	assert.Equals(t, "pi", attr.Key())
	assert.Equals(t, attribute.TypeFloat, attr.Type())
	assert.FloatEquals(t, 3.14159, attr.FloatValue(), 0.00001)
}

func TestFloat_NegativeValue_CreatesFloatAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Float("temperature", -40.5)
	assert.NotNil(t, attr)
	assert.Equals(t, "temperature", attr.Key())
	assert.FloatEquals(t, -40.5, attr.FloatValue(), 0.00001)
}

func TestFloat_ZeroValue_CreatesFloatAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Float("zero", 0.0)
	assert.NotNil(t, attr)
	assert.FloatEquals(t, 0.0, attr.FloatValue(), 0.00001)
}

func TestFloat_VerySmallValue_CreatesFloatAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Float("small", 0.000001)
	assert.NotNil(t, attr)
	assert.FloatEquals(t, 0.000001, attr.FloatValue(), 0.0000001)
}

func TestFloat_VeryLargeValue_CreatesFloatAttribute(t *testing.T) {
	t.Parallel()
	attr := attribute.Float("large", 1e308)
	assert.NotNil(t, attr)
	assert.FloatEquals(t, 1e308, attr.FloatValue(), 1e300)
}

func TestConstructors_ConcurrentCreation_IsThreadSafe(t *testing.T) {
	t.Parallel()
	const goroutines = 10
	const iterations = 100
	var waitGroup sync.WaitGroup
	for goroutineIdx := range goroutines {
		waitGroup.Go(func() {
			for iterIdx := range iterations {
				attr := attribute.String("key", "value")
				assert.NotNil(t, attr, assert.Continue())
				attr = attribute.Int("key", int64(iterIdx))
				assert.NotNil(t, attr, assert.Continue())
				attr = attribute.Float("key", float64(goroutineIdx))
				assert.NotNil(t, attr, assert.Continue())
				attr = attribute.Bool("key", iterIdx%2 == 0)
				assert.NotNil(t, attr, assert.Continue())
			}
		})
	}
	waitGroup.Wait()
}
