package ptr_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestOf_UintValue_ReturnsPointerToUint(t *testing.T) {
	t.Parallel()
	ptrVal := ptr.Of[uint](123)
	assert.NotNil(t, ptrVal)
	_, ok := any(ptrVal).(*uint)
	assert.True(t, ok)
	assert.Equals(t, *ptrVal, uint(123))
}

func TestOf_Float32Value_ReturnsPointerToFloat32(t *testing.T) {
	t.Parallel()
	ptrVal := ptr.Of[float32](123.45)
	assert.NotNil(t, ptrVal)
	_, ok := any(ptrVal).(*float32)
	assert.True(t, ok)
	assert.Equals(t, *ptrVal, float32(123.45))
}

func TestOf_StructValue_ReturnsPointerToStruct(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value int
	}
	ptrVal := ptr.Of[testStruct](testStruct{Value: 123})
	assert.NotNil(t, ptrVal)
	_, ok := any(ptrVal).(*testStruct)
	assert.True(t, ok)
	assert.Equals(t, *ptrVal, testStruct{Value: 123})
}

func TestIs_IntType_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, ptr.Is[int]())
}

func TestIs_PointerToIntType_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, ptr.Is[*int]())
}

func TestIs_StructType_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, ptr.Is[struct{}]())
}

func TestIs_PointerToStructType_ReturnsTrue(t *testing.T) {
	t.Parallel()
	type testStruct struct{}
	assert.True(t, ptr.Is[*testStruct]())
}

func TestIs_Float64Type_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, ptr.Is[float64]())
}

func TestIs_PointerToFloat64Type_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, ptr.Is[*float64]())
}

func TestOf_StringValue_ReturnsPointerToString(t *testing.T) {
	t.Parallel()
	ptrVal := ptr.Of[string]("test")
	assert.NotNil(t, ptrVal)
	_, ok := any(ptrVal).(*string)
	assert.True(t, ok)
	assert.Equals(t, *ptrVal, "test")
}

func TestOf_ZeroValue_ReturnsPointerToZeroValue(t *testing.T) {
	t.Parallel()
	ptrVal := ptr.Of[int](0)
	assert.NotNil(t, ptrVal)
	assert.Equals(t, *ptrVal, 0)
}

func TestOf_SliceValue_ReturnsPointerToSlice(t *testing.T) {
	t.Parallel()
	slice := []int{1, 2, 3}
	ptrVal := ptr.Of(slice)
	assert.NotNil(t, ptrVal)
	assert.Equals(t, *ptrVal, []int{1, 2, 3})
}

func TestOf_MapValue_ReturnsPointerToMap(t *testing.T) {
	t.Parallel()
	m := map[string]int{"a": 1}
	ptrVal := ptr.Of(m)
	assert.NotNil(t, ptrVal)
	assert.Equals(t, (*ptrVal)["a"], 1)
}

func TestOf_NilPointerValue_ReturnsPointerToNilPointer(t *testing.T) {
	t.Parallel()
	var nilPtr *int
	ptrVal := ptr.Of(nilPtr)
	assert.NotNil(t, ptrVal)
	assert.Nil(t, *ptrVal)
}

func TestIs_StringType_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, ptr.Is[string]())
}

func TestIs_PointerToStringType_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, ptr.Is[*string]())
}

func TestIs_SliceType_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, ptr.Is[[]int]())
}

func TestIs_PointerToSliceType_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, ptr.Is[*[]int]())
}

func TestIs_MapType_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, ptr.Is[map[string]int]())
}

func TestIs_PointerToMapType_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, ptr.Is[*map[string]int]())
}

func TestIs_DoublePointerType_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, ptr.Is[**int]())
}

func TestIs_InterfaceType_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, ptr.Is[any]())
}

func TestIs_PointerToInterfaceType_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, ptr.Is[*any]())
}
