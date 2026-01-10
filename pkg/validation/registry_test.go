package validation_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
	"github.com/TriangleSide/go-toolkit/pkg/test/once"
	"github.com/TriangleSide/go-toolkit/pkg/validation"
)

func TestMustRegisterValidator_RegisteredValidator_CanBeUsedInVar(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_var",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult().PassValidation(), nil
			})
	})

	err := validation.Var("anything", "rt_var")
	assert.NoError(t, err)
}

func TestMustRegisterValidator_DuplicateName_Panics(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_dup",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult().PassValidation(), nil
			})
	})

	assert.PanicPart(t, func() {
		validation.MustRegisterValidator("rt_dup",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult().PassValidation(), nil
			})
	}, "already exists")
}

func TestNewCallbackResult_EmptyResult_ReturnsIncorrectlyFilledError(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_empty",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult(), nil
			})
	})

	err := validation.Var("anything", "rt_empty")
	assert.ErrorPart(t, err, "callback response is not correctly filled")
}

func TestCallbackResult_WithError_PropagatesError(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_err",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return nil, errors.New("some error")
			})

		validation.MustRegisterValidator("rt_field_err",
			func(parameters *validation.CallbackParameters) (*validation.CallbackResult, error) {
				fieldErr := validation.NewFieldError(parameters, errors.New("some field error"))
				return validation.NewCallbackResult().AddFieldError(fieldErr), nil
			})
	})

	type testCase struct {
		name          string
		validatorName string
		expectedError string
		expectedAs    func(err error) bool
	}

	testCases := []testCase{
		{
			name:          "when a callback returns a normal error it should be returned directly",
			validatorName: "rt_err",
			expectedError: "some error",
		},
		{
			name:          "when a callback returns a field error it should be joined",
			validatorName: "rt_field_err",
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

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_stop1",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult().StopValidation(), nil
			})
		validation.MustRegisterValidator("rt_stop2",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				panic(errors.New("should not be called"))
			})
	})

	validators := "rt_stop1" +
		validation.ValidatorsSep + "rt_stop2"
	err := validation.Var("anything", validators)
	assert.NoError(t, err)
}

func TestCallbackResult_WithStop_MalformedRemainingInstruction_ReturnsError(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_malformed",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult().StopValidation(), nil
			})
	})

	validators := "rt_malformed" +
		validation.ValidatorsSep + "malformed=1=2"
	err := validation.Var("anything", validators)
	assert.ErrorPart(t, err, "malformed validator and instruction")
}

func TestCallback_PassValidation_ContinuesToNextValidator(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_pass1",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult().PassValidation(), nil
			})
		validation.MustRegisterValidator("rt_pass2",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return nil, errors.New("second validator error")
			})
	})

	validators := "rt_pass1" +
		validation.ValidatorsSep + "rt_pass2"
	err := validation.Var("anything", validators)
	assert.ErrorPart(t, err, "second validator error")
}

func TestCallback_ReturnsNil_ReturnsError(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_nil",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				var result *validation.CallbackResult
				return result, nil
			})
	})

	err := validation.Var("anything", "rt_nil")
	assert.ErrorPart(t, err, "callback returned nil result")
}

func TestCallbackResult_AddValue_ValidatesRemainingInstructionsAgainstNewValues(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_addval",
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
	})

	instructions := "rt_addval" +
		validation.ValidatorsSep + string(validation.RequiredValidatorName)
	err := validation.Var([]int{1, 0, 2}, instructions)
	assert.Error(t, err)

	var fieldErr *validation.FieldError
	assert.True(t, errors.As(err, &fieldErr))
	assert.ErrorPart(t, err, "zero-value")
}

func TestCallbackResult_AddValue_NoRemainingInstructions_ReturnsEmptyInstructionsError(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_noinstr",
			func(*validation.CallbackParameters) (*validation.CallbackResult, error) {
				return validation.NewCallbackResult().AddValue(reflect.ValueOf("anything")), nil
			})
	})

	err := validation.Var("anything", "rt_noinstr")
	assert.ErrorPart(t, err, "empty validate instructions")
}

func TestCallbackResult_AddValue_ValidatesRestAgainstElementsOnly(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterValidator("rt_elem",
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
		validation.MustRegisterValidator("rt_int",
			func(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
				if params.Value.Kind() != reflect.Int {
					return nil, fmt.Errorf("expected int but got %s", params.Value.Kind())
				}
				return validation.NewCallbackResult().PassValidation(), nil
			})
	})

	validators := "rt_elem" +
		validation.ValidatorsSep + "rt_int"
	err := validation.Var([]int{1, 2, 3}, validators)
	assert.NoError(t, err)
}
