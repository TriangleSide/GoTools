package validation

import "errors"

const (
	RequiredValidatorName Validator = "required"
)

// init registers the validator.
func init() {
	MustRegisterValidator(RequiredValidatorName, func(params *CallbackParameters) *CallbackResult {
		return required(params)
	})
}

// required checks if the value is a zero value for its type.
func required(params *CallbackParameters) *CallbackResult {
	result := NewCallbackResult()

	value, err := dereferenceAndNilCheck(params.Value)
	if err != nil {
		return result.WithError(NewViolation(params, err))
	}

	if value.IsZero() {
		return result.WithError(NewViolation(params, errors.New("the value is the zero-value")))
	}

	return nil
}
