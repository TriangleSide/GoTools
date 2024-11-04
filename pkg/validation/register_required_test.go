package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestRequiredValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		value         interface{}
		validation    string
		expectedError string
	}{
		{
			name:          "when the value is a non-zero integer it should pass",
			value:         42,
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is zero integer it should fail",
			value:         0,
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is a non-empty string it should pass",
			value:         "hello",
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is an empty string it should fail",
			value:         "",
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is a non-zero float it should pass",
			value:         3.14,
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is zero float it should fail",
			value:         0.0,
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is a non-empty slice it should pass",
			value:         []int{1, 2, 3},
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is an empty slice it should pass",
			value:         []int{},
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is a non-empty map it should pass",
			value:         map[string]int{"a": 1},
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is an empty map it should pass",
			value:         map[string]int{},
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is a nil pointer it should fail",
			value:         (*int)(nil),
			validation:    "required",
			expectedError: "found nil while dereferencing",
		},
		{
			name:          "when the value is a pointer to zero value it should fail",
			value:         ptr.Of(0),
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is a pointer to non-zero value it should pass",
			value:         ptr.Of(1),
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is a struct with zero fields it should fail",
			value:         struct{ A int }{},
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is a struct with non-zero fields it should pass",
			value:         struct{ A int }{A: 1},
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is an interface holding zero value it should fail",
			value:         interface{}(""),
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is an interface holding non-zero value it should pass",
			value:         interface{}("non-empty"),
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the an interface is nil it should fail",
			value:         interface{}(nil),
			validation:    "required",
			expectedError: "value is nil",
		},
		{
			name:          "when using 'dive' on a slice with all non-zero elements it should pass",
			value:         []int{1, 2, 3},
			validation:    "dive,required",
			expectedError: "",
		},
		{
			name:          "when using 'dive' on a slice with zero elements it should fail",
			value:         []int{1, 0, 3},
			validation:    "dive,required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when using 'dive' on a slice with nil pointers it should fail",
			value:         []*int{ptr.Of(1), nil, ptr.Of(3)},
			validation:    "dive,required",
			expectedError: "found nil while dereferencing",
		},
		{
			name:          "when the value is boolean true it should pass",
			value:         true,
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is boolean false it should fail",
			value:         false,
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is a non-nil channel it should pass",
			value:         make(chan int),
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is a nil channel it should fail",
			value:         (chan int)(nil),
			validation:    "required",
			expectedError: "value is nil",
		},
		{
			name:          "when the value is a non-nil function it should pass",
			value:         func() {},
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is a nil function it should fail",
			value:         (func())(nil),
			validation:    "required",
			expectedError: "value is nil",
		},
		{
			name:          "when the value is a zero complex number it should fail",
			value:         complex(0, 0),
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is a non-zero complex number it should pass",
			value:         complex(1, 1),
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is zero uintptr it should fail",
			value:         uintptr(0),
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is non-zero uintptr it should pass",
			value:         uintptr(12345),
			validation:    "required",
			expectedError: "",
		},
		{
			name:          "when the value is zero rune it should fail",
			value:         rune(0),
			validation:    "required",
			expectedError: "the value is the zero-value",
		},
		{
			name:          "when the value is non-zero rune it should pass",
			value:         rune('a'),
			validation:    "required",
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validation.Var(tc.value, tc.validation)
			if tc.expectedError != "" {
				assert.ErrorPart(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
