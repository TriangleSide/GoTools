package validation

const (
	RequiredValidatorName Validator = "required"
)

// init registers the validator.
func init() {
	MustRegisterValidator(RequiredValidatorName, func(params *CallbackParameters) *CallbackResult {
		return required(params)
	})
}

// required check if the value is a zero value for its type.
func required(params *CallbackParameters) *CallbackResult {
	result := NewCallbackResult()

	value, err := DereferenceAndNilCheck(params.Value)
	if err != nil {
		return result.WithError(NewViolation(params, err.Error()))
	}

	if value.IsZero() {
		return result.WithError(NewViolation(params, "the value is the zero-value"))
	}

	return nil
}
