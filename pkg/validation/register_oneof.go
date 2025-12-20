package validation

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	// OneOfValidatorName is the name of the validator that checks if a value matches one of the allowed values.
	OneOfValidatorName Validator = "oneof"
)

// init registers the oneof validator that checks if a value is one of a space-separated list of allowed values.
func init() {
	MustRegisterValidator(OneOfValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		if strings.TrimSpace(params.Parameters) == "" {
			return result.SetError(errors.New("no parameters provided"))
		}
		allowedValues := strings.Fields(params.Parameters)

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.SetError(NewFieldError(params, err))
		}

		var valueStr string
		switch value.Kind() {
		case reflect.String:
			valueStr = value.String()
		default:
			valueStr = fmt.Sprintf("%v", value)
		}

		if slices.Contains(allowedValues, valueStr) {
			return nil
		}

		return result.SetError(NewFieldError(params, errors.New("the value is not one of the allowed values")))
	})
}
