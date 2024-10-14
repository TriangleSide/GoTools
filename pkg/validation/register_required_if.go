package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/TriangleSide/GoBase/pkg/utils/fields"
)

const (
	RequiredIfValidatorName Validator = "required_if"
)

// init registers the validator.
func init() {
	MustRegisterValidator(RequiredIfValidatorName, func(params *CallbackParameters) error {
		if !params.StructValidation {
			return errors.New("required_if can only be used on struct fields")
		}

		parts := strings.Fields(params.Parameters)
		if len(parts) != 2 {
			return errors.New("required_if requires a field name and a value to compare")
		}
		requiredIfFieldName := parts[0]
		requiredIfStrValue := parts[1]

		requiredFieldValue, err := fields.StructValueFromName(params.StructValue.Interface(), requiredIfFieldName)
		if err != nil {
			return NewViolation(RequiredIfValidatorName, params, err.Error())
		}
		// If the value to check is nil, it can never match, therefore the value is not required.
		if ValueIsNil(requiredFieldValue) {
			return nil
		}
		DereferenceValue(&requiredFieldValue)

		var requiredFieldValueStr string
		switch requiredFieldValue.Kind() {
		case reflect.String:
			requiredFieldValueStr = requiredFieldValue.String()
		default:
			requiredFieldValueStr = fmt.Sprintf("%v", requiredFieldValue.Interface())
		}

		if requiredFieldValueStr == requiredIfStrValue {
			return required(params)
		}

		return nil
	})
}