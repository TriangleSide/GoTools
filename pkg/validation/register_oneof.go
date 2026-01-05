package validation

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/TriangleSide/GoTools/pkg/reflection"
)

const (
	// OneOfValidatorName is the name of the validator that checks if a value matches one of the allowed values.
	OneOfValidatorName Validator = "oneof"
)

// init registers the oneof validator that checks if a value is one of a space-separated list of allowed values.
func init() {
	MustRegisterValidator(OneOfValidatorName, func(params *CallbackParameters) (*CallbackResult, error) {
		if strings.TrimSpace(params.Parameters) == "" {
			return nil, errors.New("no parameters provided")
		}
		allowedValues := strings.Fields(params.Parameters)

		value := reflection.Dereference(params.Value)
		if reflection.IsNil(value) {
			return NewCallbackResult().AddFieldError(NewFieldError(params, errValueIsNil)), nil
		}

		var valueStr string
		if value.Kind() == reflect.String {
			valueStr = value.String()
		} else {
			valueStr = fmt.Sprintf("%v", value)
		}

		if slices.Contains(allowedValues, valueStr) {
			return NewCallbackResult().PassValidation(), nil
		}

		fieldErr := NewFieldError(params, errors.New("the value is not one of the allowed values"))
		return NewCallbackResult().AddFieldError(fieldErr), nil
	})
}
