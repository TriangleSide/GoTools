package validation

import (
	"fmt"
	"reflect"
	"sync"
)

// Callback checks a value against the instructions for the validator.
type Callback func(*CallbackParameters) error

var (
	// registeredValidations is a map of validator name to Callback.
	registeredValidations = sync.Map{}
)

// CallbackParameters are the parameters sent to the validation callback.
// Struct fields are only set on Struct validation.
type CallbackParameters struct {
	StructValidation bool
	StructValue      reflect.Value
	StructFieldName  string
	Value            reflect.Value
	Parameters       string
}

// MustRegisterValidator sets the callback for a validator.
func MustRegisterValidator(name Validator, callback Callback) {
	_, alreadyExists := registeredValidations.LoadOrStore(string(name), callback)
	if alreadyExists {
		panic(fmt.Sprintf("Validation named %s already exists.", name))
	}
}
