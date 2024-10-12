package validation

import (
	"reflect"
)

// Validator is the name of a validate rule.
// For example: oneof, required, dive, etc...
type Validator string

var (
	// nilReflectValue is used to check if a reflect.Value is nil.
	nilReflectValue = reflect.ValueOf(nil)

	// defaultNilErrorMessage is returned if the validator encounters a nil value.
	defaultNilErrorMessage = "the value is nil"
)

// ValueIsNil returns true if the value is nil.
func ValueIsNil(value reflect.Value) bool {
	if value == nilReflectValue {
		return true
	}
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// DereferenceValue returns the value referenced by the pointer.
func DereferenceValue(value *reflect.Value) {
	if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		*value = value.Elem()
	}
}
