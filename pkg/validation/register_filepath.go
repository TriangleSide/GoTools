package validation

import (
	"fmt"
	"os"
	"reflect"
)

const (
	FilepathValidatorName Validator = "filepath"
)

// init registers the validator.
func init() {
	MustRegisterValidator(FilepathValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		value := params.Value
		if !DereferenceValue(&value) {
			return result.WithError(NewViolation(FilepathValidatorName, params, defaultDeferenceErrorMessage))
		}

		if value.Kind() != reflect.String {
			return result.WithError(fmt.Errorf("the value must be a string for the %s validator", FilepathValidatorName))
		}

		if _, err := os.Stat(value.String()); err != nil {
			return result.WithError(NewViolation(FilepathValidatorName, params, fmt.Sprintf("the file '%s' is not accessible", value)))
		}

		return nil
	})
}
