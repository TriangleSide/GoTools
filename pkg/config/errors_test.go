package config_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/structs"
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

func TestProcessorNotRegisteredError_Error_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	err := &config.ProcessorNotRegisteredError{ProcessorName: "TEST_PROCESSOR"}
	assert.Equals(t, err.Error(), "processor \"TEST_PROCESSOR\" not registered")
}

func TestProcessorNotRegisteredError_ErrorsAs_ExtractsProcessorNotRegisteredError(t *testing.T) {
	t.Parallel()
	notRegErr := &config.ProcessorNotRegisteredError{ProcessorName: "TEST_PROCESSOR"}
	wrapped := fmt.Errorf("wrapped: %w", notRegErr)

	var extracted *config.ProcessorNotRegisteredError
	assert.True(t, errors.As(wrapped, &extracted))
	assert.Equals(t, extracted.ProcessorName, "TEST_PROCESSOR")
}

func TestProcessorNotRegisteredError_ErrorsIs_MatchesSameInstance(t *testing.T) {
	t.Parallel()
	err := &config.ProcessorNotRegisteredError{ProcessorName: "TEST_PROCESSOR"}
	wrapped := fmt.Errorf("wrapped: %w", err)
	assert.True(t, errors.Is(wrapped, err))
}

func TestProcessorNotRegisteredError_Process_ReturnsProcessorNotRegisteredError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string `config:"DOES_NOT_EXIST"`
	}
	_, err := config.Process[testStruct]()
	assert.NotNil(t, err)

	var notRegErr *config.ProcessorNotRegisteredError
	assert.True(t, errors.As(err, &notRegErr))
	assert.Equals(t, notRegErr.ProcessorName, "DOES_NOT_EXIST")
}

func TestNoValueFoundError_Error_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	err := &config.NoValueFoundError{FieldName: "TestField"}
	assert.Equals(t, err.Error(), "no value found for field TestField")
}

func TestNoValueFoundError_ErrorsAs_ExtractsNoValueFoundError(t *testing.T) {
	t.Parallel()
	noValErr := &config.NoValueFoundError{FieldName: "TestField"}
	wrapped := fmt.Errorf("wrapped: %w", noValErr)

	var extracted *config.NoValueFoundError
	assert.True(t, errors.As(wrapped, &extracted))
	assert.Equals(t, extracted.FieldName, "TestField")
}

func TestNoValueFoundError_ErrorsIs_MatchesSameInstance(t *testing.T) {
	t.Parallel()
	err := &config.NoValueFoundError{FieldName: "TestField"}
	wrapped := fmt.Errorf("wrapped: %w", err)
	assert.True(t, errors.Is(wrapped, err))
}

func TestNoValueFoundError_Process_ReturnsNoValueFoundError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string `config:"ENV"`
	}
	_, err := config.Process[testStruct]()
	assert.NotNil(t, err)

	var noValErr *config.NoValueFoundError
	assert.True(t, errors.As(err, &noValErr))
	assert.Equals(t, noValErr.FieldName, "Value")
}

func TestSourceFetchError_Error_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying cause")
	err := &config.SourceFetchError{
		FieldName:     "TestField",
		ProcessorName: "TEST_PROCESSOR",
		Err:           cause,
	}
	expected := "error while fetching the value for field TestField using processor TEST_PROCESSOR: underlying cause"
	assert.Equals(t, err.Error(), expected)
}

func TestSourceFetchError_Unwrap_NilReceiver_ReturnsNil(t *testing.T) {
	t.Parallel()
	var err *config.SourceFetchError
	assert.Nil(t, err.Unwrap())
}

func TestSourceFetchError_Unwrap_NilErr_ReturnsNil(t *testing.T) {
	t.Parallel()
	err := &config.SourceFetchError{
		FieldName:     "TestField",
		ProcessorName: "TEST_PROCESSOR",
		Err:           nil,
	}
	assert.Nil(t, err.Unwrap())
}

func TestSourceFetchError_Unwrap_ReturnsUnderlyingError(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying cause")
	err := &config.SourceFetchError{
		FieldName:     "TestField",
		ProcessorName: "TEST_PROCESSOR",
		Err:           cause,
	}
	assert.Equals(t, err.Unwrap(), cause)
}

func TestSourceFetchError_ErrorsIs_FindsUnderlyingError(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying cause")
	err := &config.SourceFetchError{
		FieldName:     "TestField",
		ProcessorName: "TEST_PROCESSOR",
		Err:           cause,
	}
	assert.True(t, errors.Is(err, cause))
}

func TestSourceFetchError_ErrorsAs_ExtractsSourceFetchError(t *testing.T) {
	t.Parallel()
	cause := errors.New("underlying cause")
	fetchErr := &config.SourceFetchError{
		FieldName:     "TestField",
		ProcessorName: "TEST_PROCESSOR",
		Err:           cause,
	}
	wrapped := fmt.Errorf("wrapped: %w", fetchErr)

	var extracted *config.SourceFetchError
	assert.True(t, errors.As(wrapped, &extracted))
	assert.Equals(t, extracted.FieldName, "TestField")
	assert.Equals(t, extracted.ProcessorName, "TEST_PROCESSOR")
	assert.Equals(t, extracted.Err, cause)
}

func TestSourceFetchError_Process_ReturnsSourceFetchError(t *testing.T) {
	t.Parallel()
	fetchError := errors.New("fetch failed")
	config.MustRegisterProcessor("FAILING_SOURCE", func(string, *structs.FieldMetadata) (string, bool, error) {
		return "", false, fetchError
	})

	type testStruct struct {
		Value string `config:"FAILING_SOURCE"`
	}
	_, err := config.Process[testStruct]()
	assert.NotNil(t, err)

	var fetchErr *config.SourceFetchError
	assert.True(t, errors.As(err, &fetchErr))
	assert.Equals(t, fetchErr.FieldName, "Value")
	assert.Equals(t, fetchErr.ProcessorName, "FAILING_SOURCE")
	assert.True(t, errors.Is(fetchErr, fetchError))
}
