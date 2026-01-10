package validation

import (
	"errors"

	"github.com/TriangleSide/go-toolkit/pkg/reflection"
)

const (
	// RequiredValidatorName is the name of the validator that checks if a value is non-nil and non-zero.
	RequiredValidatorName Validator = "required"
)

// init registers the required validator that ensures a value is not nil or the zero value for its type.
func init() {
	MustRegisterValidator(RequiredValidatorName, required)
}

// required checks if the value is a zero value for its type.
func required(params *CallbackParameters) (*CallbackResult, error) {
	value := reflection.Dereference(params.Value)
	if reflection.IsNil(value) {
		return NewCallbackResult().AddFieldError(NewFieldError(params, errValueIsNil)), nil
	}

	if value.IsZero() {
		return NewCallbackResult().AddFieldError(NewFieldError(params, errors.New("the value is the zero-value"))), nil
	}

	return NewCallbackResult().PassValidation(), nil
}
