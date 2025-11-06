package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestIPAddrValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		value         any
		expectedError string
	}{
		{
			name:          "when value is a valid IPv4 address string, it should succeed",
			value:         "192.168.1.1",
			expectedError: "",
		},
		{
			name:          "when value is a valid IPv6 address string, it should succeed",
			value:         "2001:db8::1",
			expectedError: "",
		},
		{
			name:          "when value is an invalid IP address string, it should return an error",
			value:         "invalid_ip",
			expectedError: "value 'invalid_ip' could not be parsed as an IP address",
		},
		{
			name:          "when value is a non-string value, it should return an error",
			value:         12345,
			expectedError: "value must be a string",
		},
		{
			name:          "when value is a nil pointer, it should fail",
			value:         (*string)(nil),
			expectedError: "found nil while dereferencing",
		},
		{
			name:          "when value is a pointer to string with valid IP, it should succeed",
			value:         ptr.Of("8.8.8.8"),
			expectedError: "",
		},
		{
			name:          "when value is a pointer to string with invalid IP, it should return an error",
			value:         ptr.Of("not_an_ip"),
			expectedError: "value 'not_an_ip' could not be parsed as an IP address",
		},
		{
			name:          "when value is an interface wrapping a string with valid IP, it should succeed",
			value:         interface{}("127.0.0.1"),
			expectedError: "",
		},
		{
			name:          "when value is an interface wrapping a string with invalid IP, it should return an error",
			value:         interface{}("invalid_ip"),
			expectedError: "value 'invalid_ip' could not be parsed as an IP address",
		},
		{
			name:          "when value is an interface wrapping a nil pointer, it should fail",
			value:         interface{}((*string)(nil)),
			expectedError: "found nil while dereferencing",
		},
		{
			name:          "when value is an interface wrapping a non-string value, it should return an error",
			value:         interface{}(12345),
			expectedError: "the value must be a string",
		},
		{
			name:          "when value is a string with extra spaces, it should return an error",
			value:         " 192.168.1.1 ",
			expectedError: "value ' 192.168.1.1 ' could not be parsed as an IP address",
		},
		{
			name:          "when value is an empty string, it should return an error",
			value:         "",
			expectedError: "value '' could not be parsed as an IP address",
		},
		{
			name:          "when value is a malformed IP address, it should return an error",
			value:         "256.256.256.256",
			expectedError: "value '256.256.256.256' could not be parsed as an IP address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validation.Var(tc.value, "ip_addr")
			if tc.expectedError != "" {
				assert.ErrorPart(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
