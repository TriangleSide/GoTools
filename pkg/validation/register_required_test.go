package validation_test

import (
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestRequiredValidator_NonZeroInteger_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(42, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonEmptyString_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonZeroFloat_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(3.14, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonEmptySlice_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 2, 3}, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_EmptySlice_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{}, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_PointerToEmptySlice_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of([]int{}), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonEmptyMap_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(map[string]int{"a": 1}, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_EmptyMap_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(map[string]int{}, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_PointerToEmptyMap_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(map[string]int{}), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_PointerToNonZeroValue_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(1), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_PointerToPointerToNonZeroValue_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(ptr.Of(1)), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_ArrayWithNonZeroElement_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var([1]int{1}, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_StructWithNonZeroField_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(struct{ A int }{A: 1}, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_InterfaceHoldingNonZeroValue_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(any("non-empty"), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_BooleanTrue_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(true, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonNilChannel_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(make(chan int), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonNilFunction_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(func() {}, "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonZeroComplexNumber_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(complex(1, 1), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonZeroUintptr_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(uintptr(12345), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_NonZeroRune_PassesValidation(t *testing.T) {
	t.Parallel()
	err := validation.Var(rune('a'), "required")
	assert.NoError(t, err)
}

func TestRequiredValidator_ZeroInteger_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_EmptyString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_ZeroFloat_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(0.0, "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_NilPointer_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*int)(nil), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_PointerToZeroValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(0), "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_NilSlice_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(([]int)(nil), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_PointerToNilSlice_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(([]int)(nil)), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_NilMap_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((map[string]int)(nil), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_PointerToNilMap_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of((map[string]int)(nil)), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_NilPointerToPointer_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((**int)(nil), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_PointerToPointerToZeroValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(ptr.Of(0)), "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_ArrayWithZeroElement_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([1]int{0}, "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_StructWithZeroFields_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(struct{ A int }{}, "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_InterfaceHoldingZeroValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any(""), "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_NilInterface_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any(nil), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_BooleanFalse_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(false, "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_NilChannel_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((chan int)(nil), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_NilFunction_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((func())(nil), "required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestRequiredValidator_ZeroComplexNumber_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(complex(0, 0), "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_ZeroUintptr_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uintptr(0), "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_ZeroRune_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(rune(0), "required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestRequiredValidator_DiveWithValidSlice_PassesValidation(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		value any
		rule  string
	}

	testCases := []testCase{
		{
			name:  "slice with non-zero elements",
			value: []int{1, 2, 3},
			rule:  "dive,required",
		},
		{
			name:  "empty slice",
			value: []int{},
			rule:  "dive,required",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.Var(testCase.value, testCase.rule)
			assert.NoError(t, err)
		})
	}
}

func TestRequiredValidator_DiveWithInvalidSlice_ReturnsError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		value             any
		rule              string
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name:              "slice with zero element",
			value:             []int{1, 0, 3},
			rule:              "dive,required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "slice with nil pointer element",
			value:             []*int{ptr.Of(1), nil, ptr.Of(3)},
			rule:              "dive,required",
			expectedErrorPart: "value is nil",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.Var(testCase.value, testCase.rule)
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestRequiredValidator_ConcurrentValidation_PassesConsistently(t *testing.T) {
	t.Parallel()

	const goroutineCount = 50
	errorsCh := make(chan error, goroutineCount)

	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)
	for range goroutineCount {
		go func() {
			defer waitGroup.Done()
			errorsCh <- validation.Var(1, "required")
		}()
	}
	waitGroup.Wait()
	close(errorsCh)

	for err := range errorsCh {
		assert.NoError(t, err)
	}
}
