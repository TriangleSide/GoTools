package validation

import (
	"errors"
	"fmt"
	"reflect"
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

		value, err := DereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.WithError(NewViolation(params, err.Error()))
		}

		var valueStr string
		switch value.Kind() {
		case reflect.String:
			valueStr = value.String()
		default:
			valueStr = fmt.Sprintf("%v", value)
		}

		for _, allowed := range allowedValues {
			if valueStr == allowed {
				return nil
			}
		}

		return result.WithError(NewViolation(params, fmt.Sprintf("the value is not one of the allowed values")))
	})
}
