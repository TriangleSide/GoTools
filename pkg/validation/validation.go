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
	//     Value int `validate:"gt=0"`
	// }
	Tag = "validate"
)

// parseValidatorNameAndInstruction takes a `validator=instructions` string and splits it.
func parseValidatorNameAndInstruction(nameToInstruction string) (string, string, error) {
	const maxNameToInstructionParts = 2
	const validatorNameIndex = 0
	const validatorInstructionsIndex = 1
	nameToInstructionParts := strings.Split(nameToInstruction, NameAndInstructionsSep)
	if len(nameToInstructionParts) > maxNameToInstructionParts {
		return "", "", fmt.Errorf("malformed validator and instruction")
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
			StructValidation: isStructValue,
			StructValue:      structValue,
			StructFieldName:  structFieldName,
			Value:            fieldValue,
			Parameters:       instruction,
		}

		var violation *Violation
		var validatorsStop *stopValidators
		var valuesNew *newValues

		if err := callback(callbackParameters); err != nil {
			if errors.As(err, &violation) {
				violations.AddViolation(violation)
				return true, nil
			} else if errors.As(err, &validatorsStop) {
				return false, nil
			} else if errors.As(err, &valuesNew) {
				for _, newValue := range valuesNew.values {
					if newValErr := checkValidatorsAgainstValue(isStructValue, structValue, structFieldName, newValue, rest(), violations); newValErr != nil {
						return false, newValErr
					}
				}
				return false, nil
			} else {
				return false, err
			}
		}

		return true, nil
	})
}

// validateRecursively checks if the value is a container, like a slice, and checks if the
// contents can be validated as well.
func validateRecursively(val reflect.Value, violations *Violations) error {
	if ValueIsNil(val) {
		return nil
	}
	DereferenceValue(&val)

	switch val.Kind() {
	case reflect.Struct:
		if err := Struct(val.Interface()); err != nil {
			var structViolations *Violations
			if errors.As(err, &structViolations) {
				violations.AddViolations(structViolations)
			} else {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if err := validateRecursively(val.Index(i), violations); err != nil {
				return err
			}
		}
	case reflect.Map:
		mapRange := val.MapRange()
		for mapRange.Next() {
			if err := validateRecursively(mapRange.Key(), violations); err != nil {
				return err
			}
			if err := validateRecursively(mapRange.Value(), violations); err != nil {
				return err
			}
		}
	default:
		// Do nothing.
	}

	return nil
}

// Struct validates all struct fields using their validation tags, returning an error if any fail.
func Struct[T any](val T) error {
	reflectVal := reflect.ValueOf(val)
	if ValueIsNil(reflectVal) {
		return errors.New("nil parameter on struct validation")
	}
	DereferenceValue(&reflectVal)
	if reflectVal.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Struct validation parameter must be a struct but got %s.", reflectVal.Kind()))
	}

	violations := NewViolations()
	structMetadataMap := fields.StructMetadataFromType(reflectVal.Type())

	for fieldName, fieldMetadata := range structMetadataMap.All() {
		fieldValueFromStruct, _ := fields.StructValueFromName(val, fieldName)

		if validationTag, hasValidationTag := fieldMetadata.Tags().Fetch(Tag); hasValidationTag {
			if err := checkValidatorsAgainstValue(true, reflectVal, fieldName, fieldValueFromStruct, validationTag, violations); err != nil {
				return err
			}
		}

		if err := validateRecursively(fieldValueFromStruct, violations); err != nil {
			return err
		}
	}

	return violations.NilIfEmpty()
}

// Var validates a single variable with the given instructions, returning an error if it fails.
func Var[T any](val T, validatorInstructions string) error {
	reflectValue := reflect.ValueOf(val)
	violations := NewViolations()
	if err := checkValidatorsAgainstValue(false, reflect.Value{}, "", reflectValue, validatorInstructions, violations); err != nil {
		return err
	}
	if err := validateRecursively(reflectValue, violations); err != nil {
		return err
	}
	return violations.NilIfEmpty()
}
