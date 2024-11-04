package validation

import (
	"errors"
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

		value, err := DereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.WithError(NewViolation(params, err))
		}
		if value.Kind() != reflect.String {
			return result.WithError(errors.New("the value must be a string"))
		}

		var valueStr = value.String()
		if ip := net.ParseIP(valueStr); ip == nil {
			return result.WithError(NewViolation(params, fmt.Errorf("the value '%s' could not be parsed as an IP address", valueStr)))
		}

		return nil
	})
}
