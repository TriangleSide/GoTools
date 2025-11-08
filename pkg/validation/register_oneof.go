package validation

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	OneOfValidatorName Validator = "oneof"
)

// init registers the validator.
func init() {
	MustRegisterValidator(OneOfValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		if strings.TrimSpace(params.Parameters) == "" {
			return result.WithError(errors.New("no parameters provided"))
		}
		allowedValues := strings.Fields(params.Parameters)

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.WithError(NewViolation(params, err))
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

		return result.WithError(NewViolation(params, errors.New("the value is not one of the allowed values")))
	})
}
