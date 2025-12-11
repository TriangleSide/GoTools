package structs_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type unmarshallTestStruct struct {
	Value string
}

func (t *unmarshallTestStruct) UnmarshalText(text []byte) error {
	t.Value = string(text)
	return nil
}

type failingUnmarshallTestStruct struct{}

func (t *failingUnmarshallTestStruct) UnmarshalText(_ []byte) error {
	return errors.New("intentional unmarshal failure")
}

type testDeepEmbeddedStruct struct {
	DeepEmbeddedValue *string
}

type testEmbeddedStruct struct {
	testDeepEmbeddedStruct
	EmbeddedValue string
}

type testInternalStruct struct {
	Value string `json:"value"`
}

type testStruct struct {
	testEmbeddedStruct

	StringValue     string
	IntValue        int
	Int8Value       int8
	Int16Value      int16
	Int32Value      int32
	Int64Value      int64
	UintValue       uint
	Uint8Value      uint8
	Uint16Value     uint16
	Uint32Value     uint32
	Uint64Value     uint64
	Float32Value    float32
	FloatValue      float64
	BoolValue       bool
	StructValue     testInternalStruct
	MapValue        map[string]testInternalStruct
	UnmarshallValue unmarshallTestStruct
	TimeValue       time.Time

	StringPtrValue     *string
	IntPtrValue        *int
	Int8PtrValue       *int8
	Int16PtrValue      *int16
	Int32PtrValue      *int32
	Int64PtrValue      *int64
	UintPtrValue       *uint
	Uint8PtrValue      *uint8
	Uint16PtrValue     *uint16
	Uint32PtrValue     *uint32
	Uint64PtrValue     *uint64
	Float32PtrValue    *float32
	FloatPtrValue      *float64
	BoolPtrValue       *bool
	StructPtrValue     *testInternalStruct
	MapPtrValue        *map[string]testInternalStruct
	UnmarshallPtrValue *unmarshallTestStruct
	TimePtrValue       *time.Time

	ListStringValue []string
	ListIntValue    []int
	ListFloatValue  []float64
	ListBoolValue   []bool
	ListStructValue []testInternalStruct

	ListStringPtrValue []*string
	ListIntPtrValue    []*int
	ListFloatPtrValue  []*float64
	ListBoolPtrValue   []*bool
	ListStructPtrValue []*testInternalStruct

	FailingUnmarshallValue failingUnmarshallTestStruct

	UnhandledValue uintptr
}

func TestAssignToField_NonStructPointer_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		_ = structs.AssignToField(new(int), "StringValue", "test")
	}, "obj must be a pointer to a struct")
}

func TestAssignToField_UnknownField_Panics(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.PanicPart(t, func() {
		_ = structs.AssignToField(values, "NonExistentField", "some value")
	}, "no field 'NonExistentField' in struct")
}

func TestAssignToField_TimeField_SetsValue(t *testing.T) {
	t.Parallel()
	const setValue = "2024-01-01T12:34:56Z"
	expectedTime, err := time.Parse(time.RFC3339, setValue)
	assert.NoError(t, err)
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "TimeValue", setValue))
	assert.Equals(t, expectedTime, values.TimeValue)
	assert.NoError(t, structs.AssignToField(values, "TimePtrValue", setValue))
	assert.Equals(t, expectedTime, *values.TimePtrValue)
}

