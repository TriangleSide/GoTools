package validation_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func init() {
	validation.MustRegisterValidator(
		"registry_test_registered_can_be_used_in_var",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return validation.NewCallbackResult().PassValidation(), nil
		})

	validation.MustRegisterValidator(
		"registry_test_duplicate_panics",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return validation.NewCallbackResult().PassValidation(), nil
		})

	validation.MustRegisterValidator(
		"registry_test_empty_callback_result",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return validation.NewCallbackResult(), nil
		})

	validation.MustRegisterValidator(
		"registry_test_with_error_plain",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return nil, errors.New("some error")
		})

	validation.MustRegisterValidator(
		"registry_test_with_error_field_error",
		func(parameters *validation.CallbackParameters) (*validation.CallbackResult, error) {
			fieldErr := validation.NewFieldError(parameters, errors.New("some field error"))
			return validation.NewCallbackResult().AddFieldError(fieldErr), nil
		})

	validation.MustRegisterValidator(
		"registry_test_stop_skips_remaining_first",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return validation.NewCallbackResult().StopValidation(), nil
		})
	validation.MustRegisterValidator(
		"registry_test_stop_skips_remaining_second",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			panic(errors.New("should not be called"))
		})

	validation.MustRegisterValidator(
		"registry_test_stop_malformed_remaining_first",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return validation.NewCallbackResult().StopValidation(), nil
		})

	validation.MustRegisterValidator(
		"registry_test_callback_pass_continues_first",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return validation.NewCallbackResult().PassValidation(), nil
		})
	validation.MustRegisterValidator(
		"registry_test_callback_pass_continues_second",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return nil, errors.New("second validator error")
		})

	validation.MustRegisterValidator(
		"registry_test_callback_returns_nil_error",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			var result *validation.CallbackResult
			return result, nil
		})

	validation.MustRegisterValidator(
		"registry_test_add_value_validates_remaining",
		func(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
			if params.Value.Kind() != reflect.Slice && params.Value.Kind() != reflect.Array {
				return nil, fmt.Errorf("expected slice or array but got %s", params.Value.Kind())
			}

			result := validation.NewCallbackResult()
			for i := range params.Value.Len() {
				result.AddValue(params.Value.Index(i))
			}
			return result, nil
		})

	validation.MustRegisterValidator(
		"registry_test_add_value_no_remaining_instructions",
		func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
			return validation.NewCallbackResult().AddValue(reflect.ValueOf("anything")), nil
		})

	validation.MustRegisterValidator(
		"registry_test_add_value_validates_elements_only_add",
		func(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
			if params.Value.Kind() != reflect.Slice && params.Value.Kind() != reflect.Array {
				return nil, fmt.Errorf("expected slice or array but got %s", params.Value.Kind())
			}

			result := validation.NewCallbackResult()
			for i := range params.Value.Len() {
				result.AddValue(params.Value.Index(i))
			}
			return result, nil
		})
	validation.MustRegisterValidator(
		"registry_test_add_value_validates_elements_only_expect_int",
		func(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
			if params.Value.Kind() != reflect.Int {
				return nil, fmt.Errorf("expected int but got %s", params.Value.Kind())
			}
			return validation.NewCallbackResult().PassValidation(), nil
		})
}

func TestMustRegisterValidator_RegisteredValidator_CanBeUsedInVar(t *testing.T) {
	t.Parallel()

	err := validation.Var("anything", "registry_test_registered_can_be_used_in_var")
	assert.NoError(t, err)
}

func TestMustRegisterValidator_DuplicateName_Panics(t *testing.T) {
	t.Parallel()

	assert.PanicPart(t, func() {
		validation.MustRegisterValidator(
			"registry_test_duplicate_panics",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult().PassValidation(), nil
			})
	}, "already exists")
}

func TestNewCallbackResult_EmptyResult_ReturnsIncorrectlyFilledError(t *testing.T) {
	t.Parallel()

	err := validation.Var("anything", "registry_test_empty_callback_result")
	assert.ErrorPart(t, err, "callback response is not correctly filled")
}

func TestCallbackResult_WithError_PropagatesError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		validatorName string
		expectedError string
		expectedAs    func(err error) bool
	}

	testCases := []testCase{
		{
			name:          "when a callback returns a normal error it should be returned directly",
			validatorName: "registry_test_with_error_plain",
			expectedError: "some error",
		},
		{
			name:          "when a callback returns a field error it should be joined",
			validatorName: "registry_test_with_error_field_error",
			expectedError: "some field error",
			expectedAs: func(err error) bool {
				var fieldErr *validation.FieldError
				return errors.As(err, &fieldErr)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.Var("anything", testCase.validatorName)
			assert.Error(t, err)
			assert.ErrorPart(t, err, testCase.expectedError)
			if testCase.expectedAs != nil {
				assert.True(t, testCase.expectedAs(err))
			}
		})
	}
}

func TestCallbackResult_WithStop_SkipsRemainingValidators(t *testing.T) {
	t.Parallel()

	validators := "registry_test_stop_skips_remaining_first" +
		validation.ValidatorsSep + "registry_test_stop_skips_remaining_second"
	err := validation.Var("anything", validators)
	assert.NoError(t, err)
}

func TestCallbackResult_WithStop_MalformedRemainingInstruction_ReturnsError(t *testing.T) {
	t.Parallel()

	validators := "registry_test_stop_malformed_remaining_first" +
		validation.ValidatorsSep + "malformed=1=2"
	err := validation.Var("anything", validators)
	assert.ErrorPart(t, err, "malformed validator and instruction")
}

func TestCallback_PassValidation_ContinuesToNextValidator(t *testing.T) {
	t.Parallel()

	validators := "registry_test_callback_pass_continues_first" +
		validation.ValidatorsSep + "registry_test_callback_pass_continues_second"
	err := validation.Var("anything", validators)
	assert.ErrorPart(t, err, "second validator error")
}

func TestCallback_ReturnsNil_ReturnsError(t *testing.T) {
	t.Parallel()

	err := validation.Var("anything", "registry_test_callback_returns_nil_error")
	assert.ErrorPart(t, err, "callback returned nil result")
}

func TestCallbackResult_AddValue_ValidatesRemainingInstructionsAgainstNewValues(t *testing.T) {
	t.Parallel()

	instructions := "registry_test_add_value_validates_remaining" +
		validation.ValidatorsSep + string(validation.RequiredValidatorName)
	err := validation.Var([]int{1, 0, 2}, instructions)
	assert.Error(t, err)

	var fieldErr *validation.FieldError
	assert.True(t, errors.As(err, &fieldErr))
	assert.ErrorPart(t, err, "zero-value")
}

func TestCallbackResult_AddValue_NoRemainingInstructions_ReturnsEmptyInstructionsError(t *testing.T) {
	t.Parallel()

	err := validation.Var("anything", "registry_test_add_value_no_remaining_instructions")
	assert.ErrorPart(t, err, "empty validate instructions")
}

func TestCallbackResult_AddValue_ValidatesRestAgainstElementsOnly(t *testing.T) {
	t.Parallel()

	validators := "registry_test_add_value_validates_elements_only_add" +
		validation.ValidatorsSep + "registry_test_add_value_validates_elements_only_expect_int"
	err := validation.Var([]int{1, 2, 3}, validators)
	assert.NoError(t, err)
}
