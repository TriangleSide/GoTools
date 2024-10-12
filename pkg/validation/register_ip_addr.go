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
	MustRegisterValidator(IPAddrValidatorName, func(params *CallbackParameters) error {
		value := params.Value
		if ValueIsNil(value) {
			return NewViolation(IPAddrValidatorName, params, defaultNilErrorMessage)
		}
		DereferenceValue(&value)

		if value.Kind() != reflect.String {
			return fmt.Errorf("value must be a string for the %s validator", IPAddrValidatorName)
		}

		var valueStr = value.String()
		if ip := net.ParseIP(valueStr); ip == nil {
			return NewViolation(IPAddrValidatorName, params, fmt.Sprintf("the value '%s' could not be parsed as an IP address", valueStr))
		}

		return nil
	})
}
