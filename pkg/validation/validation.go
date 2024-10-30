package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/TriangleSide/GoBase/pkg/utils/fields"
)

const (
	// ValidatorsSep is the separator between validation names. For example: "required,oneof=THIS THAT".
	ValidatorsSep = ","

	// NameAndInstructionsSep is the separator between the validation name and the instructions.
	// For example: "oneof=THIS THAT".
	NameAndInstructionsSep = "="

	// Tag is the name of the struct field tag.
	//
	// type Example struct {
	//     Value *int `validate:"required,gt=0"`
	// }
	//
	// The tag contents contains the validators and their respective instructions.
	Tag = "validate"
)

// parseValidatorNameAndInstruction takes a `validator=instructions` string and splits it.
func parseValidatorNameAndInstruction(nameToInstruction string) (string, string, error) {
	const maxNameToInstructionParts = 2
	const validatorNameIndex = 0
	const validatorInstructionsIndex = 1
	nameToInstructionParts := strings.Split(nameToInstruction, NameAndInstructionsSep)
	if len(nameToInstructionParts) > maxNameToInstructionParts {
		return "", "", errors.New("malformed validator and instruction")
	}
	validatorName := nameToInstructionParts[validatorNameIndex]
	validatorInstructions := ""
	if len(nameToInstructionParts) >= (validatorInstructionsIndex + 1) {
		validatorInstructions = nameToInstructionParts[validatorInstructionsIndex]
	}
	return validatorName, validatorInstructions, nil
}

// forEachValidatorAndInstruction invokes the callback for each validator name and instruction.
func forEachValidatorAndInstruction(validateTagContents string, callback func(name string, instruction string, rest func() string) (bool, error)) error {
	if strings.TrimSpace(validateTagContents) == "" {
		return fmt.Errorf("empty %s instructions", Tag)
	}
	namesToInstructions := strings.Split(validateTagContents, ValidatorsSep)

	for i := 0; i < len(namesToInstructions); i++ {
		validatorName, validatorInstructions, parseErr := parseValidatorNameAndInstruction(namesToInstructions[i])
		if parseErr != nil {
			return parseErr
		}

		restOfValidationTag := func() string {
			return strings.Join(namesToInstructions[i+1:], ValidatorsSep)
		}

		if shouldContinue, err := callback(validatorName, validatorInstructions, restOfValidationTag); err != nil {
			return err
		} else if !shouldContinue {
			return nil
		}
	}

	return nil
}

// checkValidatorsAgainstValue validates a value based on the provided validation tag.
// It returns an error if anything went wrong while validating.
func checkValidatorsAgainstValue(isStructValue bool, structValue reflect.Value, structFieldName string, fieldValue reflect.Value, validationTagContents string, violations *Violations) error {
	return forEachValidatorAndInstruction(validationTagContents, func(name string, instruction string, rest func() string) (bool, error) {
		callbackNotCast, callbackFound := registeredValidations.Load(name)
		if !callbackFound {
			return false, fmt.Errorf("validation with name '%s' is not registered", name)
		}
		callback := callbackNotCast.(Callback)
		callbackParameters := &CallbackParameters{
			IsStructValidation: isStructValue,
			StructValue:        structValue,
			StructFieldName:    structFieldName,
			Value:              fieldValue,
			Parameters:         instruction,
		}

		if callbackResponse := callback(callbackParameters); callbackResponse != nil {
			if callbackResponse.err != nil {
				var violation *Violation
				if errors.As(callbackResponse.err, &violation) {
					violations.AddViolation(violation)
					return false, nil
				} else {
					return false, callbackResponse.err
				}
			} else if callbackResponse.stop {
				return false, nil
			} else if callbackResponse.newValues != nil {
				for _, newValue := range callbackResponse.newValues {
					if newValErr := checkValidatorsAgainstValue(isStructValue, structValue, structFieldName, newValue, rest(), violations); newValErr != nil {
						return false, newValErr
					}
				}
				return false, nil
			} else {
				return false, fmt.Errorf("callback response is not correctly filled for validator %s", name)
			}
		}

		return true, nil
	})
}

// validateRecursively checks if the value is a container, like a slice, and checks if the
// contents can be validated as well.
func validateRecursively(depth int, val reflect.Value, violations *Violations) error {
	const maxDepth = 32
	if depth >= maxDepth {
		return errors.New("cycle found in the validation")
	}

	if !DereferenceValue(&val) {
		return nil
	}

	switch val.Kind() {
	case reflect.Struct:
		if err := validateStruct(val.Interface(), depth+1); err != nil {
			var structViolations *Violations
			if errors.As(err, &structViolations) {
				violations.AddViolations(structViolations)
			} else {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if err := validateRecursively(depth+1, val.Index(i), violations); err != nil {
				return err
			}
		}
	case reflect.Map:
		mapRange := val.MapRange()
		for mapRange.Next() {
			if err := validateRecursively(depth+1, mapRange.Key(), violations); err != nil {
				return err
			}
			if err := validateRecursively(depth+1, mapRange.Value(), violations); err != nil {
				return err
			}
		}
	default:
		// Do nothing.
	}

	return nil
}

// Struct validates all struct fields using their validation tags, returning an error if any fail.
// In the case that the struct has tag violations, a Violations error is returned.
func Struct[T any](val T) error {
	return validateStruct(val, 0)
}

// validateStruct is a helper for the Struct and validateRecursively functions.
func validateStruct[T any](val T, depth int) error {
	reflectValue := reflect.ValueOf(val)
	if !DereferenceValue(&reflectValue) {
		return errors.New(DefaultDeferenceErrorMessage)
	}
	if reflectValue.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Struct validation parameter must be a struct but got %s.", reflectValue.Kind()))
	}

	violations := NewViolations()
	structMetadataMap := fields.StructMetadataFromType(reflectValue.Type())

	for fieldName, fieldMetadata := range structMetadataMap.All() {
		fieldValueFromStruct, _ := fields.StructValueFromName(val, fieldName)

		if validationTag, hasValidationTag := fieldMetadata.Tags().Fetch(Tag); hasValidationTag {
			if err := checkValidatorsAgainstValue(true, reflectValue, fieldName, fieldValueFromStruct, validationTag, violations); err != nil {
				return err
			}
		}

		if err := validateRecursively(depth, fieldValueFromStruct, violations); err != nil {
			return err
		}
	}

	return violations.NilIfEmpty()
}

// Var validates a single variable with the given instructions, returning an error if it fails.
// In the case that the variable has tag violations, a Violations error is returned.
func Var[T any](val T, validatorInstructions string) error {
	reflectValue := reflect.ValueOf(val)
	violations := NewViolations()
	if err := checkValidatorsAgainstValue(false, reflect.Value{}, "", reflectValue, validatorInstructions, violations); err != nil {
		return err
	}
	if err := validateRecursively(0, reflectValue, violations); err != nil {
		return err
	}
	return violations.NilIfEmpty()
}
