package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/TriangleSide/GoTools/pkg/structs"
)

const (
	RequiredIfValidatorName Validator = "required_if"
)

// init registers the validator.
func init() {
	MustRegisterValidator(RequiredIfValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		if !params.IsStructValidation {
			return result.WithError(errors.New("required_if can only be used on struct fields"))
		}

		const requiredPartCount = 2
		parts := strings.Fields(params.Parameters)
		if len(parts) != requiredPartCount {
			return result.WithError(errors.New("required_if requires a field name and a value to compare"))
		}
		requiredIfFieldName := parts[0]
		requiredIfStrValue := parts[1]

		requiredFieldValue, err := structs.ValueFromName(params.StructValue.Interface(), requiredIfFieldName)
		if err != nil {
			return result.WithError(NewViolation(params, err))
		}

		// If the value to check is nil, it can never match, therefore the value is not required.
		requiredFieldValue, err = dereferenceAndNilCheck(requiredFieldValue)
		if err != nil {
			return nil
		}

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
