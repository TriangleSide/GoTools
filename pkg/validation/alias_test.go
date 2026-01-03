package validation_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/test/once"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestMustRegisterAlias_RegisteredAlias_CanBeUsedInVar(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_single", "required")
	})

	err := validation.Var("value", "alias_test_single")
	assert.NoError(t, err)
}

func TestMustRegisterAlias_DuplicateName_Panics(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_duplicate", "required")
	})

	assert.PanicPart(t, func() {
		validation.MustRegisterAlias("alias_test_duplicate", string(validation.RequiredValidatorName))
	}, "already exists")
}

func TestMustRegisterAlias_MultipleValidators_ExpandsCorrectly(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_multi", "dive,required,gt=0")
	})

	type testCase struct {
		name              string
		value             []int
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name:  "all_values_valid",
			value: []int{1, 2, 3},
		},
		{
			name:              "contains_zero",
			value:             []int{1, 0, 3},
			expectedErrorPart: "zero-value",
		},
		{
			name:              "contains_negative",
			value:             []int{1, -5, 3},
			expectedErrorPart: "must be greater than 0",
		},
		{
			name:              "empty_slice_element",
			value:             []int{},
			expectedErrorPart: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.Var(testCase.value, "alias_test_multi")
			if testCase.expectedErrorPart == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestMustRegisterAlias_UsedInStruct_ValidatesCorrectly(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_struct", "required,gt=0")
	})

	type testCase struct {
		name              string
		value             int
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name:  "valid_value",
			value: 10,
		},
		{
			name:              "zero_value",
			value:             0,
			expectedErrorPart: "zero-value",
		},
		{
			name:              "negative_value",
			value:             -5,
			expectedErrorPart: "must be greater than 0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			type testStruct struct {
				Value int `validate:"alias_test_struct"`
			}

			err := validation.Struct(&testStruct{Value: testCase.value})
			if testCase.expectedErrorPart == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestMustRegisterAlias_WithAdditionalValidators_ProcessesAll(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_with_additional", "required")
	})

	err := validation.Var(5, "alias_test_with_additional,gt=0")
	assert.NoError(t, err)

	err = validation.Var(0, "alias_test_with_additional,gt=0")
	assert.ErrorPart(t, err, "zero-value")

	err = validation.Var(-1, "alias_test_with_additional,gt=0")
	assert.ErrorPart(t, err, "must be greater than 0")
}

func TestMustRegisterAlias_MultipleAliases_AllExpand(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_first", "required")
		validation.MustRegisterAlias("alias_test_second", "gt=0")
	})

	err := validation.Var(5, "alias_test_first,alias_test_second")
	assert.NoError(t, err)

	err = validation.Var(0, "alias_test_first,alias_test_second")
	assert.ErrorPart(t, err, "zero-value")

	err = validation.Var(-5, "alias_test_first,alias_test_second")
	assert.ErrorPart(t, err, "must be greater than 0")
}

func TestMustRegisterAlias_DiveWithAlias_ExpandsInRest(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_dive_rest", "required,gt=0")
	})

	err := validation.Var([]int{1, 2, 3}, "dive,alias_test_dive_rest")
	assert.NoError(t, err)

	err = validation.Var([]int{1, 0, 3}, "dive,alias_test_dive_rest")
	assert.ErrorPart(t, err, "zero-value")

	err = validation.Var([]int{1, -1, 3}, "dive,alias_test_dive_rest")
	assert.ErrorPart(t, err, "must be greater than 0")
}

func TestMustRegisterAlias_AliasWithParameters_WorksCorrectly(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_params", "oneof=apple banana cherry")
	})

	err := validation.Var("apple", "alias_test_params")
	assert.NoError(t, err)

	err = validation.Var("orange", "alias_test_params")
	assert.ErrorPart(t, err, "not one of the allowed values")
}

func TestMustRegisterAlias_ReturnsFieldError_WithExpandedValidator(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_field_error", "gt=10")
	})

	err := validation.Var(5, "alias_test_field_error")
	assert.Error(t, err)

	var fieldErr *validation.FieldError
	assert.True(t, errors.As(err, &fieldErr))
	assert.ErrorPart(t, err, "validator 'gt'")
}

func TestMustRegisterAlias_ConcurrentUsage_NoRaces(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_concurrent", "required,gt=0")
	})

	const workers = 32
	errs := make(chan error, workers)

	var waitGroup sync.WaitGroup
	for workerIdx := range workers {
		waitGroup.Go(func() {
			if workerIdx%2 == 0 {
				errs <- validation.Var(0, "alias_test_concurrent")
				return
			}
			errs <- validation.Var(5, "alias_test_concurrent")
		})
	}
	waitGroup.Wait()
	close(errs)

	var gotError bool
	var gotNoError bool
	for err := range errs {
		if err == nil {
			gotNoError = true
			continue
		}
		gotError = true
	}

	assert.True(t, gotError)
	assert.True(t, gotNoError)
}

func TestMustRegisterAlias_StructFieldWithAliasFieldError_ReportsFieldName(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_field_name", "required")
	})

	type testStruct struct {
		MyField string `validate:"alias_test_field_name"`
	}

	err := validation.Struct(&testStruct{MyField: ""})
	assert.ErrorPart(t, err, "field 'MyField'")
}

func TestMustRegisterAlias_EmptyExpansion_ReturnsError(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_empty", "")
	})

	err := validation.Var("value", "alias_test_empty")
	assert.ErrorPart(t, err, "validation with name '' is not registered")
}

func TestMustRegisterAlias_SliceOfStructs_ValidatesElements(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		validation.MustRegisterAlias("alias_test_slice_struct", "dive,required")
	})

	type inner struct {
		Value int `validate:"gt=0"`
	}

	err := validation.Var([]inner{{Value: 1}, {Value: 2}}, "alias_test_slice_struct")
	assert.NoError(t, err)

	err = validation.Var([]inner{{Value: 1}, {Value: 0}}, "alias_test_slice_struct")
	assert.ErrorPart(t, err, "must be greater than 0")
}
