package validation

const (
	RequiredValidatorName Validator = "required"
)

// init registers the validator.
func init() {
	MustRegisterValidator(RequiredValidatorName, required)
}

// required check if the value is a zero value for its type.
func required(params *CallbackParameters) error {
	value := params.Value
	if ValueIsNil(value) {
		return NewViolation(RequiredValidatorName, params, defaultNilErrorMessage)
	}
	DereferenceValue(&value)

	if value.IsZero() {
		return NewViolation(RequiredValidatorName, params, "the value is the zero-value")
	}

	return nil
}
