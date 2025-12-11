package reflection_test

import (
	"reflect"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestDereference_NilPassed_DoesNothing(t *testing.T) {
	t.Parallel()
	invalidValue := reflect.ValueOf(nil)
	dereferenced := reflection.Dereference(invalidValue)
	assert.Equals(t, invalidValue, dereferenced)
}

func TestDereference_IntPassed_DoesNothing(t *testing.T) {
	t.Parallel()
	intValue := reflect.ValueOf(0)
	assert.Equals(t, intValue.Kind(), reflect.Int)
	dereferenced := reflection.Dereference(intValue)
	assert.Equals(t, intValue, dereferenced)
	assert.Equals(t, dereferenced.Kind(), reflect.Int)
}

func TestDereference_NilPtrToInt_DoesNothing(t *testing.T) {
	t.Parallel()
	var nilPtr *int = nil
	value := reflect.ValueOf(nilPtr)
	assert.Equals(t, value.Kind(), reflect.Ptr)
	dereferenced := reflection.Dereference(value)
	assert.False(t, dereferenced.IsValid())
}

func TestDereference_NilMap_DoesNothing(t *testing.T) {
	t.Parallel()
	var nilMap map[string]string = nil
	value := reflect.ValueOf(nilMap)
	assert.Equals(t, value.Kind(), reflect.Map)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Map)
}

func TestDereference_PointerChainOfInt_ReturnsInitialInteger(t *testing.T) {
	t.Parallel()
	value := reflect.ValueOf(ptr.Of(ptr.Of(ptr.Of(1))))
	assert.Equals(t, value.Kind(), reflect.Ptr)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Int)
	assert.Equals(t, dereferenced.Int(), int64(1))
}

func TestDereference_ZeroValueReflectValue_DoesNothing(t *testing.T) {
	t.Parallel()
	var zero reflect.Value
	dereferenced := reflection.Dereference(zero)
	assert.False(t, dereferenced.IsValid())
}

func TestDereference_PointerToNilPointer_ReturnsNilPointerValue(t *testing.T) {
	t.Parallel()
	var nilPtr *int = nil
	ptrToNil := &nilPtr
	value := reflect.ValueOf(ptrToNil)
	assert.Equals(t, value.Kind(), reflect.Ptr)
	dereferenced := reflection.Dereference(value)
	assert.True(t, dereferenced.IsValid())
	assert.Equals(t, dereferenced.Kind(), reflect.Ptr)
	assert.True(t, reflection.IsNil(dereferenced))
}

func TestDereference_PointerToInterfaceContainingInt_ReturnsInt(t *testing.T) {
	t.Parallel()
	var iface any = 42
	value := reflect.ValueOf(&iface)
	assert.Equals(t, value.Kind(), reflect.Ptr)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Int)
	assert.Equals(t, dereferenced.Int(), int64(42))
}

func TestDereference_PointerToNilInterface_ReturnsNilInterfaceValue(t *testing.T) {
	t.Parallel()
	var nilIface any = nil
	value := reflect.ValueOf(&nilIface)
	assert.Equals(t, value.Kind(), reflect.Ptr)
	dereferenced := reflection.Dereference(value)
	assert.True(t, dereferenced.IsValid())
	assert.Equals(t, dereferenced.Kind(), reflect.Interface)
	assert.True(t, reflection.IsNil(dereferenced))
}

func TestDereference_PointerToNestedInterfaces_ReturnsUnderlyingValue(t *testing.T) {
	t.Parallel()
	var inner any = 42
	outer := inner
	value := reflect.ValueOf(&outer)
	assert.Equals(t, value.Kind(), reflect.Ptr)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Int)
	assert.Equals(t, dereferenced.Int(), int64(42))
}

func TestDereference_NilSlice_DoesNothing(t *testing.T) {
	t.Parallel()
	var nilSlice []int = nil
	value := reflect.ValueOf(nilSlice)
	assert.Equals(t, value.Kind(), reflect.Slice)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Slice)
	assert.True(t, reflection.IsNil(dereferenced))
}

