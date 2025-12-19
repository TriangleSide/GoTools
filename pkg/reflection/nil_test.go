package reflection_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
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
