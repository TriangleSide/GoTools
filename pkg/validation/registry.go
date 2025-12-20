package validation

import (
	"fmt"
	"reflect"
	"sync"
)

// CallbackResult carries a validator outcome and directs validation flow.
type CallbackResult struct {
	err       error
	stop      bool
	newValues []reflect.Value
}

// NewCallbackResult provides a blank result for validators to populate.
func NewCallbackResult() *CallbackResult {
	return &CallbackResult{
		err:       nil,
		stop:      false,
		newValues: nil,
	}
}

// WithError records a validation failure for the current validator.
func (c *CallbackResult) WithError(err error) *CallbackResult {
	c.err = err
	return c
}

// WithStop signals that remaining validators should be skipped.
func (c *CallbackResult) WithStop() *CallbackResult {
	c.stop = true
	return c
}

// AddValue queues additional values for the remaining validators in the tag.
// For example, dive adds each element so later validators apply to them.
func (c *CallbackResult) AddValue(val reflect.Value) *CallbackResult {
	c.newValues = append(c.newValues, val)
	return c
}

// Callback executes a validator and returns how validation should proceed.
type Callback func(*CallbackParameters) *CallbackResult

var (
	// registeredValidations stores validator callbacks by name for tag evaluation.
	registeredValidations = sync.Map{}
)

// CallbackParameters holds context for a validator callback, including struct data when available.
type CallbackParameters struct {
	Validator          Validator
	IsStructValidation bool
	StructValue        reflect.Value
	StructFieldName    string
	Value              reflect.Value
	Parameters         string
}

// MustRegisterValidator registers a validator callback and panics on duplicates.
func MustRegisterValidator(name Validator, callback Callback) {
	_, alreadyExists := registeredValidations.LoadOrStore(string(name), callback)
	if alreadyExists {
		panic(fmt.Errorf("validation named %s already exists", name))
	}
}
