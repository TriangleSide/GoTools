package validation

const (
	RequiredValidatorName Validator = "required"
)

// init registers the validator.
func init() {
	MustRegisterValidator(RequiredValidatorName, func(params *CallbackParameters) *CallbackResult {
		return required(RequiredValidatorName, params)
	})
}

// required check if the value is a zero value for its type.
func required(validator Validator, params *CallbackParameters) *CallbackResult {
	result := NewCallbackResult()

	value := params.Value
	if !DereferenceValue(&value) {
		return result.WithError(NewViolation(validator, params, DefaultDeferenceErrorMessage))
	}

	if value.IsZero() {
		return result.WithError(NewViolation(validator, params, "the value is the zero-value"))
	}

	return nil
}
