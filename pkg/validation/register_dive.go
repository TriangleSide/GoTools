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
	MustRegisterValidator(DiveValidatorName, func(params *CallbackParameters) error {
		value := params.Value
		if ValueIsNil(value) {
			return NewViolation(DiveValidatorName, params, defaultNilErrorMessage)
		}
		DereferenceValue(&value)

		if value.Kind() != reflect.Slice {
			return errors.New("the dive validator only accepts slice values")
		}

		valuesToValidate := make([]reflect.Value, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			valuesToValidate = append(valuesToValidate, value.Index(i))
		}

		return &newValues{
			values: valuesToValidate,
		}
	})
}
