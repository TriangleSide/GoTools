package validation

import (
	"fmt"
	"reflect"
	"sync"
)

// CallbackResult carries a validator outcome and directs validation flow.
type CallbackResult struct {
	fieldErrors []*FieldError
	stop        bool
	newValues   []reflect.Value
	pass        bool
}

// NewCallbackResult provides a blank result for validators to populate.
func NewCallbackResult() *CallbackResult {
	return &CallbackResult{}
}

// AddFieldError appends a field error to the result.
func (c *CallbackResult) AddFieldError(fieldError *FieldError) *CallbackResult {
	c.fieldErrors = append(c.fieldErrors, fieldError)
	return c
}

// StopValidation signals that remaining validators should be skipped for the field.
func (c *CallbackResult) StopValidation() *CallbackResult {
	c.stop = true
	return c
}

// AddValue queues additional values for the remaining validators in the tag.
// For example, dive adds each element so later validators apply to them.
func (c *CallbackResult) AddValue(val reflect.Value) *CallbackResult {
	c.newValues = append(c.newValues, val)
	return c
}

// PassValidation signals that validation passed and should continue to the next validator.
func (c *CallbackResult) PassValidation() *CallbackResult {
	c.pass = true
	return c
}

// Callback executes a validator and returns how validation should proceed.
type Callback func(*CallbackParameters) (*CallbackResult, error)

var (
	// registeredValidations stores validator callbacks by name for tag evaluation.
	registeredValidations = sync.Map{}
)

// CallbackParameters holds context for a validator callback, including struct data when available.
type CallbackParameters struct {
	// Validator is the name of the validator being executed.
	Validator Validator
	// IsStructValidation reports whether validation is running against a struct field.
	IsStructValidation bool
	// StructValue holds the parent struct when validating a struct field.
	StructValue reflect.Value
	// StructFieldName names the struct field being validated.
	StructFieldName string
	// Value is the value currently being validated.
	Value reflect.Value
	// Parameters carries the validator instruction string after the name.
	Parameters string
}

// MustRegisterValidator registers a validator callback and panics on duplicates.
func MustRegisterValidator(name Validator, callback Callback) {
	_, alreadyExists := registeredValidations.LoadOrStore(string(name), callback)
	if alreadyExists {
		panic(fmt.Errorf("validation named %s already exists", name))
	}
}
