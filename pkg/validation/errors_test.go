package validation_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestNewFieldError_StructValidationWithParameters_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	fieldErr := validation.NewFieldError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: true,
		StructValue: reflect.ValueOf(struct {
			Value int
		}{}),
		StructFieldName: "Value",
		Value:           reflect.ValueOf(1),
		Parameters:      "parameters",
	}, errors.New("test message"))
	assert.Equals(t, fieldErr.Error(), "validation failed on field 'Value' with validator 'test'"+
		" and parameters 'parameters' because test message")
}

func TestNewFieldError_StructValidationWithoutParameters_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	fieldErr := validation.NewFieldError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: true,
		StructValue: reflect.ValueOf(struct {
			Value int
		}{}),
		StructFieldName: "Value",
		Value:           reflect.ValueOf(1),
		Parameters:      "",
	}, errors.New("test message"))
	assert.Equals(t, fieldErr.Error(), "validation failed on field 'Value' with validator 'test' because test message")
}

func TestNewFieldError_NonStructValidationWithParameters_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	fieldErr := validation.NewFieldError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
		Parameters:         "parameters",
	}, errors.New("test message"))
	assert.Equals(t, fieldErr.Error(), "validation failed with validator 'test' and "+
		"parameters 'parameters' because test message")
}

func TestNewFieldError_NonStructValidationWithoutParameters_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	fieldErr := validation.NewFieldError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
		Parameters:         "",
	}, errors.New("test message"))
	assert.Equals(t, fieldErr.Error(), "validation failed with validator 'test' because test message")
}

func TestFieldError_Unwrap_NilFieldError_ReturnsNil(t *testing.T) {
	t.Parallel()
	var fieldErr *validation.FieldError
	assert.Nil(t, fieldErr.Unwrap())
}

func TestFieldError_Unwrap_CauseIsDiscoverableViaErrorsIs(t *testing.T) {
	t.Parallel()
	cause := errors.New("cause")
	fieldErr := validation.NewFieldError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, cause)

	assert.True(t, errors.Is(fieldErr, cause))
}
