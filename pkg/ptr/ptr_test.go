package ptr_test

import (
	"reflect"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestPtr(t *testing.T) {
	t.Parallel()

	t.Run("it should convert a uint to a pointer", func(t *testing.T) {
		t.Parallel()
		ptrVal := ptr.Of[uint](123)
		assert.NotNil(t, ptrVal)
		assert.Equals(t, reflect.TypeOf(ptrVal).Elem().Kind(), reflect.Uint)
		assert.Equals(t, *ptrVal, uint(123))
	})

	t.Run("it should convert a float32 to a pointer", func(t *testing.T) {
		t.Parallel()
		ptrVal := ptr.Of[float32](123.45)
		assert.NotNil(t, ptrVal)
		assert.Equals(t, reflect.TypeOf(ptrVal).Elem().Kind(), reflect.Float32)
		assert.Equals(t, *ptrVal, float32(123.45))
	})

	t.Run("it should convert a struct to a pointer", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Value int
		}
		ptrVal := ptr.Of[testStruct](testStruct{Value: 123})
		assert.NotNil(t, ptrVal)
		assert.Equals(t, reflect.TypeOf(ptrVal).Elem().Kind(), reflect.Struct)
		assert.Equals(t, *ptrVal, testStruct{Value: 123})
	})
}

func TestIs(t *testing.T) {
	t.Parallel()

	t.Run("it should return false for a non-pointer type (int)", func(t *testing.T) {
		t.Parallel()
		assert.False(t, ptr.Is[int]())
	})

	t.Run("it should return true for a pointer type (*int)", func(t *testing.T) {
		t.Parallel()
		assert.True(t, ptr.Is[*int]())
	})

	t.Run("it should return false for a non-pointer type (struct)", func(t *testing.T) {
		t.Parallel()
		assert.False(t, ptr.Is[struct{}]())
	})

	t.Run("it should return true for a pointer type (*struct)", func(t *testing.T) {
		t.Parallel()
		type testStruct struct{}
		assert.True(t, ptr.Is[*testStruct]())
	})

	t.Run("it should return false for a non-pointer type (float64)", func(t *testing.T) {
		t.Parallel()
		assert.False(t, ptr.Is[float64]())
	})

	t.Run("it should return true for a pointer type (*float64)", func(t *testing.T) {
		t.Parallel()
		assert.True(t, ptr.Is[*float64]())
	})
}
