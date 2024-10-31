package validation

import (
	"errors"
	"reflect"

	"github.com/TriangleSide/GoBase/pkg/reflection"
)

// DereferenceAndNilCheck is used to get the base type and ensure it's not nil.
func DereferenceAndNilCheck(value reflect.Value) (reflect.Value, error) {
	dereferenced, err := reflection.Dereference(value)
	if err != nil {
		return reflect.Value{}, err
	}
	if reflection.IsNil(dereferenced) {
		return reflect.Value{}, errors.New("the value is nil")
	}
	return dereferenced, nil
}
