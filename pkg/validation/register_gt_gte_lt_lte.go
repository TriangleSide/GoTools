package validation

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	// GreaterThanValidatorName is the name of the validator that checks
	// if a numeric value is greater than a threshold.
	GreaterThanValidatorName Validator = "gt"
	// GreaterThanOrEqualValidatorName is the name of the validator that checks
	// if a numeric value is greater than or equal to a threshold.
	GreaterThanOrEqualValidatorName Validator = "gte"
	// LessThanValidatorName is the name of the validator that checks
	// if a numeric value is less than a threshold.
	LessThanValidatorName Validator = "lt"
	// LessThanOrEqualValidatorName is the name of the validator that checks
	// if a numeric value is less than or equal to a threshold.
	LessThanOrEqualValidatorName Validator = "lte"
)

// init registers the numeric comparison validators (gt, gte, lt, lte) for comparing values against thresholds.
func init() {
	registerComparisonValidation(GreaterThanValidatorName,
		func(a, b float64) bool { return a > b }, "greater than")
	registerComparisonValidation(GreaterThanOrEqualValidatorName,
		func(a, b float64) bool { return a >= b }, "greater than or equal to")
	registerComparisonValidation(LessThanValidatorName,
		func(a, b float64) bool { return a < b }, "less than")
	registerComparisonValidation(LessThanOrEqualValidatorName,
		func(a, b float64) bool { return a <= b }, "less than or equal to")
}

// registerComparisonValidation consolidates the common logic for comparison validations.
func registerComparisonValidation(name Validator, compareFunc func(a, b float64) bool, operator string) {
	MustRegisterValidator(name, func(params *CallbackParameters) (*CallbackResult, error) {
		threshold, err := strconv.ParseFloat(params.Parameters, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid parameters '%s' for %s: %w", params.Parameters, name, err)
		}

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return NewCallbackResult().AddFieldError(NewFieldError(params, err)), nil
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
			fieldErr := NewFieldError(params, fmt.Errorf("the %s validation not supported for kind %s", name, kind))
			return NewCallbackResult().AddFieldError(fieldErr), nil
		}

		if !compareFunc(val, threshold) {
			fieldErr := NewFieldError(params, fmt.Errorf("the value %v must be %s %v", val, operator, threshold))
			return NewCallbackResult().AddFieldError(fieldErr), nil
		}

		return nil, nil //nolint:nilnil // nil, nil means validation passed
	})
}
