package reflection

import (
	"reflect"
)

// DereferenceNil is returned Dereference is called with nil pointers.
type DereferenceNil struct{}

// Error ensures DereferenceNil implements the error interface.
func (d *DereferenceNil) Error() string {
	return "found nil while dereferencing"
}

// Dereference returns the value after all pointers and interfaces.
// Is case of types that can be nil but aren't pointers, like maps or slices, it does nothing.
func Dereference(value reflect.Value) (reflect.Value, error) {
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		if IsNil(value) {
			return reflect.Value{}, &DereferenceNil{}
		}
		value = value.Elem()
	}

	return value, nil
}
