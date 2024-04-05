package validation

import (
	"errors"
	"strings"

	enLocale "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	defaultTranslations "github.com/go-playground/validator/v10/translations/en"
)

const (
	defaultLocale = "en"
)

var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
)

// init will create the validator and configure it. The default translation error
func init() {
	validate = validator.New()
	en := enLocale.New()
	uni = ut.New(en, en)
	defaultTranslator, _ := uni.GetTranslator(defaultLocale)
	if defaultTranslations.RegisterDefaultTranslations(validate, defaultTranslator) != nil {
		// A logger can't be used here since the logger uses validation, and it would create a circular dependency.
		panic("Failed to register the default translation for the validator.")
	}
}

// Validate returns an error if one or many of the struct members violate validation rules.
func Validate(val any) error {
	if err := validate.Struct(val); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			errorList := make([]string, 0)
			for _, validationErr := range validationErrs {
				errorList = append(errorList, validationErr.Error())
			}
			return errors.New(strings.Join(errorList, "; "))
		} else {
			return err
		}
	}
	return nil
}
