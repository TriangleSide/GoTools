package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestIPAddrValidator_StructField(t *testing.T) {
	t.Parallel()

	type Config struct {
		IPAddress string `validate:"ip_addr"`
	}

	t.Run("when struct field has valid IP address it should succeed", func(t *testing.T) {
		t.Parallel()
		cfg := Config{IPAddress: "192.168.1.1"}
		err := validation.Struct(&cfg)
		assert.NoError(t, err)
	})

	t.Run("when struct field has invalid IP address it should fail", func(t *testing.T) {
		t.Parallel()
		cfg := Config{IPAddress: "invalid_ip"}
		err := validation.Struct(&cfg)
		assert.ErrorPart(t, err, "could not be parsed as an IP address")
	})
}

func TestIPAddrValidator_VariousInputs_ReturnsExpectedErrors(t *testing.T) {
	t.Parallel()

	type testCaseDefinition struct {
		Name          string
		Value         any
		ExpectedError string
	}

	testCases := []testCaseDefinition{
		{
			Name:          "when value is a valid IPv4 address string, it should succeed",
			Value:         "192.168.1.1",
			ExpectedError: "",
		},
		{
			Name:          "when value is a valid IPv6 address string, it should succeed",
			Value:         "2001:db8::1",
			ExpectedError: "",
		},
		{
			Name:          "when value is an invalid IP address string, it should return an error",
			Value:         "invalid_ip",
			ExpectedError: "value 'invalid_ip' could not be parsed as an IP address",
		},
		{
			Name:          "when value is a non-string value, it should return an error",
			Value:         12345,
			ExpectedError: "value must be a string",
		},
		{
			Name:          "when value is a nil pointer, it should fail",
			Value:         (*string)(nil),
			ExpectedError: "value is nil",
		},
		{
			Name:          "when value is a pointer to string with valid IP, it should succeed",
			Value:         ptr.Of("8.8.8.8"),
			ExpectedError: "",
		},
		{
			Name:          "when value is a pointer to string with invalid IP, it should return an error",
			Value:         ptr.Of("not_an_ip"),
			ExpectedError: "value 'not_an_ip' could not be parsed as an IP address",
		},
		{
			Name:          "when value is an interface wrapping a string with valid IP, it should succeed",
			Value:         any("127.0.0.1"),
			ExpectedError: "",
		},
		{
			Name:          "when value is an interface wrapping a string with invalid IP, it should return an error",
			Value:         any("invalid_ip"),
			ExpectedError: "value 'invalid_ip' could not be parsed as an IP address",
		},
		{
			Name:          "when value is an interface wrapping a nil pointer, it should fail",
			Value:         any((*string)(nil)),
			ExpectedError: "value is nil",
		},
		{
			Name:          "when value is an interface wrapping a non-string value, it should return an error",
			Value:         any(12345),
			ExpectedError: "the value must be a string",
		},
		{
			Name:          "when value is a string with extra spaces, it should return an error",
			Value:         " 192.168.1.1 ",
			ExpectedError: "value ' 192.168.1.1 ' could not be parsed as an IP address",
		},
		{
			Name:          "when value is an empty string, it should return an error",
			Value:         "",
			ExpectedError: "value '' could not be parsed as an IP address",
		},
		{
			Name:          "when value is a malformed IP address, it should return an error",
			Value:         "256.256.256.256",
			ExpectedError: "value '256.256.256.256' could not be parsed as an IP address",
		},
		{
			Name:          "when value is IPv4 all zeros, it should succeed",
			Value:         "0.0.0.0",
			ExpectedError: "",
		},
		{
			Name:          "when value is IPv4 broadcast address, it should succeed",
			Value:         "255.255.255.255",
			ExpectedError: "",
		},
		{
			Name:          "when value is IPv6 loopback address, it should succeed",
			Value:         "::1",
			ExpectedError: "",
		},
		{
			Name:          "when value is IPv6 all zeros, it should succeed",
			Value:         "::",
			ExpectedError: "",
		},
		{
			Name:          "when value is IPv6 full form, it should succeed",
			Value:         "2001:0db8:0000:0000:0000:0000:0000:0001",
			ExpectedError: "",
		},
		{
			Name:          "when value is IPv4-mapped IPv6 address, it should succeed",
			Value:         "::ffff:192.168.1.1",
			ExpectedError: "",
		},
		{
			Name:          "when value is IPv6 link-local address, it should succeed",
			Value:         "fe80::1",
			ExpectedError: "",
		},
		{
			Name:          "when value is a pointer to pointer to valid IP, it should succeed",
			Value:         ptr.Of(ptr.Of("10.0.0.1")),
			ExpectedError: "",
		},
		{
			Name:          "when value is a pointer to pointer to invalid IP, it should return an error",
			Value:         ptr.Of(ptr.Of("invalid")),
			ExpectedError: "value 'invalid' could not be parsed as an IP address",
		},
		{
			Name:          "when value is IPv4 with leading zeros, it should return an error",
			Value:         "192.168.01.1",
			ExpectedError: "could not be parsed as an IP address",
		},
		{
			Name:          "when value is IPv4 with too few octets, it should return an error",
			Value:         "192.168.1",
			ExpectedError: "could not be parsed as an IP address",
		},
		{
			Name:          "when value is IPv4 with too many octets, it should return an error",
			Value:         "192.168.1.1.1",
			ExpectedError: "could not be parsed as an IP address",
		},
		{
			Name:          "when value is IPv6 with invalid segment, it should return an error",
			Value:         "2001:db8::gggg",
			ExpectedError: "could not be parsed as an IP address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := validation.Var(tc.Value, "ip_addr")
			if tc.ExpectedError != "" {
				assert.ErrorPart(t, err, tc.ExpectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
