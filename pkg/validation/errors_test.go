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

func TestNewViolations_ReturnsEmptyViolations(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	assert.NotNil(t, violations)
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolations_NilViolations_DoesNotPanic(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	violations.AddViolations(nil)
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolations_EmptyViolations_RemainsEmpty(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	other := validation.NewViolations()
	violations.AddViolations(other)
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolations_NonEmptyViolations_AppendsViolations(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	other := validation.NewViolations()
	other.AddViolation(validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("test message")))
	violations.AddViolations(other)
	assert.NotNil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolation_NilViolation_DoesNotAdd(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	violations.AddViolation(nil)
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolation_NonNilViolation_AddsViolation(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	violations.AddViolation(validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("test message")))
	assert.NotNil(t, violations.NilIfEmpty())
}

func TestViolations_NilIfEmpty_EmptyViolations_ReturnsNil(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_NilIfEmpty_NonEmptyViolations_ReturnsViolations(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	violations.AddViolation(validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("test message")))
	assert.NotNil(t, violations.NilIfEmpty())
}

func TestViolations_Error_SingleViolation_ReturnsSingleMessage(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	violations.AddViolation(validation.NewViolation(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("test message")))
	assert.Equals(t, violations.Error(), "validation failed with validator 'test' because test message")
}

func TestViolations_Error_MultipleViolations_ReturnsJoinedMessages(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolations()
	violations.AddViolation(validation.NewViolation(&validation.CallbackParameters{
		Validator:          "first",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("first message")))
	violations.AddViolation(validation.NewViolation(&validation.CallbackParameters{
		Validator:          "second",
		IsStructValidation: false,
		Value:              reflect.ValueOf(2),
	}, errors.New("second message")))
	expectedErr := "validation failed with validator 'first' because first message"
	expectedErr += "; validation failed with validator 'second' because second message"
	assert.Equals(t, violations.Error(), expectedErr)
}
