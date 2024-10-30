package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/ptr"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

func TestOmitemptyValidator(t *testing.T) {
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
