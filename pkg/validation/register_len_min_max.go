package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const (
	LenValidatorName Validator = "len"
	MinValidatorName Validator = "min"
	MaxValidatorName Validator = "max"
)

// init registers the validators.
func init() {
	registerStringLengthValidation(LenValidatorName, func(length, target int) bool { return length == target }, "exactly")
	registerStringLengthValidation(MinValidatorName, func(length, target int) bool { return length >= target }, "at least")
	registerStringLengthValidation(MaxValidatorName, func(length, target int) bool { return length <= target }, "at most")
}

// registerStringLengthValidation consolidates common logic for string length validations.
func registerStringLengthValidation(name Validator, compareFunc func(length, target int) bool, descriptor string) {
	MustRegisterValidator(name, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		targetLength, err := strconv.Atoi(params.Parameters)
		if err != nil {
			return result.WithError(fmt.Errorf("invalid instruction '%s' for %s: %w", params.Parameters, name, err))
		}
		if targetLength < 0 {
			return result.WithError(errors.New("the length parameter can't be negative"))
		}

		value := params.Value
		if !DereferenceValue(&value) {
			return result.WithError(NewViolation(name, params, DefaultDeferenceErrorMessage))
		}

		if value.Kind() != reflect.String {
			return result.WithError(fmt.Errorf("the value must be a string for the %s validator", name))
		}

		var valueStr = value.String()
		if !compareFunc(len(valueStr), targetLength) {
			return result.WithError(NewViolation(name, params, fmt.Sprintf("the length %d must be %s %d", len(valueStr), descriptor, targetLength)))
		}

		return nil
	})
}
