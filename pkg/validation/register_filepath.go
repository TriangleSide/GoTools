package validation

import (
	"fmt"
	"os"
	"reflect"
)

const (
	// FilepathValidatorName is the name of the validator that checks if a file path is accessible.
	FilepathValidatorName Validator = "filepath"
)

// init registers the filepath validator that checks if a file path exists and is accessible.
func init() {
	MustRegisterValidator(FilepathValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.SetError(NewFieldError(params, err))
		}
		if value.Kind() != reflect.String {
			return result.SetError(fmt.Errorf("the value must be a string for the %s validator", FilepathValidatorName))
		}

		if _, err := os.Stat(value.String()); err != nil {
			return result.SetError(NewFieldError(params, fmt.Errorf("the file '%s' is not accessible", value)))
		}

		return nil
	})
}
