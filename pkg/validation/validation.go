package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate                      = validator.New(validator.WithRequiredStructEnabled(), validator.WithPrivateFieldValidation())
	customValidationErrorMessages = make(map[string]func(err validator.FieldError) string)
)

// RegisterValidation registers a custom validator and error message generator for a tag.
// If it is called more than once for a tag, a panic occurs.
func RegisterValidation(tag string, validationFunc validator.Func, validationErrorMsg func(err validator.FieldError) string) {
	if _, ok := customValidationErrorMessages[tag]; ok {
		panic(fmt.Sprintf("Tag '%s' already has a registered validation function.", tag))
	}
	if validationErrorMsg == nil {
		panic(fmt.Sprintf("Tag '%s' has a nil error message function.", tag))
	}
	customValidationErrorMessages[tag] = validationErrorMsg
	if err := validate.RegisterValidation(tag, validationFunc, true); err != nil {
		panic(fmt.Sprintf("Failed to register the validation function for the tag '%s'.", tag))
	}
}

// Struct returns an error if one or many of the struct members violate validation rules.
func Struct[T any](val T) error {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return errors.New("struct validation on nil value")
	}
	if v.Kind() != reflect.Struct && !(v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		panic(fmt.Sprintf("Type must be a struct or a pointer to a struct."))
	}
	if err := validate.Struct(val); err != nil {
		return formatErrorMessage(err)
	}
	return nil
}

// Var validates a single variable using tag style validation that would be set on a struct field.
func Var[T any](val T, tag string) error {
	if err := validate.Var(val, tag); err != nil {
		return formatErrorMessage(err)
	}
	return nil
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
				sb := strings.Builder{}
				sb.WriteString("validation failed")
				if fieldError.Field() != "" {
					sb.WriteString(" on field '")
					sb.WriteString(fieldError.Field())
					sb.WriteString("'")
				}
				sb.WriteString(" with validator '")
				sb.WriteString(fieldError.Tag())
				sb.WriteString("'")
				if fieldError.Param() != "" {
					sb.WriteString(" and parameter(s) '")
					sb.WriteString(fieldError.Param())
					sb.WriteString("'")
				}
				errorList = append(errorList, sb.String())
			}
		}
		return errors.New(strings.Join(errorList, "; "))
	}
	return err
}
