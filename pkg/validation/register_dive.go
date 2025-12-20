package validation

import (
	"errors"
	"reflect"
)

const (
	// DiveValidatorName is the name of the validator that iterates over slice elements for validation.
	DiveValidatorName Validator = "dive"
)

// init registers the dive validator that iterates over slice elements and applies subsequent validators to each.
func init() {
	MustRegisterValidator(DiveValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.SetError(NewFieldError(params, err))
		}
		if value.Kind() != reflect.Slice {
			return result.SetError(errors.New("the dive validator only accepts slice values"))
		}

		if value.Len() == 0 {
			return result.StopValidation()
		}

		for i := range value.Len() {
			result.AddValue(value.Index(i))
		}
		return result
	})
}
