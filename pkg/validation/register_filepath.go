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
	MustRegisterValidator(FilepathValidatorName, func(params *CallbackParameters) (*CallbackResult, error) {
		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return NewCallbackResult().AddFieldError(NewFieldError(params, err)), nil
		}
		if value.Kind() != reflect.String {
			return nil, fmt.Errorf("the value must be a string for the %s validator", FilepathValidatorName)
		}

		if _, err := os.Stat(value.String()); err != nil {
			fieldErr := NewFieldError(params, fmt.Errorf("the file '%s' is not accessible", value))
			return NewCallbackResult().AddFieldError(fieldErr), nil //nolint:nilerr // returning field error
		}

		return nil, nil //nolint:nilnil // nil, nil means validation passed
	})
}
