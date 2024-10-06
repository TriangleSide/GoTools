package ptr_test

import (
	"reflect"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/utils/ptr"
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

	t.Run("it should panic with a pointer type", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			ptr.Of[*int](nil)
		}, "type cannot be a pointer")
	})
}
