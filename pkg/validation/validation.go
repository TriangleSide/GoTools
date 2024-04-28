package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate                      *validator.Validate
	customValidationErrorMessages map[string]func(err validator.FieldError) string
)

// init will create the validator and configure it.
func init() {
	validate = validator.New(validator.WithRequiredStructEnabled(), validator.WithPrivateFieldValidation())
	customValidationErrorMessages = make(map[string]func(err validator.FieldError) string)
}

// Struct returns an error if one or many of the struct members violate validation rules.
func Struct(val any) error {
	if err := validate.Struct(val); err != nil {
		return formatErrorMessage(err)
	}
	return nil
}

// Var validates a single variable using tag style validation that would be set on a struct field.
func Var(val any, tag string) error {
	if err := validate.Var(val, tag); err != nil {
		return formatErrorMessage(err)
	}
	return nil
}

// RegisterValidation registers a custom validator and error message generator for a tag.
// If it is called more than once for a tag, a panic occurs.
func RegisterValidation(tag string, validationFunc validator.Func, validationErrorMsg func(err validator.FieldError) string) {
	if _, ok := customValidationErrorMessages[tag]; ok {
		panic(fmt.Sprintf("Tag '%s' already has a registered validation function.", tag))
	}
	customValidationErrorMessages[tag] = validationErrorMsg
	err := validate.RegisterValidation(tag, validationFunc, true)
	if err != nil {
		panic(fmt.Sprintf("Failed to register the validation function for the tag '%s'.", tag))
	}
}

// formatErrorMessage takes a validation error and formats it.
func formatErrorMessage(err error) error {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		errorList := make([]string, 0)
		for _, fieldError := range validationErrs {
			if customErrorMsg, isCustomTag := customValidationErrorMessages[fieldError.Tag()]; isCustomTag {
				errorList = append(errorList, customErrorMsg(fieldError))
			} else {
				errorMsg := "validation failed"
				if fieldError.Field() != "" {
					errorMsg = errorMsg + " on field '" + fieldError.Field() + "'"
				}
				errorMsg = errorMsg + " with validator '" + fieldError.Tag() + "'"
				if fieldError.Param() != "" {
					errorMsg = errorMsg + " and parameter(s) '" + fieldError.Param() + "'"
				}
				errorList = append(errorList, errorMsg)
			}
		}
		return errors.New(strings.Join(errorList, "; "))
	}
	return err
}
