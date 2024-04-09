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

// init will create the validator and configure it. The default translation error
func init() {
	validate = validator.New()
	customValidationErrorMessages = make(map[string]func(err validator.FieldError) string)
}

// Validate returns an error if one or many of the struct members violate validation rules.
func Validate(val any) error {
	if err := validate.Struct(val); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			errorList := make([]string, 0)
			for _, fieldError := range validationErrs {
				if customErrorMsg, ok := customValidationErrorMessages[fieldError.Tag()]; ok {
					errorList = append(errorList, customErrorMsg(fieldError))
				} else {
					errorList = append(errorList, fmt.Sprintf("validation failed on field '%s' with validator '%s' with parameter(s) '%s'", fieldError.Field(), fieldError.Tag(), fieldError.Param()))
				}
			}
			return errors.New(strings.Join(errorList, "; "))
		} else {
			return err
		}
	}
	return nil
}

// RegisterValidation registers a custom validator and error message generator for a tag.
// RegisterValidation may only be called once for a specific tag, else a panic occurs.
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
