package ptr

import (
	"reflect"
)

// Of allocates a new instance of the given type, copies the value into it, and returns it.
// This can be used as a utility to make pointers to static values.
// For example: Of[uint](123) returns a uint pointer containing the value 123.
func Of[T any](val T) *T {
	if reflect.TypeOf(val).Kind() == reflect.Ptr {
		panic("type cannot be a pointer")
	}
	valPtr := new(T)
	*valPtr = val
	return valPtr
}
