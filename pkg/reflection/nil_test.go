package reflection_test

import (
	"reflect"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/ptr"
	"github.com/TriangleSide/GoBase/pkg/reflection"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
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
}
