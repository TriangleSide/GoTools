package validation

const (
	RequiredValidatorName Validator = "required"
)

// init registers the validator.
func init() {
	MustRegisterValidator(RequiredValidatorName, func(params *CallbackParameters) error {
		return required(RequiredValidatorName, params)
	})
}

// required check if the value is a zero value for its type.
func required(validator Validator, params *CallbackParameters) error {
	value := params.Value
	if ValueIsNil(value) {
		return NewViolation(validator, params, defaultNilErrorMessage)
	}
	DereferenceValue(&value)

	if value.IsZero() {
		return NewViolation(validator, params, "the value is the zero-value")
	}

	return nil
}
