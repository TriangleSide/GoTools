package validation_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestMustRegisterValidator_RegisteredValidator_CanBeUsedInVar(t *testing.T) {
	t.Parallel()

	validatorName := validation.Validator("registry_test_registered_can_be_used_in_var")
	callback := func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return validation.NewCallbackResult().PassValidation(), nil
	}
	validation.MustRegisterValidator(validatorName, callback)

	err := validation.Var("anything", string(validatorName))
	assert.NoError(t, err)
}

func TestMustRegisterValidator_DuplicateName_Panics(t *testing.T) {
	t.Parallel()

	validatorName := validation.Validator("registry_test_duplicate_panics")
	callback := func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return validation.NewCallbackResult().PassValidation(), nil
	}
	validation.MustRegisterValidator(validatorName, callback)

	assert.PanicPart(t, func() {
		validation.MustRegisterValidator(validatorName, callback)
	}, "already exists")
}

func TestNewCallbackResult_EmptyResult_ReturnsIncorrectlyFilledError(t *testing.T) {
	t.Parallel()

	validatorName := validation.Validator("registry_test_empty_callback_result")
	callback := func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return validation.NewCallbackResult(), nil
	}
	validation.MustRegisterValidator(validatorName, callback)

	err := validation.Var("anything", string(validatorName))
	assert.ErrorPart(t, err, "callback response is not correctly filled")
}

func TestCallbackResult_WithError_PropagatesError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		validatorName validation.Validator
		callback      validation.Callback
		expectedError string
		expectedAs    func(err error) bool
	}

	testCases := []testCase{
		{
			name:          "when a callback returns a normal error it should be returned directly",
			validatorName: "registry_test_with_error_plain",
			callback: func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return nil, errors.New("some error")
			},
			expectedError: "some error",
		},
		{
			name:          "when a callback returns a field error it should be joined",
			validatorName: "registry_test_with_error_field_error",
			callback: func(parameters *validation.CallbackParameters) (*validation.CallbackResult, error) {
				fieldErr := validation.NewFieldError(parameters, errors.New("some field error"))
				return validation.NewCallbackResult().AddFieldError(fieldErr), nil
			},
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

			validation.MustRegisterValidator(testCase.validatorName, testCase.callback)

			err := validation.Var("anything", string(testCase.validatorName))
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

	stopName := validation.Validator("registry_test_stop_skips_remaining_first")
	panicName := validation.Validator("registry_test_stop_skips_remaining_second")

	validation.MustRegisterValidator(stopName, func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return validation.NewCallbackResult().StopValidation(), nil
	})
	validation.MustRegisterValidator(panicName, func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		panic(errors.New("should not be called"))
	})

	err := validation.Var("anything", string(stopName)+validation.ValidatorsSep+string(panicName))
	assert.NoError(t, err)
}

func TestCallbackResult_WithStop_SkipsMalformedRemainingInstruction(t *testing.T) {
	t.Parallel()

	stopName := validation.Validator("registry_test_stop_skips_malformed_remaining_first")

	validation.MustRegisterValidator(stopName, func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return validation.NewCallbackResult().StopValidation(), nil
	})

	err := validation.Var("anything", string(stopName)+validation.ValidatorsSep+"malformed=1=2")
	assert.NoError(t, err)
}

func TestCallback_PassValidation_ContinuesToNextValidator(t *testing.T) {
	t.Parallel()

	firstName := validation.Validator("registry_test_callback_pass_continues_first")
	secondName := validation.Validator("registry_test_callback_pass_continues_second")

	passCallback := func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return validation.NewCallbackResult().PassValidation(), nil
	}
	validation.MustRegisterValidator(firstName, passCallback)
	validation.MustRegisterValidator(secondName, func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return nil, errors.New("second validator error")
	})

	err := validation.Var("anything", string(firstName)+validation.ValidatorsSep+string(secondName))
	assert.ErrorPart(t, err, "second validator error")
}

func TestCallback_ReturnsNil_ReturnsError(t *testing.T) {
	t.Parallel()

	validatorName := validation.Validator("registry_test_callback_returns_nil_error")

	nilCallback := func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return nil, nil //nolint:nilnil
	}
	validation.MustRegisterValidator(validatorName, nilCallback)

	err := validation.Var("anything", string(validatorName))
	assert.ErrorPart(t, err, "callback returned nil result")
}

func TestCallbackResult_AddValue_ValidatesRemainingInstructionsAgainstNewValues(t *testing.T) {
	t.Parallel()

	addValueName := validation.Validator("registry_test_add_value_validates_remaining")

	callback := func(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
		if params.Value.Kind() != reflect.Slice && params.Value.Kind() != reflect.Array {
			return nil, fmt.Errorf("expected slice or array but got %s", params.Value.Kind())
		}

		result := validation.NewCallbackResult()
		for i := range params.Value.Len() {
			result.AddValue(params.Value.Index(i))
		}
		return result, nil
	}
	validation.MustRegisterValidator(addValueName, callback)

	instructions := string(addValueName) + validation.ValidatorsSep + string(validation.RequiredValidatorName)
	err := validation.Var([]int{1, 0, 2}, instructions)
	assert.Error(t, err)

	var fieldErr *validation.FieldError
	assert.True(t, errors.As(err, &fieldErr))
	assert.ErrorPart(t, err, "zero-value")
}

func TestCallbackResult_AddValue_NoRemainingInstructions_ReturnsEmptyInstructionsError(t *testing.T) {
	t.Parallel()

	addValueName := validation.Validator("registry_test_add_value_no_remaining_instructions")

	callback := func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
		return validation.NewCallbackResult().AddValue(reflect.ValueOf("anything")), nil
	}
	validation.MustRegisterValidator(addValueName, callback)

	err := validation.Var("anything", string(addValueName))
	assert.ErrorPart(t, err, "empty validate instructions")
}

func TestCallbackResult_AddValue_ValidatesRestAgainstElementsOnly(t *testing.T) {
	t.Parallel()

	addValueName := validation.Validator("registry_test_add_value_validates_elements_only_add")
	expectIntName := validation.Validator("registry_test_add_value_validates_elements_only_expect_int")

	addValueCallback := func(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
		if params.Value.Kind() != reflect.Slice && params.Value.Kind() != reflect.Array {
			return nil, fmt.Errorf("expected slice or array but got %s", params.Value.Kind())
		}

		result := validation.NewCallbackResult()
		for i := range params.Value.Len() {
			result.AddValue(params.Value.Index(i))
		}
		return result, nil
	}
	validation.MustRegisterValidator(addValueName, addValueCallback)
	expectIntCallback := func(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
		if params.Value.Kind() != reflect.Int {
			return nil, fmt.Errorf("expected int but got %s", params.Value.Kind())
		}
		return validation.NewCallbackResult().PassValidation(), nil
	}
	validation.MustRegisterValidator(expectIntName, expectIntCallback)

	err := validation.Var([]int{1, 2, 3}, string(addValueName)+validation.ValidatorsSep+string(expectIntName))
	assert.NoError(t, err)
}
