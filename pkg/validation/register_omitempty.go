package validation

const (
	// OmitemptyValidatorName is the name of the validator that skips subsequent validators if the value is empty.
	OmitemptyValidatorName Validator = "omitempty"
)

// init registers the omitempty validator that stops validation if the value is empty or zero.
func init() {
	MustRegisterValidator(OmitemptyValidatorName, func(params *CallbackParameters) *CallbackResult {
		if result := required(params); result != nil {
			return NewCallbackResult().StopValidation()
		}
		return nil
	})
}
