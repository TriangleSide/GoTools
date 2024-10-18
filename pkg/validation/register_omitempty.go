package validation

const (
	OmitemptyValidatorName Validator = "omitempty"
)

// init registers the validator.
func init() {
	MustRegisterValidator(OmitemptyValidatorName, func(params *CallbackParameters) *CallbackResult {
		if err := required(OmitemptyValidatorName, params); err != nil {
			return NewCallbackResult().WithStop()
		}
		return nil
	})
}