func TestAssignToField_ValidValues_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	subTests := []struct {
		name      string
		fieldName string
		strValue  string
		value     func(values *testStruct) any
		expected  any
	}{
		{"EmbeddedValue", "EmbeddedValue", "embedded", func(ts *testStruct) any { return ts.EmbeddedValue }, "embedded"},
		{"DeepEmbeddedValue", "DeepEmbeddedValue", "deepEmbedded", func(ts *testStruct) any { return *ts.DeepEmbeddedValue }, "deepEmbedded"},
		{"StringValue", "StringValue", "str", func(ts *testStruct) any { return ts.StringValue }, "str"},
		{"StringPtrValue", "StringPtrValue", "strPtr", func(ts *testStruct) any { return *ts.StringPtrValue }, "strPtr"},
		{"IntValue", "IntValue", "123", func(ts *testStruct) any { return ts.IntValue }, 123},
		{"IntPtrValue", "IntPtrValue", "123", func(ts *testStruct) any { return *ts.IntPtrValue }, 123},
		{"Int8Value", "Int8Value", "-32", func(ts *testStruct) any { return ts.Int8Value }, int8(-32)},
		{"Int8PtrValue", "Int8PtrValue", "-32", func(ts *testStruct) any { return *ts.Int8PtrValue }, int8(-32)},
		{"Int16Value", "Int16Value", "-1000", func(ts *testStruct) any { return ts.Int16Value }, int16(-1000)},
		{"Int16PtrValue", "Int16PtrValue", "-1000", func(ts *testStruct) any { return *ts.Int16PtrValue }, int16(-1000)},
		{"Int32Value", "Int32Value", "-100000", func(ts *testStruct) any { return ts.Int32Value }, int32(-100000)},
		{"Int32PtrValue", "Int32PtrValue", "-100000", func(ts *testStruct) any { return *ts.Int32PtrValue }, int32(-100000)},
		{"Int64Value", "Int64Value", "-9223372036854775808", func(ts *testStruct) any { return ts.Int64Value }, int64(-9223372036854775808)},
		{"Int64PtrValue", "Int64PtrValue", "-9223372036854775808", func(ts *testStruct) any { return *ts.Int64PtrValue }, int64(-9223372036854775808)},
		{"UintValue", "UintValue", "456", func(ts *testStruct) any { return ts.UintValue }, uint(456)},
		{"UintPtrValue", "UintPtrValue", "456", func(ts *testStruct) any { return *ts.UintPtrValue }, uint(456)},
		{"Uint8Value", "Uint8Value", "99", func(ts *testStruct) any { return ts.Uint8Value }, uint8(99)},
		{"Uint8PtrValue", "Uint8PtrValue", "99", func(ts *testStruct) any { return *ts.Uint8PtrValue }, uint8(99)},
		{"Uint16Value", "Uint16Value", "65535", func(ts *testStruct) any { return ts.Uint16Value }, uint16(65535)},
		{"Uint16PtrValue", "Uint16PtrValue", "65535", func(ts *testStruct) any { return *ts.Uint16PtrValue }, uint16(65535)},
		{"Uint32Value", "Uint32Value", "4294967295", func(ts *testStruct) any { return ts.Uint32Value }, uint32(4294967295)},
		{"Uint32PtrValue", "Uint32PtrValue", "4294967295", func(ts *testStruct) any { return *ts.Uint32PtrValue }, uint32(4294967295)},
		{"Uint64Value", "Uint64Value", "18446744073709551615", func(ts *testStruct) any { return ts.Uint64Value }, uint64(18446744073709551615)},
		{"Uint64PtrValue", "Uint64PtrValue", "18446744073709551615", func(ts *testStruct) any { return *ts.Uint64PtrValue }, uint64(18446744073709551615)},
		{"Float32Value", "Float32Value", "123.45", func(ts *testStruct) any { return ts.Float32Value }, float32(123.45)},
		{"Float32PtrValue", "Float32PtrValue", "123.45", func(ts *testStruct) any { return *ts.Float32PtrValue }, float32(123.45)},
		{"FloatValue", "FloatValue", "678.90", func(ts *testStruct) any { return ts.FloatValue }, float64(678.90)},
		{"FloatPtrValue", "FloatPtrValue", "678.90", func(ts *testStruct) any { return *ts.FloatPtrValue }, float64(678.90)},
		{"BoolValueTrue", "BoolValue", "true", func(ts *testStruct) any { return ts.BoolValue }, true},
		{"BoolValueFalse", "BoolValue", "false", func(ts *testStruct) any { return ts.BoolValue }, false},
		{"BoolValue1", "BoolValue", "1", func(ts *testStruct) any { return ts.BoolValue }, true},
		{"BoolValue0", "BoolValue", "0", func(ts *testStruct) any { return ts.BoolValue }, false},
		{"BoolPtrValue", "BoolPtrValue", "true", func(ts *testStruct) any { return *ts.BoolPtrValue }, true},
		{"StructValue", "StructValue", `{"value":"nested"}`, func(ts *testStruct) any { return ts.StructValue }, testInternalStruct{Value: "nested"}},
		{"StructPtrValue", "StructPtrValue", `{"value":"nestedPtr"}`, func(ts *testStruct) any { return *ts.StructPtrValue }, testInternalStruct{Value: "nestedPtr"}},
		{"MapValue", "MapValue", `{"key1":{"value":"value1"}}`, func(ts *testStruct) any { return ts.MapValue }, map[string]testInternalStruct{"key1": {Value: "value1"}}},
		{"MapPtrValue", "MapPtrValue", `{"key1":{"value":"valuePtr"}}`, func(ts *testStruct) any { return *ts.MapPtrValue }, map[string]testInternalStruct{"key1": {Value: "valuePtr"}}},
		{"UnmarshallValue", "UnmarshallValue", "custom text", func(ts *testStruct) any { return ts.UnmarshallValue }, unmarshallTestStruct{Value: "custom text"}},
		{"UnmarshallPtrValue", "UnmarshallPtrValue", "custom text ptr", func(ts *testStruct) any { return *ts.UnmarshallPtrValue }, unmarshallTestStruct{Value: "custom text ptr"}},
		{"ListStringValue", "ListStringValue", `["one", "two", "three"]`, func(ts *testStruct) any { return ts.ListStringValue }, []string{"one", "two", "three"}},
		{"ListStringPtrValue", "ListStringPtrValue", `["one", "two", "three"]`, func(ts *testStruct) any { return ts.ListStringPtrValue }, []*string{ptr.Of("one"), ptr.Of("two"), ptr.Of("three")}},
		{"ListIntValue", "ListIntValue", `[1, 2, 3]`, func(ts *testStruct) any { return ts.ListIntValue }, []int{1, 2, 3}},
		{"ListIntPtrValue", "ListIntPtrValue", `[1, 2, 3]`, func(ts *testStruct) any { return ts.ListIntPtrValue }, []*int{ptr.Of(1), ptr.Of(2), ptr.Of(3)}},
		{"ListFloatValue", "ListFloatValue", `[1.1, 2.2, 3.3]`, func(ts *testStruct) any { return ts.ListFloatValue }, []float64{1.1, 2.2, 3.3}},
		{"ListFloatPtrValue", "ListFloatPtrValue", `[1.1, 2.2, 3.3]`, func(ts *testStruct) any { return ts.ListFloatPtrValue }, []*float64{ptr.Of(1.1), ptr.Of(2.2), ptr.Of(3.3)}},
		{"ListBoolValue", "ListBoolValue", `[true, false, true]`, func(ts *testStruct) any { return ts.ListBoolValue }, []bool{true, false, true}},
		{"ListBoolPtrValue", "ListBoolPtrValue", `[true, false, true]`, func(ts *testStruct) any { return ts.ListBoolPtrValue }, []*bool{ptr.Of(true), ptr.Of(false), ptr.Of(true)}},
		{"ListStructValue", "ListStructValue", `[{"value":"nested1"}, {"value":"nested2"}]`, func(ts *testStruct) any { return ts.ListStructValue }, []testInternalStruct{{Value: "nested1"}, {Value: "nested2"}}},
		{"ListStructPtrValue", "ListStructPtrValue", `[{"value":"nested1"}]`, func(ts *testStruct) any { return ts.ListStructPtrValue }, []*testInternalStruct{ptr.Of(testInternalStruct{Value: "nested1"})}},
	}
	for _, subTest := range subTests {
		t.Run(subTest.name, func(t *testing.T) {
			t.Parallel()
			values := &testStruct{}
			assert.NoError(t, structs.AssignToField(values, subTest.fieldName, subTest.strValue))
			assert.Equals(t, subTest.expected, subTest.value(values))
		})
	}
}

