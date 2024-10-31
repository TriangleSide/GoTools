package validation

import (
	"fmt"
	"reflect"
	"sync"
)

// CallbackResult instructs the validation on how to proceed after the validator is complete.
type CallbackResult struct {
	err       error
	stop      bool
	newValues []reflect.Value
}

// NewCallbackResult instantiates a CallbackResult.
func NewCallbackResult() *CallbackResult {
	return &CallbackResult{
		err:       nil,
		stop:      false,
		newValues: nil,
	}
}

// WithError sets the error on the CallbackResult.
func (c *CallbackResult) WithError(err error) *CallbackResult {
	c.err = err
	return c
}

// WithStop sets the stop value in the CallbackResult.
func (c *CallbackResult) WithStop() *CallbackResult {
	c.stop = true
	return c
}

// AddValue adds a new value in the CallbackResult.
func (c *CallbackResult) AddValue(val reflect.Value) *CallbackResult {
	c.newValues = append(c.newValues, val)
	return c
}

// Callback checks a value against the instructions for the validator.
type Callback func(*CallbackParameters) *CallbackResult

var (
	// registeredValidations is a map of validator name to Callback.
	registeredValidations = sync.Map{}
)

// CallbackParameters are the parameters sent to the validation callback.
// Struct fields are only set on Struct validation.
type CallbackParameters struct {
	Validator          Validator
	IsStructValidation bool
	StructValue        reflect.Value
	StructFieldName    string
	Value              reflect.Value
	Parameters         string
}

// MustRegisterValidator sets the callback for a validator.
func MustRegisterValidator(name Validator, callback Callback) {
	_, alreadyExists := registeredValidations.LoadOrStore(string(name), callback)
	if alreadyExists {
		panic(fmt.Sprintf("Validation named %s already exists.", name))
	}
}
