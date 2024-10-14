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
	MustRegisterValidator(FilepathValidatorName, func(params *CallbackParameters) error {
		value := params.Value
		if ValueIsNil(value) {
			return NewViolation(FilepathValidatorName, params, defaultNilErrorMessage)
		}
		DereferenceValue(&value)

		if value.Kind() != reflect.String {
			return fmt.Errorf("the value must be a string for the %s validator", FilepathValidatorName)
		}

		if _, err := os.Stat(value.String()); err != nil {
			return NewViolation(FilepathValidatorName, params, fmt.Sprintf("the file '%s' is not accessible", value))
		}

		return nil
	})
}