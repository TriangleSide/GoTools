package reflection_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/TriangleSide/go-toolkit/pkg/ptr"
	"github.com/TriangleSide/go-toolkit/pkg/reflection"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
)

type textUnmarshaler struct {
	Value string
}

func (t *textUnmarshaler) UnmarshalText(text []byte) error {
	t.Value = string(text)
	return nil
}

type failingTextUnmarshaler struct{}

func (t *failingTextUnmarshaler) UnmarshalText(_ []byte) error {
	return errors.New("intentional unmarshal failure")
}

type testInternalStruct struct {
	Value string `json:"value"`
}

func TestAssignFromString_NotSettable_ReturnsError(t *testing.T) {
	t.Parallel()
	err := reflection.AssignFromString(reflect.ValueOf(123), "test")
	assert.ErrorPart(t, err, "value is not settable")
}

func TestAssignFromString_InvalidValue_ReturnsError(t *testing.T) {
	t.Parallel()
	var invalidValue reflect.Value
	err := reflection.AssignFromString(invalidValue, "test")
	assert.ErrorPart(t, err, "value is not valid")
}

func TestAssignFromString_String_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value string
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "hello world")
	assert.NoError(t, err)
	assert.Equals(t, "hello world", value)
}

func TestAssignFromString_StringPtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value *string
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "hello ptr")
	assert.NoError(t, err)
	assert.Equals(t, "hello ptr", *value)
}

func TestAssignFromString_Int_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "123")
	assert.NoError(t, err)
	assert.Equals(t, 123, value)
}

func TestAssignFromString_IntPtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value *int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "456")
	assert.NoError(t, err)
	assert.Equals(t, 456, *value)
}

func TestAssignFromString_Int8_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value int8
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "-32")
	assert.NoError(t, err)
	assert.Equals(t, int8(-32), value)
}

func TestAssignFromString_Int16_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value int16
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "-1000")
	assert.NoError(t, err)
	assert.Equals(t, int16(-1000), value)
}

func TestAssignFromString_Int32_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value int32
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "-100000")
	assert.NoError(t, err)
	assert.Equals(t, int32(-100000), value)
}

func TestAssignFromString_Int64_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value int64
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "-9223372036854775808")
	assert.NoError(t, err)
	assert.Equals(t, int64(-9223372036854775808), value)
}

func TestAssignFromString_Uint_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value uint
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "456")
	assert.NoError(t, err)
	assert.Equals(t, uint(456), value)
}

func TestAssignFromString_Uint8_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value uint8
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "99")
	assert.NoError(t, err)
	assert.Equals(t, uint8(99), value)
}

func TestAssignFromString_Uint16_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value uint16
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "65535")
	assert.NoError(t, err)
	assert.Equals(t, uint16(65535), value)
}

func TestAssignFromString_Uint32_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value uint32
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "4294967295")
	assert.NoError(t, err)
	assert.Equals(t, uint32(4294967295), value)
}

func TestAssignFromString_Uint64_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value uint64
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "18446744073709551615")
	assert.NoError(t, err)
	assert.Equals(t, uint64(18446744073709551615), value)
}

func TestAssignFromString_Float32_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value float32
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "123.45")
	assert.NoError(t, err)
	assert.Equals(t, float32(123.45), value)
}

func TestAssignFromString_Float64_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value float64
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "678.90")
	assert.NoError(t, err)
	assert.Equals(t, float64(678.90), value)
}

func TestAssignFromString_BoolTrue_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value bool
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "true")
	assert.NoError(t, err)
	assert.Equals(t, true, value)
}

func TestAssignFromString_BoolFalse_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value bool
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "false")
	assert.NoError(t, err)
	assert.Equals(t, false, value)
}

func TestAssignFromString_Bool1_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value bool
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "1")
	assert.NoError(t, err)
	assert.Equals(t, true, value)
}

func TestAssignFromString_Bool0_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value bool
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "0")
	assert.NoError(t, err)
	assert.Equals(t, false, value)
}

