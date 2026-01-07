package reflection

import (
	"reflect"
)

var (
	// valueSupportedNilKinds lists the kinds of reflect.Value that can be nil.
	valueSupportedNilKinds = map[reflect.Kind]struct{}{
		reflect.Chan:          {},
		reflect.Func:          {},
		reflect.Interface:     {},
		reflect.Map:           {},
		reflect.Ptr:           {},
		reflect.Slice:         {},
		reflect.UnsafePointer: {},
	}
)

// Nillable returns true if the kind supports the IsNil function.
func Nillable(kind reflect.Kind) bool {
	_, ok := valueSupportedNilKinds[kind]
	return ok
}

// IsNil returns true if the value is nil.
func IsNil(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	if !Nillable(value.Kind()) {
		return false
	}

	return value.IsNil()
}
