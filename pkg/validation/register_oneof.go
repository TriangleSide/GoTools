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
	MustRegisterValidator(OneOfValidatorName, func(params *CallbackParameters) error {
		if strings.TrimSpace(params.Parameters) == "" {
			return errors.New("no parameters provided")
		}
		allowedValues := strings.Fields(params.Parameters)

		value := params.Value
		if ValueIsNil(value) {
			return NewViolation(OneOfValidatorName, params, defaultNilErrorMessage)
		}
		DereferenceValue(&value)

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

		return NewViolation(OneOfValidatorName, params, fmt.Sprintf("the value '%s' is not one of the allowed values %v", valueStr, allowedValues))
	})
}