func TestAssignFromString_Complex64_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value complex64
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "(1+2i)")
	assert.NoError(t, err)
	assert.Equals(t, complex64(1+2i), value)
}

func TestAssignFromString_Complex128_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value complex128
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "(3+4i)")
	assert.NoError(t, err)
	assert.Equals(t, complex128(3+4i), value)
}

func TestAssignFromString_Struct_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value testInternalStruct
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `{"value":"nested"}`)
	assert.NoError(t, err)
	assert.Equals(t, testInternalStruct{Value: "nested"}, value)
}

func TestAssignFromString_StructPtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value *testInternalStruct
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `{"value":"nestedPtr"}`)
	assert.NoError(t, err)
	assert.Equals(t, testInternalStruct{Value: "nestedPtr"}, *value)
}

func TestAssignFromString_Map_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value map[string]testInternalStruct
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `{"key1":{"value":"value1"}}`)
	assert.NoError(t, err)
	assert.Equals(t, map[string]testInternalStruct{"key1": {Value: "value1"}}, value)
}

func TestAssignFromString_MapPtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value *map[string]testInternalStruct
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `{"key1":{"value":"valuePtr"}}`)
	assert.NoError(t, err)
	assert.Equals(t, map[string]testInternalStruct{"key1": {Value: "valuePtr"}}, *value)
}

func TestAssignFromString_Slice_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value []string
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `["one", "two", "three"]`)
	assert.NoError(t, err)
	assert.Equals(t, []string{"one", "two", "three"}, value)
}

func TestAssignFromString_SlicePtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value *[]string
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `["a", "b"]`)
	assert.NoError(t, err)
	assert.Equals(t, []string{"a", "b"}, *value)
}

func TestAssignFromString_SliceInt_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value []int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `[1, 2, 3]`)
	assert.NoError(t, err)
	assert.Equals(t, []int{1, 2, 3}, value)
}

func TestAssignFromString_SliceIntPtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value []*int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `[1, 2, 3]`)
	assert.NoError(t, err)
	assert.Equals(t, []*int{ptr.Of(1), ptr.Of(2), ptr.Of(3)}, value)
}

func TestAssignFromString_SliceFloat_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value []float64
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `[1.1, 2.2, 3.3]`)
	assert.NoError(t, err)
	assert.Equals(t, []float64{1.1, 2.2, 3.3}, value)
}

func TestAssignFromString_SliceBool_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value []bool
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `[true, false, true]`)
	assert.NoError(t, err)
	assert.Equals(t, []bool{true, false, true}, value)
}

func TestAssignFromString_SliceStruct_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value []testInternalStruct
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `[{"value":"nested1"}, {"value":"nested2"}]`)
	assert.NoError(t, err)
	assert.Equals(t, []testInternalStruct{{Value: "nested1"}, {Value: "nested2"}}, value)
}

func TestAssignFromString_Array_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value [3]int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `[1, 2, 3]`)
	assert.NoError(t, err)
	assert.Equals(t, [3]int{1, 2, 3}, value)
}

func TestAssignFromString_ArrayPtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value *[2]string
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `["a", "b"]`)
	assert.NoError(t, err)
	assert.Equals(t, [2]string{"a", "b"}, *value)
}

func TestAssignFromString_TextUnmarshaler_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value textUnmarshaler
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "custom text")
	assert.NoError(t, err)
	assert.Equals(t, textUnmarshaler{Value: "custom text"}, value)
}

func TestAssignFromString_TextUnmarshalerPtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value *textUnmarshaler
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "custom text ptr")
	assert.NoError(t, err)
	assert.Equals(t, textUnmarshaler{Value: "custom text ptr"}, *value)
}

func TestAssignFromString_Time_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	const setValue = "2024-01-01T12:34:56Z"
	expectedTime, err := time.Parse(time.RFC3339, setValue)
	assert.NoError(t, err)
	var value time.Time
	err = reflection.AssignFromString(reflect.ValueOf(&value).Elem(), setValue)
	assert.NoError(t, err)
	assert.Equals(t, expectedTime, value)
}

