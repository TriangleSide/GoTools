package validation

import (
	"errors"
	"reflect"

	"github.com/TriangleSide/GoTools/pkg/reflection"
)

const (
	// DiveValidatorName is the name of the validator that iterates over slice elements for validation.
	DiveValidatorName Validator = "dive"
)

// init registers the dive validator that iterates over slice elements and applies subsequent validators to each.
func init() {
	MustRegisterValidator(DiveValidatorName, func(params *CallbackParameters) (*CallbackResult, error) {
		value := reflection.Dereference(params.Value)
		if reflection.IsNil(value) {
			return NewCallbackResult().AddFieldError(NewFieldError(params, errValueIsNil)), nil
		}

		if value.Kind() != reflect.Slice {
			return nil, errors.New("the dive validator only accepts slice values")
		}

		if value.Len() == 0 {
			return NewCallbackResult().StopValidation(), nil
		}

		result := NewCallbackResult()
		for i := range value.Len() {
			result.AddValue(value.Index(i))
		}
		return result, nil
	})
}
