package reflection

import (
	"reflect"
)

// Dereference returns the value after dereferencing pointers and interfaces.
// In the case of types that can be nil but aren't pointers, such as maps or slices, it does nothing.
func Dereference(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		value = value.Elem()
		if IsNil(value) {
			return value
		}
	}
	return value
}

// DereferenceType returns the base type after dereferencing all pointer indirections.
func DereferenceType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
