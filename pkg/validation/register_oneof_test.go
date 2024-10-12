package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/utils/ptr"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

func TestOneOfValidator(t *testing.T) {
	type testCase struct {
		name          string
		value         any
		validation    string
		expectedError string
	}

	testCases := []testCase{
		{
			name:       "when the value matches one of the allowed strings it should pass",
			value:      "apple",
			validation: "oneof=apple banana cherry",
		},
		{
			name:          "when the value does not match any of the allowed strings it should fail",
			value:         "orange",
			validation:    "oneof=apple banana cherry",
			expectedError: "value 'orange' is not one of the allowed values [apple banana cherry]",
		},
		{
			name:       "when the value matches one of the allowed integers it should pass",
			value:      42,
			validation: "oneof=42 100 200",
		},
		{
			name:          "when the value does not match any of the allowed integers it should fail",
			value:         50,
			validation:    "oneof=42 100 200",
			expectedError: "value '50' is not one of the allowed values [42 100 200]",
		},
		{
			name:       "when the value is a pointer matching an allowed value it should pass",
			value:      ptr.Of("banana"),
			validation: "oneof=apple banana cherry",
		},
		{
			name:          "when the value is a pointer not matching any allowed value it should fail",
			value:         ptr.Of("grape"),
			validation:    "oneof=apple banana cherry",
			expectedError: "value 'grape' is not one of the allowed values [apple banana cherry]",
		},
		{
			name:          "when the value is nil it should not pass",
			value:         nil,
			validation:    "oneof=apple banana cherry",
			expectedError: "value is nil",
		},
		{
			name:       "when the value is an interface matching an allowed value it should pass",
			value:      interface{}("cherry"),
			validation: "oneof=apple banana cherry",
		},
		{
			name:          "when the value is an interface not matching any allowed value it should fail",
			value:         interface{}("pear"),
			validation:    "oneof=apple banana cherry",
			expectedError: "value 'pear' is not one of the allowed values [apple banana cherry]",
		},
		{
			name:          "when the value is an unsupported type it should convert to string and validate",
			value:         true,
			validation:    "oneof=true false",
			expectedError: "",
		},
		{
			name:          "when the allowed values list is empty it should always fail",
			value:         "anything",
			validation:    "oneof=",
			expectedError: "no parameters provided",
		},
		{
			name:       "when using 'dive' with 'oneof' and all elements match it should pass",
			value:      []string{"apple", "banana"},
			validation: "dive,oneof=apple banana cherry",
		},
		{
			name:          "when using 'dive' with 'oneof' and an element does not match it should fail",
			value:         []string{"apple", "orange"},
			validation:    "dive,oneof=apple banana cherry",
			expectedError: "value 'orange' is not one of the allowed values [apple banana cherry]",
		},
		{
			name:          "when the value is a slice and not using 'dive' it should treat slice as a single value",
			value:         []string{"apple", "banana"},
			validation:    "oneof=apple banana cherry",
			expectedError: "value '[apple banana]' is not one of the allowed values [apple banana cherry]",
		},
		{
			name:          "when the value is empty string and empty is allowed it should pass",
			value:         "",
			validation:    "oneof=",
			expectedError: "no parameters",
		},
		{
			name:          "when the value is empty string and empty is not allowed it should fail",
			value:         "",
			validation:    "oneof=apple banana cherry",
			expectedError: "value '' is not one of the allowed values [apple banana cherry]",
		},
		{
			name:       "when using 'omitempty' and the value is empty it should skip 'oneof' validation",
			value:      "",
			validation: "omitempty,oneof=apple banana cherry",
		},
		{
			name:          "when using 'omitempty' and the value is non-empty but invalid it should fail",
			value:         "orange",
			validation:    "omitempty,oneof=apple banana cherry",
			expectedError: "value 'orange' is not one of the allowed values [apple banana cherry]",
		},
		{
			name:       "when the value is an integer and matches string representation it should pass",
			value:      100,
			validation: "oneof=100 200 300",
		},
		{
			name:          "when the value is an integer and does not match string representation it should fail",
			value:         150,
			validation:    "oneof=100 200 300",
			expectedError: "value '150' is not one of the allowed values [100 200 300]",
		},
		{
			name: "when using 'dive' with pointers and all elements match it should pass",
			value: []*string{
				ptr.Of("apple"),
				ptr.Of("banana"),
			},
			validation: "dive,oneof=apple banana cherry",
		},
		{
			name: "when using 'dive' with pointers and some elements are nil it should not pass",
			value: []*string{
				ptr.Of("apple"),
				nil,
				ptr.Of("cherry"),
			},
			validation:    "dive,oneof=apple banana cherry",
			expectedError: "the value is nil",
		},
		{
			name: "when using 'dive' with pointers and an element does not match it should fail",
			value: []*string{
				ptr.Of("apple"),
				ptr.Of("grape"),
			},
			validation:    "dive,oneof=apple banana cherry",
			expectedError: "value 'grape' is not one of the allowed values [apple banana cherry]",
		},
		{
			name:          "when the allowed values contain special characters it should match correctly",
			value:         "@pple!",
			validation:    "oneof=@pple! #banana$ %cherry%",
			expectedError: "",
		},
		{
			name:          "when the allowed values are numeric strings and value matches it should pass",
			value:         "42",
			validation:    "oneof=42 100 200",
			expectedError: "",
		},
		{
			name:       "when using multiple validators and 'oneof' passes it should continue to next validator",
			value:      "apple",
			validation: "oneof=apple banana,len=5",
		},
		{
			name:          "when using multiple validators and 'oneof' fails it should not proceed to next validator",
			value:         "orange",
			validation:    "oneof=apple banana,len=6",
			expectedError: "value 'orange' is not one of the allowed values [apple banana]",
		},
		{
			name:       "when using 'oneof' with numbers and value is a float matching an allowed value it should pass",
			value:      3.14,
			validation: "oneof=3.14 2.71",
		},
		{
			name:          "when using 'oneof' with numbers and value is a float not matching allowed values it should fail",
			value:         1.62,
			validation:    "oneof=3.14 2.71",
			expectedError: "value '1.62' is not one of the allowed values [3.14 2.71]",
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
