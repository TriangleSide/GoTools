package validation

import "github.com/TriangleSide/GoTools/pkg/reflection"

const (
	// OmitemptyValidatorName is the name of the validator that skips subsequent validators if the value is empty.
	OmitemptyValidatorName Validator = "omitempty"
)

// init registers the omitempty validator that stops validation if the value is empty or zero.
func init() {
	MustRegisterValidator(OmitemptyValidatorName, func(params *CallbackParameters) (*CallbackResult, error) {
		value := reflection.Dereference(params.Value)
		if reflection.IsNil(value) {
			return NewCallbackResult().StopValidation(), nil
		}

		if value.IsZero() {
			return NewCallbackResult().StopValidation(), nil
		}

		return NewCallbackResult().PassValidation(), nil
	})
}
