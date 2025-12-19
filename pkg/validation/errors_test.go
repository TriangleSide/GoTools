package validation_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestNewViolation_StructValidationWithParameters_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	violation := validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: true,
		StructValue: reflect.ValueOf(struct {
			Value int
		}{}),
		StructFieldName: "Value",
		Value:           reflect.ValueOf(1),
		Parameters:      "parameters",
	}, errors.New("test message"))
	assert.Equals(t, violation.Error(), "validation failed on field 'Value' with validator 'test'"+
		" and parameters 'parameters' because test message")
}

func TestNewViolation_StructValidationWithoutParameters_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	violation := validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: true,
		StructValue: reflect.ValueOf(struct {
			Value int
		}{}),
		StructFieldName: "Value",
		Value:           reflect.ValueOf(1),
		Parameters:      "",
	}, errors.New("test message"))
	assert.Equals(t, violation.Error(), "validation failed on field 'Value' with validator 'test' because test message")
}

func TestNewViolation_NonStructValidationWithParameters_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	violation := validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
		Parameters:         "parameters",
	}, errors.New("test message"))
	assert.Equals(t, violation.Error(), "validation failed with validator 'test' and "+
		"parameters 'parameters' because test message")
}

func TestNewViolation_NonStructValidationWithoutParameters_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	violation := validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
		Parameters:         "",
	}, errors.New("test message"))
	assert.Equals(t, violation.Error(), "validation failed with validator 'test' because test message")
}

func TestViolation_Unwrap_NilViolation_ReturnsNil(t *testing.T) {
	t.Parallel()
	var violation *validation.Violation
	assert.Nil(t, violation.Unwrap())
}

func TestViolation_Unwrap_CauseIsDiscoverableViaErrorsIs(t *testing.T) {
	t.Parallel()
	cause := errors.New("cause")
	violation := validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, cause)

	assert.True(t, errors.Is(violation, cause))
}
