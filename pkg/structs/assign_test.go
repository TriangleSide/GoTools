package structs_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/TriangleSide/go-toolkit/pkg/ptr"
	"github.com/TriangleSide/go-toolkit/pkg/structs"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
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

func TestAssignToField_NonStructPointer_ReturnsError(t *testing.T) {
	t.Parallel()
	err := structs.AssignToField(new(int), "StringValue", "test")
	assert.ErrorPart(t, err, "obj must be a pointer to a struct")
}

func TestAssignToField_UnknownField_ReturnsError(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	err := structs.AssignToField(values, "NonExistentField", "some value")
	assert.ErrorPart(t, err, "no field 'NonExistentField' in struct")
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

func TestAssignToField_EmbeddedValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "EmbeddedValue", "embedded"))
	assert.Equals(t, "embedded", values.EmbeddedValue)
}

func TestAssignToField_DeepEmbeddedValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "DeepEmbeddedValue", "deepEmbedded"))
	assert.Equals(t, "deepEmbedded", *values.DeepEmbeddedValue)
}

func TestAssignToField_StringValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "StringValue", "str"))
	assert.Equals(t, "str", values.StringValue)
}

func TestAssignToField_StringPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "StringPtrValue", "strPtr"))
	assert.Equals(t, "strPtr", *values.StringPtrValue)
}

func TestAssignToField_IntValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "IntValue", "123"))
	assert.Equals(t, 123, values.IntValue)
}

func TestAssignToField_IntPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "IntPtrValue", "123"))
	assert.Equals(t, 123, *values.IntPtrValue)
}

func TestAssignToField_Int8Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Int8Value", "-32"))
	assert.Equals(t, int8(-32), values.Int8Value)
}

func TestAssignToField_Int8PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Int8PtrValue", "-32"))
	assert.Equals(t, int8(-32), *values.Int8PtrValue)
}

func TestAssignToField_Int16Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Int16Value", "-1000"))
	assert.Equals(t, int16(-1000), values.Int16Value)
}

func TestAssignToField_Int16PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Int16PtrValue", "-1000"))
	assert.Equals(t, int16(-1000), *values.Int16PtrValue)
}

func TestAssignToField_Int32Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Int32Value", "-100000"))
	assert.Equals(t, int32(-100000), values.Int32Value)
}

func TestAssignToField_Int32PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Int32PtrValue", "-100000"))
	assert.Equals(t, int32(-100000), *values.Int32PtrValue)
}

func TestAssignToField_Int64Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Int64Value", "-9223372036854775808"))
	assert.Equals(t, int64(-9223372036854775808), values.Int64Value)
}

func TestAssignToField_Int64PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Int64PtrValue", "-9223372036854775808"))
	assert.Equals(t, int64(-9223372036854775808), *values.Int64PtrValue)
}

func TestAssignToField_UintValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "UintValue", "456"))
	assert.Equals(t, uint(456), values.UintValue)
}

func TestAssignToField_UintPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "UintPtrValue", "456"))
	assert.Equals(t, uint(456), *values.UintPtrValue)
}

func TestAssignToField_Uint8Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Uint8Value", "99"))
	assert.Equals(t, uint8(99), values.Uint8Value)
}

func TestAssignToField_Uint8PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Uint8PtrValue", "99"))
	assert.Equals(t, uint8(99), *values.Uint8PtrValue)
}

func TestAssignToField_Uint16Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Uint16Value", "65535"))
	assert.Equals(t, uint16(65535), values.Uint16Value)
}

func TestAssignToField_Uint16PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Uint16PtrValue", "65535"))
	assert.Equals(t, uint16(65535), *values.Uint16PtrValue)
}

func TestAssignToField_Uint32Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Uint32Value", "4294967295"))
	assert.Equals(t, uint32(4294967295), values.Uint32Value)
}

func TestAssignToField_Uint32PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Uint32PtrValue", "4294967295"))
	assert.Equals(t, uint32(4294967295), *values.Uint32PtrValue)
}

func TestAssignToField_Uint64Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Uint64Value", "18446744073709551615"))
	assert.Equals(t, uint64(18446744073709551615), values.Uint64Value)
}

func TestAssignToField_Uint64PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Uint64PtrValue", "18446744073709551615"))
	assert.Equals(t, uint64(18446744073709551615), *values.Uint64PtrValue)
}

func TestAssignToField_Float32Value_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Float32Value", "123.45"))
	assert.Equals(t, float32(123.45), values.Float32Value)
}

func TestAssignToField_Float32PtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "Float32PtrValue", "123.45"))
	assert.Equals(t, float32(123.45), *values.Float32PtrValue)
}

