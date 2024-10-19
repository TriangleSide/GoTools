package validation

import (
	"reflect"
)

// Validator is the name of a validate rule.
// For example: oneof, required, dive, etc...
type Validator string

var (
	DefaultDeferenceErrorMessage = "the value could not be dereferenced"
)

// valueIsNil returns true if the value is nil.
func valueIsNil(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// DereferenceValue returns base type after pointers.
func DereferenceValue(value *reflect.Value) bool {
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		if valueIsNil(*value) {
			return false
		}
		*value = value.Elem()
	}
	return !valueIsNil(*value)
}
