package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestOmitemptyValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		value         any
		validation    string
		expectedError string
	}{
		{
			name:          "when the value is empty and 'omitempty' is used, it should pass and stop further validation",
			value:         "",
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is non-empty and 'omitempty' is used, it should run subsequent validators",
			value:         "test",
			validation:    "omitempty,len=4",
			expectedError: "",
		},
		{
			name:          "when the value is non-empty and subsequent validator fails, it should return an error",
			value:         "test",
			validation:    "omitempty,len=5",
			expectedError: "length 4 must be exactly 5",
		},
		{
			name:          "when the value is empty and 'omitempty' is used with 'gt=0', it should pass",
			value:         0,
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero and 'gt=0' passes, it should pass",
			value:         5,
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero and 'gt=0' fails, it should return an error",
			value:         -1,
			validation:    "omitempty,gt=0",
			expectedError: "value -1 must be greater than 0",
		},
		{
			name:          "when using 'dive' with 'omitempty' on an empty slice, it should pass",
			value:         []int{},
			validation:    "dive,omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when using 'dive' with 'omitempty' on a slice with zero values, it should pass",
			value:         []int{0, 0},
			validation:    "dive,omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when using 'dive' with 'omitempty' on a slice with valid non-zero values, it should pass",
			value:         []int{1, 2, 3},
			validation:    "dive,omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when using 'dive' with 'omitempty' on a slice with a non-zero value failing validation, it should return an error",
			value:         []int{1, -1, 3},
			validation:    "dive,omitempty,gt=0",
			expectedError: "value -1 must be greater than 0",
		},
		{
			name:          "when using 'dive' with 'omitempty' on a slice with mixed zero and invalid values, it should only validate non-zero values",
			value:         []int{0, -1, 2},
			validation:    "dive,omitempty,gt=0",
			expectedError: "value -1 must be greater than 0",
		},
		{
			name:          "when the value is nil and 'omitempty' is used, it should pass",
			value:         nil,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero and 'omitempty' is used with 'required', it should pass",
			value:         1,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is zero and 'omitempty' is used with 'required', it should pass",
			value:         0,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is an empty slice and 'omitempty' is used with 'required', it should pass",
			value:         []int{},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when using 'dive' with 'omitempty' on a slice of pointers with nil values, it should pass",
			value:         []*int{nil, nil},
			validation:    "dive,omitempty,gt=0",
			expectedError: "",
		},
		{
			name: "when using 'dive' with 'omitempty' on a slice of pointers with mixed nil and invalid values, it should validate non-nil values",
			value: []*int{
				ptr.Of(-1),
				nil,
				ptr.Of(1),
			},
			validation:    "dive,omitempty,gt=0",
			expectedError: "value -1 must be greater than 0",
		},
		{
			name:          "when the value is a non-empty string and subsequent validator passes, it should pass",
			value:         "hello",
			validation:    "omitempty,len=5",
			expectedError: "",
		},
		{
			name:          "when the value is a non-empty string and subsequent validator fails, it should return an error",
			value:         "hello",
			validation:    "omitempty,len=4",
			expectedError: "length 5 must be exactly 4",
		},
		{
			name:          "when the value is an empty map and 'omitempty' is used, it should pass",
			value:         map[string]int{},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is a non-empty map and 'omitempty' is used with passing validation, it should pass",
			value:         map[string]int{"a": 1},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is a nil map and 'omitempty' is used, it should pass",
			value:         (map[string]int)(nil),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is boolean false (zero-value) and 'omitempty' is used, it should pass",
			value:         false,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is boolean true and 'omitempty' is used, it should continue validation",
			value:         true,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is zero float and 'omitempty' is used, it should pass",
			value:         0.0,
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero float and subsequent validator passes, it should pass",
			value:         1.5,
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero float and subsequent validator fails, it should return an error",
			value:         -1.5,
			validation:    "omitempty,gt=0",
			expectedError: "value -1.5 must be greater than 0",
		},
		{
			name:          "when the value is a nil channel and 'omitempty' is used, it should pass",
			value:         (chan int)(nil),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is a non-nil channel and 'omitempty' is used, it should continue validation",
			value:         make(chan int),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is a nil function and 'omitempty' is used, it should pass",
			value:         (func())(nil),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is a non-nil function and 'omitempty' is used, it should continue validation",
			value:         func() {},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is a zero struct and 'omitempty' is used, it should pass",
			value:         struct{ A int }{},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is a non-zero struct and 'omitempty' is used, it should continue validation",
			value:         struct{ A int }{A: 1},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is a nil double pointer and 'omitempty' is used, it should pass",
			value:         (**int)(nil),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name: "when the value is a non-nil double pointer to zero and 'omitempty' is used, it should skip validation",
			value: func() **int {
				i := 0
				p := &i
				return &p
			}(),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name: "when the value is a non-nil double pointer to non-zero and subsequent validator passes, it should pass",
			value: func() **int {
				i := 5
				p := &i
				return &p
			}(),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is zero complex and 'omitempty' is used, it should pass",
			value:         complex(0, 0),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero complex and 'omitempty' is used, it should continue validation",
			value:         complex(1, 1),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "when 'omitempty' is used alone with empty value, it should pass",
			value:         "",
			validation:    "omitempty",
			expectedError: "",
		},
		{
			name:          "when 'omitempty' is used alone with non-empty value, it should pass",
			value:         "hello",
			validation:    "omitempty",
			expectedError: "",
		},
		{
			name:          "when 'omitempty' is used alone with zero int, it should pass",
			value:         0,
			validation:    "omitempty",
			expectedError: "",
		},
		{
			name:          "when 'omitempty' is used alone with non-zero int, it should pass",
			value:         42,
			validation:    "omitempty",
			expectedError: "",
		},
		{
			name:          "when the value is a pointer to empty string and 'omitempty' is used, it should pass",
			value:         ptr.Of(""),
			validation:    "omitempty,len=5",
			expectedError: "",
		},
		{
			name:          "when the value is a pointer to non-empty string and subsequent validator passes, it should pass",
			value:         ptr.Of("hello"),
			validation:    "omitempty,len=5",
			expectedError: "",
		},
		{
			name:          "when the value is a pointer to non-empty string and subsequent validator fails, it should return an error",
			value:         ptr.Of("hello"),
			validation:    "omitempty,len=4",
			expectedError: "length 5 must be exactly 4",
		},
		{
			name:          "when the value is a pointer to zero int and 'omitempty' is used, it should pass",
			value:         ptr.Of(0),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is a pointer to non-zero int and subsequent validator passes, it should pass",
			value:         ptr.Of(5),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is a pointer to non-zero int and subsequent validator fails, it should return an error",
			value:         ptr.Of(-5),
			validation:    "omitempty,gt=0",
			expectedError: "value -5 must be greater than 0",
		},
		{
			name: "when using 'dive' with 'omitempty' on a slice of structs with zero structs, it should pass",
			value: []struct{ A int }{
				{A: 0},
				{A: 0},
			},
			validation:    "dive,omitempty,required",
			expectedError: "",
		},
		{
			name: "when using 'dive' with 'omitempty' on a slice of structs with non-zero structs, it should validate",
			value: []struct{ A int }{
				{A: 1},
				{A: 2},
			},
			validation:    "dive,omitempty,required",
			expectedError: "",
		},
		{
			name:          "when the value is zero uint and 'omitempty' is used, it should pass",
			value:         uint(0),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero uint and subsequent validator passes, it should pass",
			value:         uint(5),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is zero int8 and 'omitempty' is used, it should pass",
			value:         int8(0),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero int8 and subsequent validator passes, it should pass",
			value:         int8(5),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is zero float32 and 'omitempty' is used, it should pass",
			value:         float32(0),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "when the value is non-zero float32 and subsequent validator passes, it should pass",
			value:         float32(1.5),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
	}

	for _, tc := range testCases {
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