func TestAssignToField_FloatValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "FloatValue", "678.90"))
	assert.Equals(t, float64(678.90), values.FloatValue)
}

func TestAssignToField_FloatPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "FloatPtrValue", "678.90"))
	assert.Equals(t, float64(678.90), *values.FloatPtrValue)
}

func TestAssignToField_BoolValueTrue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "BoolValue", "true"))
	assert.Equals(t, true, values.BoolValue)
}

func TestAssignToField_BoolValueFalse_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "BoolValue", "false"))
	assert.Equals(t, false, values.BoolValue)
}

func TestAssignToField_BoolValue1_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "BoolValue", "1"))
	assert.Equals(t, true, values.BoolValue)
}

func TestAssignToField_BoolValue0_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "BoolValue", "0"))
	assert.Equals(t, false, values.BoolValue)
}

func TestAssignToField_BoolPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "BoolPtrValue", "true"))
	assert.Equals(t, true, *values.BoolPtrValue)
}

func TestAssignToField_StructValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "StructValue", `{"value":"nested"}`))
	assert.Equals(t, testInternalStruct{Value: "nested"}, values.StructValue)
}

func TestAssignToField_StructPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "StructPtrValue", `{"value":"nestedPtr"}`))
	assert.Equals(t, testInternalStruct{Value: "nestedPtr"}, *values.StructPtrValue)
}

func TestAssignToField_MapValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "MapValue", `{"key1":{"value":"value1"}}`))
	assert.Equals(t, map[string]testInternalStruct{"key1": {Value: "value1"}}, values.MapValue)
}

func TestAssignToField_MapPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "MapPtrValue", `{"key1":{"value":"valuePtr"}}`))
	assert.Equals(t, map[string]testInternalStruct{"key1": {Value: "valuePtr"}}, *values.MapPtrValue)
}

func TestAssignToField_UnmarshallValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "UnmarshallValue", "custom text"))
	assert.Equals(t, unmarshallTestStruct{Value: "custom text"}, values.UnmarshallValue)
}

func TestAssignToField_UnmarshallPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "UnmarshallPtrValue", "custom text ptr"))
	assert.Equals(t, unmarshallTestStruct{Value: "custom text ptr"}, *values.UnmarshallPtrValue)
}

func TestAssignToField_ListStringValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListStringValue", `["one", "two", "three"]`))
	assert.Equals(t, []string{"one", "two", "three"}, values.ListStringValue)
}

func TestAssignToField_ListStringPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListStringPtrValue", `["one", "two", "three"]`))
	assert.Equals(t, []*string{ptr.Of("one"), ptr.Of("two"), ptr.Of("three")}, values.ListStringPtrValue)
}

func TestAssignToField_ListIntValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListIntValue", `[1, 2, 3]`))
	assert.Equals(t, []int{1, 2, 3}, values.ListIntValue)
}

func TestAssignToField_ListIntPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListIntPtrValue", `[1, 2, 3]`))
	assert.Equals(t, []*int{ptr.Of(1), ptr.Of(2), ptr.Of(3)}, values.ListIntPtrValue)
}

func TestAssignToField_ListFloatValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListFloatValue", `[1.1, 2.2, 3.3]`))
	assert.Equals(t, []float64{1.1, 2.2, 3.3}, values.ListFloatValue)
}

func TestAssignToField_ListFloatPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListFloatPtrValue", `[1.1, 2.2, 3.3]`))
	assert.Equals(t, []*float64{ptr.Of(1.1), ptr.Of(2.2), ptr.Of(3.3)}, values.ListFloatPtrValue)
}

func TestAssignToField_ListBoolValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListBoolValue", `[true, false, true]`))
	assert.Equals(t, []bool{true, false, true}, values.ListBoolValue)
}

func TestAssignToField_ListBoolPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListBoolPtrValue", `[true, false, true]`))
	assert.Equals(t, []*bool{ptr.Of(true), ptr.Of(false), ptr.Of(true)}, values.ListBoolPtrValue)
}

func TestAssignToField_ListStructValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListStructValue", `[{"value":"nested1"}, {"value":"nested2"}]`))
	assert.Equals(t, []testInternalStruct{{Value: "nested1"}, {Value: "nested2"}}, values.ListStructValue)
}

func TestAssignToField_ListStructPtrValue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	values := &testStruct{}
	assert.NoError(t, structs.AssignToField(values, "ListStructPtrValue", `[{"value":"nested1"}]`))
	assert.Equals(t, []*testInternalStruct{ptr.Of(testInternalStruct{Value: "nested1"})}, values.ListStructPtrValue)
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
		{"UnhandledValueUnsupported", "UnhandledValue", "unhandled", "unsupported type"},
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
