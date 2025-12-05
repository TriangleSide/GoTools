package validation

const (
	OmitemptyValidatorName Validator = "omitempty"
)

// init registers the validator.
func init() {
	MustRegisterValidator(OmitemptyValidatorName, func(params *CallbackParameters) *CallbackResult {
		if result := required(params); result != nil {
			return NewCallbackResult().WithStop()
		}
		return nil
	})
}
