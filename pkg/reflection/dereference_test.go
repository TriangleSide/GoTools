package reflection_test

import (
	"reflect"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/ptr"
	"github.com/TriangleSide/GoBase/pkg/reflection"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestDereference(t *testing.T) {
	t.Parallel()

	t.Run("when nil is passed to Dereference it should do nothing", func(t *testing.T) {
		t.Parallel()
		invalidValue := reflect.ValueOf(nil)
		dereferenced, err := reflection.Dereference(invalidValue)
		assert.NoError(t, err)
		assert.Equals(t, invalidValue, dereferenced)
	})

	t.Run("when an int is passed it should do nothing", func(t *testing.T) {
		t.Parallel()
		intValue := reflect.ValueOf(0)
		assert.Equals(t, intValue.Kind(), reflect.Int)
		dereferenced, err := reflection.Dereference(intValue)
		assert.NoError(t, err)
		assert.Equals(t, intValue, dereferenced)
		assert.Equals(t, dereferenced.Kind(), reflect.Int)
	})

	t.Run("when a ptr to an int is nil it should do nothing", func(t *testing.T) {
		t.Parallel()
		var nilPtr *int = nil
		value := reflect.ValueOf(nilPtr)
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced, err := reflection.Dereference(value)
		assert.ErrorExact(t, err, "found nil while dereferencing")
		assert.False(t, dereferenced.IsValid())
	})

	t.Run("when a nil map is passed it should do nothing", func(t *testing.T) {
		t.Parallel()
		var nilMap map[string]string = nil
		value := reflect.ValueOf(nilMap)
		assert.Equals(t, value.Kind(), reflect.Map)
		dereferenced, err := reflection.Dereference(value)
		assert.NoError(t, err)
		assert.Equals(t, dereferenced.Kind(), reflect.Map)
	})

	t.Run("when a pointer chain of int is created it should return the initial integer", func(t *testing.T) {
		t.Parallel()
		value := reflect.ValueOf(ptr.Of(ptr.Of(ptr.Of(1))))
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced, err := reflection.Dereference(value)
		assert.NoError(t, err)
		assert.Equals(t, dereferenced.Kind(), reflect.Int)
		assert.Equals(t, dereferenced.Int(), int64(1))
	})
}
