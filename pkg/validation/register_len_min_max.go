package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const (
	// LenValidatorName is the name of the validator that checks if a string has an exact byte length.
	LenValidatorName Validator = "len"
	// MinValidatorName is the name of the validator that checks if a string has a minimum byte length.
	MinValidatorName Validator = "min"
	// MaxValidatorName is the name of the validator that checks if a string has a maximum byte length.
	MaxValidatorName Validator = "max"
)

// init registers the string length validators (len, min, max) for checking string byte lengths.
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
			return result.SetError(fmt.Errorf("invalid instruction '%s' for %s: %w", params.Parameters, name, err))
		}
		if targetLength < 0 {
			return result.SetError(errors.New("the length parameter can't be negative"))
		}

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.SetError(NewFieldError(params, err))
		}
		if value.Kind() != reflect.String {
			return result.SetError(fmt.Errorf("the value must be a string for the %s validator", name))
		}

		var valueStr = value.String()
		if !compareFunc(len(valueStr), targetLength) {
			return result.SetError(NewFieldError(params, fmt.Errorf(
				"the length %d must be %s %d", len(valueStr), descriptor, targetLength)))
		}

		return nil
	})
}
