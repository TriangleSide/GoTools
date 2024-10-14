package validation

const (
	OmitemptyValidatorName Validator = "omitempty"
)

// init registers the validator.
func init() {
	MustRegisterValidator(OmitemptyValidatorName, func(params *CallbackParameters) error {
		if err := required(OmitemptyValidatorName, params); err != nil {
			return &stopValidators{}
		}
		return nil
	})
}