func TestAssignFromString_TimePtr_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	const setValue = "2024-01-01T12:34:56Z"
	expectedTime, err := time.Parse(time.RFC3339, setValue)
	assert.NoError(t, err)
	var value *time.Time
	err = reflection.AssignFromString(reflect.ValueOf(&value).Elem(), setValue)
	assert.NoError(t, err)
	assert.Equals(t, expectedTime, *value)
}

func TestAssignFromString_IntParseError_ReturnsError(t *testing.T) {
	t.Parallel()
	var value int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "not an integer")
	assert.ErrorPart(t, err, "strconv.ParseInt")
}

func TestAssignFromString_IntOverflow_ReturnsError(t *testing.T) {
	t.Parallel()
	var value int8
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "999")
	assert.ErrorPart(t, err, "strconv.ParseInt")
}

func TestAssignFromString_UintNegative_ReturnsError(t *testing.T) {
	t.Parallel()
	var value uint
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "-123")
	assert.ErrorPart(t, err, "strconv.ParseUint")
}

func TestAssignFromString_FloatParseError_ReturnsError(t *testing.T) {
	t.Parallel()
	var value float64
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "not a float")
	assert.ErrorPart(t, err, "strconv.ParseFloat")
}

func TestAssignFromString_BoolParseError_ReturnsError(t *testing.T) {
	t.Parallel()
	var value bool
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "2")
	assert.ErrorPart(t, err, "strconv.ParseBool")
}

func TestAssignFromString_ComplexParseError_ReturnsError(t *testing.T) {
	t.Parallel()
	var value complex128
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "not a complex")
	assert.ErrorPart(t, err, "strconv.ParseComplex")
}

func TestAssignFromString_JSONUnmarshalError_ReturnsError(t *testing.T) {
	t.Parallel()
	var value testInternalStruct
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "not a json object")
	assert.ErrorPart(t, err, "json unmarshal error")
}

func TestAssignFromString_TextUnmarshalError_ReturnsError(t *testing.T) {
	t.Parallel()
	var value failingTextUnmarshaler
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "any value")
	assert.ErrorPart(t, err, "text unmarshal error")
}

func TestAssignFromString_Chan_ReturnsError(t *testing.T) {
	t.Parallel()
	value := make(chan int)
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "test")
	assert.ErrorPart(t, err, "unsupported type")
}

func TestAssignFromString_Func_ReturnsError(t *testing.T) {
	t.Parallel()
	var value func()
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "test")
	assert.ErrorPart(t, err, "unsupported type")
}

func TestAssignFromString_Interface_ReturnsError(t *testing.T) {
	t.Parallel()
	var value any
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "test")
	assert.ErrorPart(t, err, "unsupported type")
}

func TestAssignFromString_UnsafePointer_ReturnsError(t *testing.T) {
	t.Parallel()
	var value uintptr
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "test")
	assert.ErrorPart(t, err, "unsupported type")
}

func TestAssignFromString_SliceInvalidJSON_ReturnsError(t *testing.T) {
	t.Parallel()
	var value []int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), `["one", "two"]`)
	assert.ErrorPart(t, err, "json unmarshal error")
}

func TestAssignFromString_MapInvalidJSON_ReturnsError(t *testing.T) {
	t.Parallel()
	var value map[string]int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "not a json object")
	assert.ErrorPart(t, err, "json unmarshal error")
}

func TestAssignFromString_NestedPointer_AssignsCorrectly(t *testing.T) {
	t.Parallel()
	var value **int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "789")
	assert.NoError(t, err)
	assert.Equals(t, 789, **value)
}

func TestAssignFromString_NestedPointerError_ReturnsError(t *testing.T) {
	t.Parallel()
	var value **int
	err := reflection.AssignFromString(reflect.ValueOf(&value).Elem(), "not an int")
	assert.ErrorPart(t, err, "strconv.ParseInt")
}
