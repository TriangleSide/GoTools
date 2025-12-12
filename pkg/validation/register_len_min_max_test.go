package validation_test

import (
	"fmt"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestLenMinMaxValidators(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		validator     string
		param         string
		value         any
		expectedError string
	}

	testCases := []testCase{
		{
			name:      "when the len validator has the exact length it should pass",
			validator: "len",
			param:     "5",
			value:     "hello",
		},
		{
			name:          "when the len validator has an incorrect length it should fail",
			validator:     "len",
			param:         "3",
			value:         "hello",
			expectedError: "length 5 must be exactly 3",
		},
		{
			name:      "when the min validator length equals the minimum it should pass",
			validator: "min",
			param:     "5",
			value:     "hello",
		},
		{
			name:      "when the min validator length is greater than the minimum it should pass",
			validator: "min",
			param:     "3",
			value:     "hello",
		},
		{
			name:          "when the min validator length is less than the minimum it should fail",
			validator:     "min",
			param:         "6",
			value:         "hello",
			expectedError: "length 5 must be at least 6",
		},
		{
			name:      "when the max validator length equals the maximum it should pass",
			validator: "max",
			param:     "5",
			value:     "hello",
		},
		{
			name:      "when the max validator length is less than the maximum it should pass",
			validator: "max",
			param:     "10",
			value:     "hello",
		},
		{
			name:          "when the max validator length is greater than the maximum it should fail",
			validator:     "max",
			param:         "4",
			value:         "hello",
			expectedError: "length 5 must be at most 4",
		},
		{
			name:      "when using len validator with a pointer to a string and correct length it should pass",
			validator: "len",
			param:     "5",
			value:     ptr.Of("hello"),
		},
		{
			name:          "when using len validator with a pointer to a string and incorrect length it should fail",
			validator:     "len",
			param:         "3",
			value:         ptr.Of("hello"),
			expectedError: "length 5 must be exactly 3",
		},
		{
			name:          "when using len validator with a nil pointer it should not pass",
			validator:     "len",
			param:         "5",
			value:         (*string)(nil),
			expectedError: "value is nil",
		},
		{
			name:      "when using len validator with an empty string and zero length it should pass",
			validator: "len",
			param:     "0",
			value:     "",
		},
		{
			name:          "when using len validator with an empty string and non-zero length it should fail",
			validator:     "len",
			param:         "1",
			value:         "",
			expectedError: "length 0 must be exactly 1",
		},
		{
			name:          "when the validator has an invalid parameter it should return an error",
			validator:     "len",
			param:         "abc",
			value:         "hello",
			expectedError: "invalid instruction 'abc' for len",
		},
		{
			name:          "when the validator has a non-string value it should return an error",
			validator:     "len",
			param:         "5",
			value:         12345,
			expectedError: "value must be a string for the len validator",
		},
		{
			name:      "when using validator with an interface containing a string it should pass",
			validator: "len",
			param:     "5",
			value:     any("hello"),
		},
		{
			name:          "when using validator with an interface containing a non-string it should return an error",
			validator:     "len",
			param:         "5",
			value:         any(12345),
			expectedError: "value must be a string for the len validator",
		},
		{
			name:      "when the min validator has zero length it should pass",
			validator: "min",
			param:     "0",
			value:     "",
		},
		{
			name:          "when the min validator has a negative parameter it should return an error",
			validator:     "min",
			param:         "-1",
			value:         "hello",
			expectedError: "the length parameter can't be negative",
		},
		{
			name:          "when the max validator has a negative parameter it should fail",
			validator:     "max",
			param:         "-1",
			value:         "hello",
			expectedError: "the length parameter can't be negative",
		},
		{
			name:          "when the validator has an empty parameter it should return an error",
			validator:     "len",
			param:         "",
			value:         "hello",
			expectedError: "invalid instruction '' for len",
		},
		{
			name:          "when using len validator with a Unicode string and expected rune length it should fail",
			validator:     "len",
			param:         "5",
			value:         "héllo", // 'é' is two bytes in UTF-8
			expectedError: "length 6 must be exactly 5",
		},
		{
			name:          "when the len validator has a negative parameter it should return an error",
			validator:     "len",
			param:         "-1",
			value:         "hello",
			expectedError: "the length parameter can't be negative",
		},
		{
			name:          "when the min validator has an invalid parameter it should return an error",
			validator:     "min",
			param:         "abc",
			value:         "hello",
			expectedError: "invalid instruction 'abc' for min",
		},
		{
			name:          "when the max validator has an invalid parameter it should return an error",
			validator:     "max",
			param:         "abc",
			value:         "hello",
			expectedError: "invalid instruction 'abc' for max",
		},
		{
			name:          "when the min validator has an empty parameter it should return an error",
			validator:     "min",
			param:         "",
			value:         "hello",
			expectedError: "invalid instruction '' for min",
		},
		{
			name:          "when the max validator has an empty parameter it should return an error",
			validator:     "max",
			param:         "",
			value:         "hello",
			expectedError: "invalid instruction '' for max",
		},
		{
			name:          "when using min validator with a nil pointer it should not pass",
			validator:     "min",
			param:         "5",
			value:         (*string)(nil),
			expectedError: "value is nil",
		},
		{
			name:          "when using max validator with a nil pointer it should not pass",
			validator:     "max",
			param:         "5",
			value:         (*string)(nil),
			expectedError: "value is nil",
		},
		{
			name:          "when the min validator has a non-string value it should return an error",
			validator:     "min",
			param:         "5",
			value:         12345,
			expectedError: "value must be a string for the min validator",
		},
		{
			name:          "when the max validator has a non-string value it should return an error",
			validator:     "max",
			param:         "5",
			value:         12345,
			expectedError: "value must be a string for the max validator",
		},
		{
			name:      "when using len validator with a pointer to an empty string and zero length it should pass",
			validator: "len",
			param:     "0",
			value:     ptr.Of(""),
		},
		{
			name:      "when using min validator with a pointer to a string it should pass",
			validator: "min",
			param:     "3",
			value:     ptr.Of("hello"),
		},
		{
			name:      "when using max validator with a pointer to a string it should pass",
			validator: "max",
			param:     "10",
			value:     ptr.Of("hello"),
		},
		{
			name:      "when using max validator with zero length on empty string it should pass",
			validator: "max",
			param:     "0",
			value:     "",
		},
		{
			name:          "when using max validator with zero length on non-empty string it should fail",
			validator:     "max",
			param:         "0",
			value:         "a",
			expectedError: "length 1 must be at most 0",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s (%s=%s)", tc.name, tc.validator, tc.param), func(t *testing.T) {
			t.Parallel()
			err := validation.Var(tc.value, fmt.Sprintf("%s=%s", tc.validator, tc.param))
			if tc.expectedError != "" {
				assert.ErrorPart(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
