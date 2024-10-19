package validation

import (
	"errors"
	"reflect"
)

const (
	DiveValidatorName Validator = "dive"
)

// init registers the validator.
func init() {
	MustRegisterValidator(DiveValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		value := params.Value
		if !DereferenceValue(&value) {
			return result.WithError(NewViolation(DiveValidatorName, params, DefaultDeferenceErrorMessage))
		}

		if value.Kind() != reflect.Slice {
			return result.WithError(errors.New("the dive validator only accepts slice values"))
		}

		if value.Len() == 0 {
			return result.WithStop()
		}

		for i := 0; i < value.Len(); i++ {
			result.AddValue(value.Index(i))
		}
		return result
	})
}
