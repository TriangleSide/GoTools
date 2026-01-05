package validation

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/TriangleSide/GoTools/pkg/reflection"
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

var (
	// convertToFloatMap maps reflect.Kind to functions that convert reflect.Value to float64.
	convertToFloatMap = map[reflect.Kind]func(reflect.Value) float64{
		reflect.Int:     func(v reflect.Value) float64 { return float64(v.Int()) },
		reflect.Int8:    func(v reflect.Value) float64 { return float64(v.Int()) },
		reflect.Int16:   func(v reflect.Value) float64 { return float64(v.Int()) },
		reflect.Int32:   func(v reflect.Value) float64 { return float64(v.Int()) },
		reflect.Int64:   func(v reflect.Value) float64 { return float64(v.Int()) },
		reflect.Uint:    func(v reflect.Value) float64 { return float64(v.Uint()) },
		reflect.Uint8:   func(v reflect.Value) float64 { return float64(v.Uint()) },
		reflect.Uint16:  func(v reflect.Value) float64 { return float64(v.Uint()) },
		reflect.Uint32:  func(v reflect.Value) float64 { return float64(v.Uint()) },
		reflect.Uint64:  func(v reflect.Value) float64 { return float64(v.Uint()) },
		reflect.Float32: func(v reflect.Value) float64 { return v.Float() },
		reflect.Float64: func(v reflect.Value) float64 { return v.Float() },
	}
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

		value := reflection.Dereference(params.Value)
		if reflection.IsNil(value) {
			return NewCallbackResult().AddFieldError(NewFieldError(params, errValueIsNil)), nil
		}

		convertFunc, ok := convertToFloatMap[value.Kind()]
		if !ok {
			fieldErr := NewFieldError(params, fmt.Errorf("the %s validation not supported for kind %s", name, value.Kind()))
			return NewCallbackResult().AddFieldError(fieldErr), nil
		}

		valFloat64 := convertFunc(value)

		if !compareFunc(valFloat64, threshold) {
			fieldErr := NewFieldError(params, fmt.Errorf("the value %v must be %s %v", valFloat64, operator, threshold))
			return NewCallbackResult().AddFieldError(fieldErr), nil
		}

		return NewCallbackResult().PassValidation(), nil
	})
}
