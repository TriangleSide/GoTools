package validation

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	GreaterThanValidatorName        Validator = "gt"
	GreaterThanOrEqualValidatorName Validator = "gte"
	LessThanValidatorName           Validator = "lt"
	LessThanOrEqualValidatorName    Validator = "lte"
)

// init registers the validators.
func init() {
	registerComparisonValidation(GreaterThanValidatorName, func(a, b float64) bool { return a > b }, "greater than")
	registerComparisonValidation(GreaterThanOrEqualValidatorName, func(a, b float64) bool { return a >= b }, "greater than or equal to")
	registerComparisonValidation(LessThanValidatorName, func(a, b float64) bool { return a < b }, "less than")
	registerComparisonValidation(LessThanOrEqualValidatorName, func(a, b float64) bool { return a <= b }, "less than or equal to")
}

// registerComparisonValidation consolidates the common logic for comparison validations.
func registerComparisonValidation(name Validator, compareFunc func(a, b float64) bool, operator string) {
	MustRegisterValidator(name, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		threshold, err := strconv.ParseFloat(params.Parameters, 64)
		if err != nil {
			return result.WithError(fmt.Errorf("invalid parameters '%s' for %s: %w", params.Parameters, name, err))
		}

		value, err := DereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.WithError(NewViolation(params, err.Error()))
		}

		var val float64
		switch kind := value.Kind(); kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val = float64(value.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val = float64(value.Uint())
		case reflect.Float32, reflect.Float64:
			val = value.Float()
		default:
			return result.WithError(NewViolation(params, fmt.Sprintf("the %s validation not supported for kind %s", name, kind)))
		}

		if !compareFunc(val, threshold) {
			return result.WithError(NewViolation(params, fmt.Sprintf("the value %v must be %s %v", val, operator, threshold)))
		}

		return nil
	})
}