func TestDereference_NilChannel_DoesNothing(t *testing.T) {
	t.Parallel()
	var nilChan chan int = nil
	value := reflect.ValueOf(nilChan)
	assert.Equals(t, value.Kind(), reflect.Chan)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Chan)
	assert.True(t, reflection.IsNil(dereferenced))
}

func TestDereference_NilFunc_DoesNothing(t *testing.T) {
	t.Parallel()
	var nilFunc func() = nil
	value := reflect.ValueOf(nilFunc)
	assert.Equals(t, value.Kind(), reflect.Func)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Func)
	assert.True(t, reflection.IsNil(dereferenced))
}

func TestDereference_PointerToPointerToInt_ReturnsInt(t *testing.T) {
	t.Parallel()
	x := 100
	ptrToPtr := ptr.Of(&x)
	value := reflect.ValueOf(ptrToPtr)
	assert.Equals(t, value.Kind(), reflect.Ptr)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Int)
	assert.Equals(t, dereferenced.Int(), int64(100))
}

func TestDereference_InterfaceContainingPointer_ReturnsUnderlyingValue(t *testing.T) {
	t.Parallel()
	x := 55
	var iface any = &x
	value := reflect.ValueOf(&iface)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Int)
	assert.Equals(t, dereferenced.Int(), int64(55))
}

func TestDereference_InterfaceContainingNilPointer_ReturnsNilPointerValue(t *testing.T) {
	t.Parallel()
	var nilPtr *int = nil
	var iface any = nilPtr
	value := reflect.ValueOf(&iface)
	dereferenced := reflection.Dereference(value)
	assert.True(t, dereferenced.IsValid())
	assert.Equals(t, dereferenced.Kind(), reflect.Ptr)
	assert.True(t, reflection.IsNil(dereferenced))
}

func TestDereference_StringPassed_DoesNothing(t *testing.T) {
	t.Parallel()
	strValue := reflect.ValueOf("hello")
	assert.Equals(t, strValue.Kind(), reflect.String)
	dereferenced := reflection.Dereference(strValue)
	assert.Equals(t, strValue, dereferenced)
	assert.Equals(t, dereferenced.Kind(), reflect.String)
}

func TestDereference_StructPassed_DoesNothing(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field int
	}
	structValue := reflect.ValueOf(testStruct{Field: 42})
	assert.Equals(t, structValue.Kind(), reflect.Struct)
	dereferenced := reflection.Dereference(structValue)
	assert.Equals(t, structValue, dereferenced)
	assert.Equals(t, dereferenced.Kind(), reflect.Struct)
}

func TestDereference_NonNilMap_DoesNothing(t *testing.T) {
	t.Parallel()
	nonNilMap := map[string]string{"key": "value"}
	value := reflect.ValueOf(nonNilMap)
	assert.Equals(t, value.Kind(), reflect.Map)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Map)
	assert.False(t, reflection.IsNil(dereferenced))
}

func TestDereference_NonNilSlice_DoesNothing(t *testing.T) {
	t.Parallel()
	nonNilSlice := []int{1, 2, 3}
	value := reflect.ValueOf(nonNilSlice)
	assert.Equals(t, value.Kind(), reflect.Slice)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Slice)
	assert.False(t, reflection.IsNil(dereferenced))
}

func TestDereference_NonNilChannel_DoesNothing(t *testing.T) {
	t.Parallel()
	nonNilChan := make(chan int)
	value := reflect.ValueOf(nonNilChan)
	assert.Equals(t, value.Kind(), reflect.Chan)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Chan)
	assert.False(t, reflection.IsNil(dereferenced))
}

func TestDereference_NonNilFunc_DoesNothing(t *testing.T) {
	t.Parallel()
	nonNilFunc := func() {}
	value := reflect.ValueOf(nonNilFunc)
	assert.Equals(t, value.Kind(), reflect.Func)
	dereferenced := reflection.Dereference(value)
	assert.Equals(t, dereferenced.Kind(), reflect.Func)
	assert.False(t, reflection.IsNil(dereferenced))
}
