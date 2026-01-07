package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestIPAddrValidator_StructFieldWithValidIP_Succeeds(t *testing.T) {
	t.Parallel()
	type Config struct {
		IPAddress string `validate:"ip_addr"`
	}
	cfg := Config{IPAddress: "192.168.1.1"}
	err := validation.Struct(&cfg)
	assert.NoError(t, err)
}

func TestIPAddrValidator_StructFieldWithInvalidIP_ReturnsError(t *testing.T) {
	t.Parallel()
	type Config struct {
		IPAddress string `validate:"ip_addr"`
	}
	cfg := Config{IPAddress: "invalid_ip"}
	err := validation.Struct(&cfg)
	assert.ErrorPart(t, err, "could not be parsed as an IP address")
}

func TestIPAddrValidator_ValidIPv4AddressString_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("192.168.1.1", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_ValidIPv6AddressString_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("2001:db8::1", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_InvalidIPAddressString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("invalid_ip", "ip_addr")
	assert.ErrorPart(t, err, "value 'invalid_ip' could not be parsed as an IP address")
}

func TestIPAddrValidator_NonStringValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(12345, "ip_addr")
	assert.ErrorPart(t, err, "value must be a string")
}

func TestIPAddrValidator_NilPointer_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*string)(nil), "ip_addr")
	assert.ErrorPart(t, err, "value is nil")
}

func TestIPAddrValidator_PointerToStringWithValidIP_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("8.8.8.8"), "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_PointerToStringWithInvalidIP_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("not_an_ip"), "ip_addr")
	assert.ErrorPart(t, err, "value 'not_an_ip' could not be parsed as an IP address")
}

func TestIPAddrValidator_InterfaceWrappingStringWithValidIP_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var(any("127.0.0.1"), "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_InterfaceWrappingStringWithInvalidIP_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any("invalid_ip"), "ip_addr")
	assert.ErrorPart(t, err, "value 'invalid_ip' could not be parsed as an IP address")
}

func TestIPAddrValidator_InterfaceWrappingNilPointer_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any((*string)(nil)), "ip_addr")
	assert.ErrorPart(t, err, "value is nil")
}

func TestIPAddrValidator_InterfaceWrappingNonStringValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any(12345), "ip_addr")
	assert.ErrorPart(t, err, "the value must be a string")
}

func TestIPAddrValidator_StringWithExtraSpaces_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(" 192.168.1.1 ", "ip_addr")
	assert.ErrorPart(t, err, "value ' 192.168.1.1 ' could not be parsed as an IP address")
}

func TestIPAddrValidator_EmptyString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "ip_addr")
	assert.ErrorPart(t, err, "value '' could not be parsed as an IP address")
}

func TestIPAddrValidator_MalformedIPAddress_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("256.256.256.256", "ip_addr")
	assert.ErrorPart(t, err, "value '256.256.256.256' could not be parsed as an IP address")
}

func TestIPAddrValidator_IPv4AllZeros_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("0.0.0.0", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_IPv4BroadcastAddress_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("255.255.255.255", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_IPv6LoopbackAddress_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("::1", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_IPv6AllZeros_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("::", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_IPv6FullForm_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("2001:0db8:0000:0000:0000:0000:0000:0001", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_IPv4MappedIPv6Address_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("::ffff:192.168.1.1", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_IPv6LinkLocalAddress_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("fe80::1", "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_PointerToPointerToValidIP_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(ptr.Of("10.0.0.1")), "ip_addr")
	assert.NoError(t, err)
}

func TestIPAddrValidator_PointerToPointerToInvalidIP_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(ptr.Of("invalid")), "ip_addr")
	assert.ErrorPart(t, err, "value 'invalid' could not be parsed as an IP address")
}

func TestIPAddrValidator_IPv4WithLeadingZeros_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("192.168.01.1", "ip_addr")
	assert.ErrorPart(t, err, "could not be parsed as an IP address")
}

func TestIPAddrValidator_IPv4WithTooFewOctets_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("192.168.1", "ip_addr")
	assert.ErrorPart(t, err, "could not be parsed as an IP address")
}

func TestIPAddrValidator_IPv4WithTooManyOctets_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("192.168.1.1.1", "ip_addr")
	assert.ErrorPart(t, err, "could not be parsed as an IP address")
}

func TestIPAddrValidator_IPv6WithInvalidSegment_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("2001:db8::gggg", "ip_addr")
	assert.ErrorPart(t, err, "could not be parsed as an IP address")
}
