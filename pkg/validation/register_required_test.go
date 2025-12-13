package validation_test

import (
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestRequiredValidator_ValidValues_PassesValidation(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		value any
		rule  string
	}

	testCases := []testCase{
		{
			name:  "non-zero integer",
			value: 42,
			rule:  "required",
		},
		{
			name:  "non-empty string",
			value: "hello",
			rule:  "required",
		},
		{
			name:  "non-zero float",
			value: 3.14,
			rule:  "required",
		},
		{
			name:  "non-empty slice",
			value: []int{1, 2, 3},
			rule:  "required",
		},
		{
			name:  "empty slice",
			value: []int{},
			rule:  "required",
		},
		{
			name:  "pointer to empty slice",
			value: ptr.Of([]int{}),
			rule:  "required",
		},
		{
			name:  "non-empty map",
			value: map[string]int{"a": 1},
			rule:  "required",
		},
		{
			name:  "empty map",
			value: map[string]int{},
			rule:  "required",
		},
		{
			name:  "pointer to empty map",
			value: ptr.Of(map[string]int{}),
			rule:  "required",
		},
		{
			name:  "pointer to non-zero value",
			value: ptr.Of(1),
			rule:  "required",
		},
		{
			name:  "pointer to pointer to non-zero value",
			value: ptr.Of(ptr.Of(1)),
			rule:  "required",
		},
		{
			name:  "array with non-zero element",
			value: [1]int{1},
			rule:  "required",
		},
		{
			name:  "struct with non-zero field",
			value: struct{ A int }{A: 1},
			rule:  "required",
		},
		{
			name:  "interface holding non-zero value",
			value: any("non-empty"),
			rule:  "required",
		},
		{
			name:  "boolean true",
			value: true,
			rule:  "required",
		},
		{
			name:  "non-nil channel",
			value: make(chan int),
			rule:  "required",
		},
		{
			name:  "non-nil function",
			value: func() {},
			rule:  "required",
		},
		{
			name:  "non-zero complex number",
			value: complex(1, 1),
			rule:  "required",
		},
		{
			name:  "non-zero uintptr",
			value: uintptr(12345),
			rule:  "required",
		},
		{
			name:  "non-zero rune",
			value: rune('a'),
			rule:  "required",
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

func TestRequiredValidator_NilOrZeroValues_ReturnsError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		value             any
		rule              string
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name:              "zero integer",
			value:             0,
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "empty string",
			value:             "",
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "zero float",
			value:             0.0,
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "nil pointer",
			value:             (*int)(nil),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "pointer to zero value",
			value:             ptr.Of(0),
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "nil slice",
			value:             ([]int)(nil),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "pointer to nil slice",
			value:             ptr.Of(([]int)(nil)),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "nil map",
			value:             (map[string]int)(nil),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "pointer to nil map",
			value:             ptr.Of((map[string]int)(nil)),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "nil pointer to pointer",
			value:             (**int)(nil),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "pointer to pointer to zero value",
			value:             ptr.Of(ptr.Of(0)),
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "array with zero element",
			value:             [1]int{0},
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "struct with zero fields",
			value:             struct{ A int }{},
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "interface holding zero value",
			value:             any(""),
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "nil interface",
			value:             any(nil),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "boolean false",
			value:             false,
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "nil channel",
			value:             (chan int)(nil),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "nil function",
			value:             (func())(nil),
			rule:              "required",
			expectedErrorPart: "value is nil",
		},
		{
			name:              "zero complex number",
			value:             complex(0, 0),
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "zero uintptr",
			value:             uintptr(0),
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
		},
		{
			name:              "zero rune",
			value:             rune(0),
			rule:              "required",
			expectedErrorPart: "the value is the zero-value",
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

	waitGroup := sync.WaitGroup{}
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
