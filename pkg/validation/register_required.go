package validation

import "errors"

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
	value, err := dereferenceAndNilCheck(params.Value)
	if err != nil {
		return NewCallbackResult().AddFieldError(NewFieldError(params, err)), nil
	}

	if value.IsZero() {
		return NewCallbackResult().AddFieldError(NewFieldError(params, errors.New("the value is the zero-value"))), nil
	}

	return nil, nil //nolint:nilnil // nil, nil means validation passed
}
