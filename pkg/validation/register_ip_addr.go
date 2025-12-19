package validation

import (
	"errors"
	"fmt"
	"net"
	"reflect"
)

const (
	// IPAddrValidatorName is the name of the validator that checks if a string is a valid IP address.
	IPAddrValidatorName Validator = "ip_addr"
)

// init registers the ip_addr validator that checks if a string value is a valid IPv4 or IPv6 address.
func init() {
	MustRegisterValidator(IPAddrValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.WithError(NewViolationError(params, err))
		}
		if value.Kind() != reflect.String {
			return result.WithError(errors.New("the value must be a string"))
		}

		var valueStr = value.String()
		if ip := net.ParseIP(valueStr); ip == nil {
			return result.WithError(NewViolationError(params, fmt.Errorf(
				"the value '%s' could not be parsed as an IP address", valueStr)))
		}

		return nil
	})
}
