package validation

import (
	"fmt"
	"os"
	"reflect"

	"github.com/TriangleSide/GoTools/pkg/reflection"
)

const (
	// FilepathValidatorName is the name of the validator that checks if a file path is accessible.
	FilepathValidatorName Validator = "filepath"
)

// init registers the filepath validator that checks if a file path exists and is accessible.
func init() {
	MustRegisterValidator(FilepathValidatorName, func(params *CallbackParameters) (*CallbackResult, error) {
		value := reflection.Dereference(params.Value)
		if reflection.IsNil(value) {
			return NewCallbackResult().AddFieldError(NewFieldError(params, errValueIsNil)), nil
		}

		if value.Kind() != reflect.String {
			return nil, fmt.Errorf("the value must be a string for the %s validator", FilepathValidatorName)
		}

		info, _ := os.Stat(value.String())
		if info == nil {
			fieldErr := NewFieldError(params, fmt.Errorf("the file '%s' is not accessible", value))
			return NewCallbackResult().AddFieldError(fieldErr), nil
		}

		return NewCallbackResult().PassValidation(), nil
	})
}
