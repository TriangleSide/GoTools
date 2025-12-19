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
	violation := validation.NewViolationError(&validation.CallbackParameters{
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
	violation := validation.NewViolationError(&validation.CallbackParameters{
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
	violation := validation.NewViolationError(&validation.CallbackParameters{
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
	violation := validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
		Parameters:         "",
	}, errors.New("test message"))
	assert.Equals(t, violation.Error(), "validation failed with validator 'test' because test message")
}

func TestNewViolations_ReturnsEmptyViolations(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	assert.NotNil(t, violations)
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolations_NilViolations_DoesNotPanic(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	violations.AddViolations(nil)
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolations_EmptyViolations_RemainsEmpty(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	other := validation.NewViolationsError()
	violations.AddViolations(other)
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolations_NonEmptyViolations_AppendsViolations(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	other := validation.NewViolationsError()
	other.AddViolation(validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("test message")))
	violations.AddViolations(other)
	assert.NotNil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolation_NilViolation_DoesNotAdd(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	violations.AddViolation(nil)
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_AddViolation_NonNilViolation_AddsViolation(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	violations.AddViolation(validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("test message")))
	assert.NotNil(t, violations.NilIfEmpty())
}

func TestViolations_NilIfEmpty_EmptyViolations_ReturnsNil(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	assert.Nil(t, violations.NilIfEmpty())
}

func TestViolations_NilIfEmpty_NonEmptyViolations_ReturnsViolations(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	violations.AddViolation(validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("test message")))
	assert.NotNil(t, violations.NilIfEmpty())
}

func TestViolations_Error_SingleViolation_ReturnsSingleMessage(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	violations.AddViolation(validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("test message")))
	assert.Equals(t, violations.Error(), "validation failed with validator 'test' because test message")
}

func TestViolations_Error_MultipleViolations_ReturnsJoinedMessages(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	violations.AddViolation(validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "first",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, errors.New("first message")))
	violations.AddViolation(validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "second",
		IsStructValidation: false,
		Value:              reflect.ValueOf(2),
	}, errors.New("second message")))
	expectedErr := "validation failed with validator 'first' because first message"
	expectedErr += "; validation failed with validator 'second' because second message"
	assert.Equals(t, violations.Error(), expectedErr)
}

func TestViolation_Unwrap_NilViolation_ReturnsNil(t *testing.T) {
	t.Parallel()
	var violation *validation.ViolationError
	assert.Nil(t, violation.Unwrap())
}

func TestViolation_Unwrap_CauseIsDiscoverableViaErrorsIs(t *testing.T) {
	t.Parallel()
	cause := errors.New("cause")
	violation := validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "test",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, cause)

	assert.True(t, errors.Is(violation, cause))
}

func TestViolations_Unwrap_NilViolations_ReturnsNil(t *testing.T) {
	t.Parallel()
	var violations *validation.ViolationsError
	assert.Nil(t, violations.Unwrap())
}

func TestViolations_Unwrap_EmptyViolations_ReturnsNil(t *testing.T) {
	t.Parallel()
	violations := validation.NewViolationsError()
	assert.Nil(t, violations.Unwrap())
}

func TestViolations_Unwrap_CausesAreDiscoverableViaErrorsIs(t *testing.T) {
	t.Parallel()
	firstCause := errors.New("first")
	secondCause := errors.New("second")

	violations := validation.NewViolationsError()
	violations.AddViolation(validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "first",
		IsStructValidation: false,
		Value:              reflect.ValueOf(1),
	}, firstCause))
	violations.AddViolation(validation.NewViolationError(&validation.CallbackParameters{
		Validator:          "second",
		IsStructValidation: false,
		Value:              reflect.ValueOf(2),
	}, secondCause))

	assert.True(t, errors.Is(violations, firstCause))
	assert.True(t, errors.Is(violations, secondCause))
}
