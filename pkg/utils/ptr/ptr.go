package ptr

import (
	"reflect"
)

// Of allocates a new instance of the given type, copies the value into it, and returns it.
// This can be used as a utility to make pointers to static values.
// For example: Of[uint](123) returns a uint pointer containing the value 123.
func Of[T any](val T) *T {
	valPtr := new(T)
	*valPtr = val
	return valPtr
}

// Is checks if the generic parameter is a pointer type.
// Returns true if T is a pointer, false otherwise.
func Is[T any]() bool {
	var zeroVal T
	return reflect.TypeOf(zeroVal).Kind() == reflect.Ptr
}
