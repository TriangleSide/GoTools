package reflection_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestNil(t *testing.T) {
	t.Parallel()

	t.Run("when an invalid reflect value is checked it should return true", func(t *testing.T) {
		t.Parallel()
		assert.True(t, reflection.IsNil(reflect.ValueOf(nil)))
	})

	t.Run("when a ptr to 0 is checked it should return false", func(t *testing.T) {
		t.Parallel()
		value := ptr.Of(0)
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a ptr to 0 is checked it should return true", func(t *testing.T) {
		t.Parallel()
		var value *int = nil
		assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a nil map is checked it should return true", func(t *testing.T) {
		t.Parallel()
		var value map[string]string = nil
		assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when an integer is checked it should return false", func(t *testing.T) {
		t.Parallel()
		var value int32 = 0
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a nil unsafe pointer is checked it should return true", func(t *testing.T) {
		t.Parallel()
		var value unsafe.Pointer = nil
		assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a non-nil unsafe pointer is checked it should return false", func(t *testing.T) {
		t.Parallel()
		x := 42
		value := unsafe.Pointer(&x) //nolint:gosec
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a nil slice is checked it should return true", func(t *testing.T) {
		t.Parallel()
		var value []int = nil
		assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a non-nil slice is checked it should return false", func(t *testing.T) {
		t.Parallel()
		value := []int{}
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a nil channel is checked it should return true", func(t *testing.T) {
		t.Parallel()
		var value chan int = nil
		assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a non-nil channel is checked it should return false", func(t *testing.T) {
		t.Parallel()
		value := make(chan int)
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a nil func is checked it should return true", func(t *testing.T) {
		t.Parallel()
		var value func() = nil
		assert.True(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a non-nil func is checked it should return false", func(t *testing.T) {
		t.Parallel()
		value := func() {}
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a nil interface is checked it should return true", func(t *testing.T) {
		t.Parallel()
		var value any = nil
		v := reflect.ValueOf(&value).Elem()
		assert.True(t, reflection.IsNil(v))
	})

	t.Run("when a non-nil interface is checked it should return false", func(t *testing.T) {
		t.Parallel()
		var value any = 42
		v := reflect.ValueOf(&value).Elem()
		assert.False(t, reflection.IsNil(v))
	})

	t.Run("when a non-nil map is checked it should return false", func(t *testing.T) {
		t.Parallel()
		value := make(map[string]string)
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a zero value reflect.Value is checked it should return true", func(t *testing.T) {
		t.Parallel()
		var zero reflect.Value
		assert.True(t, reflection.IsNil(zero))
	})

	t.Run("when a string is checked it should return false", func(t *testing.T) {
		t.Parallel()
		value := "test"
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})

	t.Run("when a struct is checked it should return false", func(t *testing.T) {
		t.Parallel()
		type testStruct struct{ X int }
		value := testStruct{X: 1}
		assert.False(t, reflection.IsNil(reflect.ValueOf(value)))
	})
}
