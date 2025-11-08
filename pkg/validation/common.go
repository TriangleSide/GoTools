package validation

import (
	"errors"
	"reflect"

	"github.com/TriangleSide/GoTools/pkg/reflection"
)

// dereferenceAndNilCheck is used to get the base type and ensure it's not nil.
func dereferenceAndNilCheck(value reflect.Value) (reflect.Value, error) {
	dereferenced := reflection.Dereference(value)
	if reflection.IsNil(dereferenced) {
		return reflect.Value{}, errors.New("the value is nil")
	}
	return dereferenced, nil
}