func TestAssignToField_InvalidValues_ReturnsError(t *testing.T) {
	t.Parallel()
	subTests := []struct {
		name      string
		fieldName string
		strValue  string
		errorPart string
	}{
		{"IntValueNotAnInteger", "IntValue", "not an integer", "strconv.ParseInt"},
		{"IntValueOverflow", "IntValue", strings.Repeat("1", 100), "strconv.ParseInt"},
		{"IntValueFloat", "IntValue", "123.456", "strconv.ParseInt"},
		{"UintValueNegative", "UintValue", "-123", "strconv.ParseUint"},
		{"FloatValueNotAFloat", "FloatValue", "not a float", "strconv.ParseFloat"},
		{"BoolValueInvalid", "BoolValue", "2", "strconv.ParseBool"},
		{"StructValueInvalidJSON", "StructValue", "not a json object", "json unmarshal error"},
		{"MapValueInvalidJSON", "MapValue", "not a json object", "json unmarshal error"},
		{"ListIntValueInvalidJSON", "ListIntValue", `["one", "two", "three"]`, "json unmarshal error"},
		{"ListIntPtrValueInvalidJSON", "ListIntPtrValue", `["one", "two", "three"]`, "json unmarshal error"},
		{"ListBoolValueInvalidJSON", "ListBoolValue", `["true", "false", "maybe"]`, "json unmarshal error"},
		{"ListStructValueInvalidJSON", "ListStructValue", `[{"value":"nested1"}, {"value":}]`, "json unmarshal error"},
		{"ListFloatValueInvalidJSON", "ListFloatValue", `["1.1", "two", "3.3"]`, "json unmarshal error"},
		{"TimeValueInvalid", "TimeValue", "this is not a time string", "parsing time"},
		{"TimePtrValueInvalid", "TimePtrValue", "this is not a time string", "parsing time"},
		{"TextUnmarshalerError", "FailingUnmarshallValue", "any value", "text unmarshal error"},
		{"UnhandledValueUnsupported", "UnhandledValue", "unhandled", "unsupported field type"},
	}
	for _, subTest := range subTests {
		t.Run(subTest.name, func(t *testing.T) {
			t.Parallel()
			values := &testStruct{}
			err := structs.AssignToField(values, subTest.fieldName, subTest.strValue)
			assert.ErrorPart(t, err, subTest.errorPart)
		})
	}
}
