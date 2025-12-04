package reflection_test

import (
	"reflect"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestDereference(t *testing.T) {
	t.Parallel()

	t.Run("when nil is passed to Dereference it should do nothing", func(t *testing.T) {
		t.Parallel()
		invalidValue := reflect.ValueOf(nil)
		dereferenced := reflection.Dereference(invalidValue)
		assert.Equals(t, invalidValue, dereferenced)
	})

	t.Run("when an int is passed it should do nothing", func(t *testing.T) {
		t.Parallel()
		intValue := reflect.ValueOf(0)
		assert.Equals(t, intValue.Kind(), reflect.Int)
		dereferenced := reflection.Dereference(intValue)
		assert.Equals(t, intValue, dereferenced)
		assert.Equals(t, dereferenced.Kind(), reflect.Int)
	})

	t.Run("when a ptr to an int is nil it should do nothing", func(t *testing.T) {
		t.Parallel()
		var nilPtr *int = nil
		value := reflect.ValueOf(nilPtr)
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced := reflection.Dereference(value)
		assert.False(t, dereferenced.IsValid())
	})

	t.Run("when a nil map is passed it should do nothing", func(t *testing.T) {
		t.Parallel()
		var nilMap map[string]string = nil
		value := reflect.ValueOf(nilMap)
		assert.Equals(t, value.Kind(), reflect.Map)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Map)
	})

	t.Run("when a pointer chain of int is created it should return the initial integer", func(t *testing.T) {
		t.Parallel()
		value := reflect.ValueOf(ptr.Of(ptr.Of(ptr.Of(1))))
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Int)
		assert.Equals(t, dereferenced.Int(), int64(1))
	})

	t.Run("when a zero value reflect.Value is passed it should do nothing", func(t *testing.T) {
		t.Parallel()
		var zero reflect.Value
		dereferenced := reflection.Dereference(zero)
		assert.False(t, dereferenced.IsValid())
	})

	t.Run("when a pointer to a nil pointer is passed it should return a nil pointer value", func(t *testing.T) {
		t.Parallel()
		var nilPtr *int = nil
		ptrToNil := &nilPtr
		value := reflect.ValueOf(ptrToNil)
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced := reflection.Dereference(value)
		assert.True(t, dereferenced.IsValid())
		assert.Equals(t, dereferenced.Kind(), reflect.Ptr)
		assert.True(t, reflection.IsNil(dereferenced))
	})

	t.Run("when a pointer to an interface containing an int is passed it should return the int", func(t *testing.T) {
		t.Parallel()
		var iface any = 42
		value := reflect.ValueOf(&iface)
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Int)
		assert.Equals(t, dereferenced.Int(), int64(42))
	})

	t.Run("when a pointer to a nil interface is passed it should return a nil interface value", func(t *testing.T) {
		t.Parallel()
		var nilIface any = nil
		value := reflect.ValueOf(&nilIface)
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced := reflection.Dereference(value)
		assert.True(t, dereferenced.IsValid())
		assert.Equals(t, dereferenced.Kind(), reflect.Interface)
		assert.True(t, reflection.IsNil(dereferenced))
	})

	t.Run("when a pointer to nested interfaces is passed it should return the underlying value", func(t *testing.T) {
		t.Parallel()
		var inner any = 42
		outer := inner
		value := reflect.ValueOf(&outer)
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Int)
		assert.Equals(t, dereferenced.Int(), int64(42))
	})

	t.Run("when a nil slice is passed it should do nothing", func(t *testing.T) {
		t.Parallel()
		var nilSlice []int = nil
		value := reflect.ValueOf(nilSlice)
		assert.Equals(t, value.Kind(), reflect.Slice)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Slice)
		assert.True(t, reflection.IsNil(dereferenced))
	})

	t.Run("when a nil channel is passed it should do nothing", func(t *testing.T) {
		t.Parallel()
		var nilChan chan int = nil
		value := reflect.ValueOf(nilChan)
		assert.Equals(t, value.Kind(), reflect.Chan)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Chan)
		assert.True(t, reflection.IsNil(dereferenced))
	})

	t.Run("when a nil func is passed it should do nothing", func(t *testing.T) {
		t.Parallel()
		var nilFunc func() = nil
		value := reflect.ValueOf(nilFunc)
		assert.Equals(t, value.Kind(), reflect.Func)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Func)
		assert.True(t, reflection.IsNil(dereferenced))
	})

	t.Run("when a pointer to a pointer to an int is passed it should return the int", func(t *testing.T) {
		t.Parallel()
		x := 100
		ptrToPtr := ptr.Of(&x)
		value := reflect.ValueOf(ptrToPtr)
		assert.Equals(t, value.Kind(), reflect.Ptr)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Int)
		assert.Equals(t, dereferenced.Int(), int64(100))
	})

	t.Run("when an interface containing a pointer is passed it should return the underlying value", func(t *testing.T) {
		t.Parallel()
		x := 55
		var iface any = &x
		value := reflect.ValueOf(&iface)
		dereferenced := reflection.Dereference(value)
		assert.Equals(t, dereferenced.Kind(), reflect.Int)
		assert.Equals(t, dereferenced.Int(), int64(55))
	})

	t.Run("when an interface containing a nil pointer is passed it should return a nil pointer value", func(t *testing.T) {
		t.Parallel()
		var nilPtr *int = nil
		var iface any = nilPtr
		value := reflect.ValueOf(&iface)
		dereferenced := reflection.Dereference(value)
		assert.True(t, dereferenced.IsValid())
		assert.Equals(t, dereferenced.Kind(), reflect.Ptr)
		assert.True(t, reflection.IsNil(dereferenced))
	})
}
