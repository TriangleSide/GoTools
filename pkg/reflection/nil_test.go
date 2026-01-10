package reflection_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/TriangleSide/go-toolkit/pkg/ptr"
	"github.com/TriangleSide/go-toolkit/pkg/reflection"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
)

func TestIsNil_InvalidReflectValue_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, reflection.IsNil(reflect.ValueOf(nil)))
}

func TestIsNil_PtrToZero_ReturnsFalse(t *testing.T) {
	t.Parallel()
	value := ptr.Of(0)
	assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NilPtr_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var value *int
	assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NilMap_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var value map[string]string
	assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_Integer_ReturnsFalse(t *testing.T) {
	t.Parallel()
	var value int32
	assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NilUnsafePointer_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var value unsafe.Pointer
	assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NilSlice_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var value []int
	assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NonNilSlice_ReturnsFalse(t *testing.T) {
	t.Parallel()
	value := []int{}
	assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NilChannel_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var value chan int
	assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NonNilChannel_ReturnsFalse(t *testing.T) {
	t.Parallel()
	value := make(chan int)
	assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NilFunc_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var value func()
	assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NonNilFunc_ReturnsFalse(t *testing.T) {
	t.Parallel()
	value := func() {}
	assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_NilInterface_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var value any
	v := reflect.ValueOf(&value).Elem()
	assert.True(t, reflection.IsNil(v))
}

func TestIsNil_NonNilInterface_ReturnsFalse(t *testing.T) {
	t.Parallel()
	var value any = 42
	v := reflect.ValueOf(&value).Elem()
	assert.False(t, reflection.IsNil(v))
}

func TestIsNil_NonNilMap_ReturnsFalse(t *testing.T) {
	t.Parallel()
	value := make(map[string]string)
	assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_ZeroValueReflectValue_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var zero reflect.Value
	assert.True(t, reflection.IsNil(zero))
}

func TestIsNil_String_ReturnsFalse(t *testing.T) {
	t.Parallel()
	value := "test"
	assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestIsNil_Struct_ReturnsFalse(t *testing.T) {
	t.Parallel()
	type testStruct struct{ X int }
	value := testStruct{X: 1}
	assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
}

func TestNillable_Invalid_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Invalid))
}

func TestNillable_Bool_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Bool))
}

func TestNillable_Int_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Int))
}

func TestNillable_Int8_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Int8))
}

func TestNillable_Int16_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Int16))
}

func TestNillable_Int32_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Int32))
}

func TestNillable_Int64_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Int64))
}

func TestNillable_Uint_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Uint))
}

func TestNillable_Uint8_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Uint8))
}

func TestNillable_Uint16_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Uint16))
}

func TestNillable_Uint32_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Uint32))
}

func TestNillable_Uint64_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Uint64))
}

func TestNillable_Uintptr_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Uintptr))
}

func TestNillable_Float32_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Float32))
}

func TestNillable_Float64_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Float64))
}

func TestNillable_Complex64_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Complex64))
}

func TestNillable_Complex128_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Complex128))
}

func TestNillable_Array_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Array))
}

func TestNillable_Chan_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, reflection.Nillable(reflect.Chan))
}

func TestNillable_Func_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, reflection.Nillable(reflect.Func))
}

func TestNillable_Interface_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, reflection.Nillable(reflect.Interface))
}

func TestNillable_Map_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, reflection.Nillable(reflect.Map))
}

func TestNillable_Ptr_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, reflection.Nillable(reflect.Ptr))
}

func TestNillable_Slice_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, reflection.Nillable(reflect.Slice))
}

func TestNillable_String_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.String))
}

func TestNillable_Struct_ReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, reflection.Nillable(reflect.Struct))
}

func TestNillable_UnsafePointer_ReturnsTrue(t *testing.T) {
	t.Parallel()
	assert.True(t, reflection.Nillable(reflect.UnsafePointer))
}
