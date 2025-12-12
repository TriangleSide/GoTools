package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestOneOfValidator_VariousInputs_ReturnsExpectedErrors(t *testing.T) {
	t.Parallel()

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
			name:       "when there is a single allowed value and it matches it should pass",
			value:      "only",
			validation: "oneof=only",
		},
		{
			name:          "when there is a single allowed value and it does not match it should fail",
			value:         "other",
			validation:    "oneof=only",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:          "when the value does not match any of the allowed strings it should fail",
			value:         "orange",
			validation:    "oneof=apple banana cherry",
			expectedError: "value is not one of the allowed values",
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
			expectedError: "value is not one of the allowed values",
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
			expectedError: "value is not one of the allowed values",
		},
		{
			name:          "when the value is nil it should not pass",
			value:         nil,
			validation:    "oneof=apple banana cherry",
			expectedError: "value is nil",
		},
		{
			name:          "when the value is a typed nil pointer it should not pass",
			value:         (*string)(nil),
			validation:    "oneof=apple banana cherry",
			expectedError: "value is nil",
		},
		{
			name:          "when the value is a typed nil int pointer it should not pass",
			value:         (*int)(nil),
			validation:    "oneof=1 2 3",
			expectedError: "value is nil",
		},
		{
			name:       "when the value is an interface matching an allowed value it should pass",
			value:      any("cherry"),
			validation: "oneof=apple banana cherry",
		},
		{
			name:          "when the value is an interface not matching any allowed value it should fail",
			value:         any("pear"),
			validation:    "oneof=apple banana cherry",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:          "when the value is a boolean true it should convert to string and validate",
			value:         true,
			validation:    "oneof=true false",
			expectedError: "",
		},
		{
			name:          "when the value is a boolean false it should convert to string and validate",
			value:         false,
			validation:    "oneof=true false",
			expectedError: "",
		},
		{
			name:          "when the value is a boolean and does not match string representation it should fail",
			value:         true,
			validation:    "oneof=yes no",
			expectedError: "value is not one of the allowed values",
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
			expectedError: "value is not one of the allowed values",
		},
		{
			name:          "when the value is a slice and not using 'dive' it should treat slice as a single value",
			value:         []string{"apple", "banana"},
			validation:    "oneof=apple banana cherry",
			expectedError: "value is not one of the allowed values",
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
			expectedError: "value is not one of the allowed values",
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
			expectedError: "value is not one of the allowed values",
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
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when the value is a negative integer and matches it should pass",
			value:      -5,
			validation: "oneof=-5 -10 -15",
		},
		{
			name:          "when the value is a negative integer and does not match it should fail",
			value:         -7,
			validation:    "oneof=-5 -10 -15",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when the value is int8 and matches it should pass",
			value:      int8(42),
			validation: "oneof=42 100",
		},
		{
			name:          "when the value is int8 and does not match it should fail",
			value:         int8(50),
			validation:    "oneof=42 100",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when the value is int16 and matches it should pass",
			value:      int16(1000),
			validation: "oneof=1000 2000",
		},
		{
			name:       "when the value is int32 and matches it should pass",
			value:      int32(100000),
			validation: "oneof=100000 200000",
		},
		{
			name:       "when the value is int64 and matches it should pass",
			value:      int64(9223372036854775807),
			validation: "oneof=9223372036854775807 -9223372036854775808",
		},
		{
			name:       "when the value is uint and matches it should pass",
			value:      uint(42),
			validation: "oneof=42 100",
		},
		{
			name:          "when the value is uint and does not match it should fail",
			value:         uint(50),
			validation:    "oneof=42 100",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when the value is uint8 and matches it should pass",
			value:      uint8(255),
			validation: "oneof=255 128",
		},
		{
			name:       "when the value is uint16 and matches it should pass",
			value:      uint16(65535),
			validation: "oneof=65535 32768",
		},
		{
			name:       "when the value is uint32 and matches it should pass",
			value:      uint32(4294967295),
			validation: "oneof=4294967295 2147483648",
		},
		{
			name:       "when the value is uint64 and matches it should pass",
			value:      uint64(18446744073709551615),
			validation: "oneof=18446744073709551615 0",
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
			expectedError: "value is nil",
		},
		{
			name: "when using 'dive' with pointers and an element does not match it should fail",
			value: []*string{
				ptr.Of("apple"),
				ptr.Of("grape"),
			},
			validation:    "dive,oneof=apple banana cherry",
			expectedError: "value is not one of the allowed values",
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
			expectedError: "value is not one of the allowed values",
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
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when the value is a unicode string and matches it should pass",
			value:      "日本語",
			validation: "oneof=日本語 中文 한국어",
		},
		{
			name:          "when the value is a unicode string and does not match it should fail",
			value:         "español",
			validation:    "oneof=日本語 中文 한국어",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when the value contains emoji and matches it should pass",
			value:      "🎉",
			validation: "oneof=🎉 🚀 ✨",
		},
		{
			name:          "when the value contains emoji and does not match it should fail",
			value:         "🔥",
			validation:    "oneof=🎉 🚀 ✨",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when the value is float32 and matches it should pass",
			value:      float32(1.5),
			validation: "oneof=1.5 2.5",
		},
		{
			name:          "when the value is float32 and does not match it should fail",
			value:         float32(3.5),
			validation:    "oneof=1.5 2.5",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when the value is zero and zero is in the allowed list it should pass",
			value:      0,
			validation: "oneof=0 1 2",
		},
		{
			name:          "when the value is zero and zero is not in the allowed list it should fail",
			value:         0,
			validation:    "oneof=1 2 3",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when using dive with integers and all match it should pass",
			value:      []int{1, 2, 3},
			validation: "dive,oneof=1 2 3",
		},
		{
			name:          "when using dive with integers and one does not match it should fail",
			value:         []int{1, 2, 4},
			validation:    "dive,oneof=1 2 3",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when using dive with an empty slice it should pass",
			value:      []string{},
			validation: "dive,oneof=a b c",
		},
		{
			name:       "when value is a pointer to integer matching it should pass",
			value:      ptr.Of(42),
			validation: "oneof=42 100",
		},
		{
			name:          "when value is a pointer to integer not matching it should fail",
			value:         ptr.Of(50),
			validation:    "oneof=42 100",
			expectedError: "value is not one of the allowed values",
		},
		{
			name:       "when allowed values have many options and value matches last one it should pass",
			value:      "z",
			validation: "oneof=a b c d e f g h i j k l m n o p q r s t u v w x y z",
		},
		{
			name:       "when allowed values have many options and value matches first one it should pass",
			value:      "a",
			validation: "oneof=a b c d e f g h i j k l m n o p q r s t u v w x y z",
		},
		{
			name:          "when parameters contain only whitespace it should fail",
			value:         "test",
			validation:    "oneof=   ",
			expectedError: "no parameters provided",
		},
		{
			name:       "when value is a byte and matches it should pass",
			value:      byte(65),
			validation: "oneof=65 66 67",
		},
		{
			name:       "when value is a rune and matches it should pass",
			value:      rune(65),
			validation: "oneof=65 66 67",
		},
		{
			name:          "when value is an empty slice without dive it should fail",
			value:         []string{},
			validation:    "oneof=a b c",
			expectedError: "value is not one of the allowed values",
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
