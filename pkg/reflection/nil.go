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

// IsNil returns true if the value is nil.
func IsNil(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	if _, ok := valueSupportedNilKinds[value.Kind()]; !ok {
		return false
	}

	return value.IsNil()
}
