package validation

import (
	"fmt"
	"net"
	"reflect"
)

const (
	IPAddrValidatorName Validator = "ip_addr"
)

// init registers the validator.
func init() {
	MustRegisterValidator(IPAddrValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		value := params.Value
		if !DereferenceValue(&value) {
			return result.WithError(NewViolation(IPAddrValidatorName, params, DefaultDeferenceErrorMessage))
		}

		if value.Kind() != reflect.String {
			return result.WithError(fmt.Errorf("value must be a string for the %s validator", IPAddrValidatorName))
		}

		var valueStr = value.String()
		if ip := net.ParseIP(valueStr); ip == nil {
			return result.WithError(NewViolation(IPAddrValidatorName, params, fmt.Sprintf("the value '%s' could not be parsed as an IP address", valueStr)))
		}

		return nil
	})
}
