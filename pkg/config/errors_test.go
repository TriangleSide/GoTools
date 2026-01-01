package config_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestFieldAssignmentError_Error_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying cause")
	err := &config.FieldAssignmentError{
		FieldName: "TestField",
		Value:     "testValue",
		Err:       cause,
	}
	assert.Equals(t, err.Error(), "failed to assign value testValue to field TestField: underlying cause")
}

func TestFieldAssignmentError_Unwrap_NilReceiver_ReturnsNil(t *testing.T) {
	t.Parallel()
	var err *config.FieldAssignmentError
	assert.Nil(t, err.Unwrap())
}

func TestFieldAssignmentError_Unwrap_NilErr_ReturnsNil(t *testing.T) {
	t.Parallel()
	err := &config.FieldAssignmentError{
		FieldName: "TestField",
		Value:     "testValue",
		Err:       nil,
	}
	assert.Nil(t, err.Unwrap())
}

func TestFieldAssignmentError_Unwrap_ReturnsUnderlyingError(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying cause")
	err := &config.FieldAssignmentError{
		FieldName: "TestField",
		Value:     "testValue",
		Err:       cause,
	}
	assert.Equals(t, err.Unwrap(), cause)
}

func TestFieldAssignmentError_ErrorsIs_FindsUnderlyingError(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying cause")
	err := &config.FieldAssignmentError{
		FieldName: "TestField",
		Value:     "testValue",
		Err:       cause,
	}
	assert.True(t, errors.Is(err, cause))
}

func TestFieldAssignmentError_ErrorsAs_ExtractsFieldAssignmentError(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying cause")
	fieldErr := &config.FieldAssignmentError{
		FieldName: "TestField",
		Value:     "testValue",
		Err:       cause,
	}
	wrapped := fmt.Errorf("wrapped: %w", fieldErr)

	var extracted *config.FieldAssignmentError
	assert.True(t, errors.As(wrapped, &extracted))
	assert.Equals(t, extracted.FieldName, "TestField")
	assert.Equals(t, extracted.Value, "testValue")
	assert.Equals(t, extracted.Err, cause)
}

func TestFieldAssignmentError_DefaultValueAssignment_ReturnsFieldAssignmentError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value *int `config:"ENV" config_default:"NOT_AN_INT"`
	}
	_, err := config.Process[testStruct]()
	assert.NotNil(t, err)

	var fieldErr *config.FieldAssignmentError
	assert.True(t, errors.As(err, &fieldErr))
	assert.Equals(t, fieldErr.FieldName, "Value")
	assert.Equals(t, fieldErr.Value, "NOT_AN_INT")
}

func TestFieldAssignmentError_RegularValueAssignment_ReturnsFieldAssignmentError(t *testing.T) {
	const EnvName = "VALUE"
	type testStruct struct {
		Value *int `config:"ENV"`
	}

	t.Setenv(EnvName, "NOT_AN_INT")
	_, err := config.Process[testStruct]()
	assert.NotNil(t, err)

	var fieldErr *config.FieldAssignmentError
	assert.True(t, errors.As(err, &fieldErr))
	assert.Equals(t, fieldErr.FieldName, "Value")
	assert.Equals(t, fieldErr.Value, "NOT_AN_INT")
}

func TestNilSourceFuncError_Error_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	err := &config.NilSourceFuncError{ProcessorName: "TEST_PROCESSOR"}
	assert.Equals(t, err.Error(), "processor \"TEST_PROCESSOR\" requires a non-nil sourcing function")
}

func TestNilSourceFuncError_ErrorsAs_ExtractsNilSourceFuncError(t *testing.T) {
	t.Parallel()
	nilErr := &config.NilSourceFuncError{ProcessorName: "TEST_PROCESSOR"}
	wrapped := fmt.Errorf("wrapped: %w", nilErr)

	var extracted *config.NilSourceFuncError
	assert.True(t, errors.As(wrapped, &extracted))
	assert.Equals(t, extracted.ProcessorName, "TEST_PROCESSOR")
}

func TestNilSourceFuncError_ErrorsIs_MatchesSameInstance(t *testing.T) {
	t.Parallel()
	err := &config.NilSourceFuncError{ProcessorName: "TEST_PROCESSOR"}
	wrapped := fmt.Errorf("wrapped: %w", err)
	assert.True(t, errors.Is(wrapped, err))
}

func TestProcessorAlreadyRegisteredError_Error_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	err := &config.ProcessorAlreadyRegisteredError{ProcessorName: "TEST_PROCESSOR"}
	assert.Equals(t, err.Error(), "processor with name \"TEST_PROCESSOR\" already registered")
}

func TestProcessorAlreadyRegisteredError_ErrorsAs_ExtractsProcessorAlreadyRegisteredError(t *testing.T) {
	t.Parallel()
	regErr := &config.ProcessorAlreadyRegisteredError{ProcessorName: "TEST_PROCESSOR"}
	wrapped := fmt.Errorf("wrapped: %w", regErr)

	var extracted *config.ProcessorAlreadyRegisteredError
	assert.True(t, errors.As(wrapped, &extracted))
	assert.Equals(t, extracted.ProcessorName, "TEST_PROCESSOR")
}

func TestProcessorAlreadyRegisteredError_ErrorsIs_MatchesSameInstance(t *testing.T) {
	t.Parallel()
	err := &config.ProcessorAlreadyRegisteredError{ProcessorName: "TEST_PROCESSOR"}
	wrapped := fmt.Errorf("wrapped: %w", err)
	assert.True(t, errors.Is(wrapped, err))
}
