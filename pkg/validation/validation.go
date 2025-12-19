package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/structs"
)

// Validator is the name of a validate rule.
// For example: oneof, required, dive, etc...
type Validator string

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
	// The tag contains the validators and their respective instructions.
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

// expandAliases expands all aliases in the validation tag.
func expandAliases(validateTagContents string) string {
	namesToInstructions := strings.Split(validateTagContents, ValidatorsSep)
	var result []string

	for _, nameToInstruction := range namesToInstructions {
		validatorName, _, parseErr := parseValidatorNameAndInstruction(nameToInstruction)
		if parseErr != nil {
			result = append(result, nameToInstruction)
			continue
		}

		if expansion, isAlias := lookupAlias(validatorName); isAlias {
			result = append(result, expansion)
		} else {
			result = append(result, nameToInstruction)
		}
	}

	return strings.Join(result, ValidatorsSep)
}

// forEachValidatorAndInstruction invokes the callback for each validator name and instruction.
func forEachValidatorAndInstruction(
	validateTagContents string,
	callback func(name string, instruction string, rest func() string) (bool, error),
) error {
	if strings.TrimSpace(validateTagContents) == "" {
		return fmt.Errorf("empty %s instructions", Tag)
	}

	expandedTagContents := expandAliases(validateTagContents)
	namesToInstructions := strings.Split(expandedTagContents, ValidatorsSep)

	for instructionIdx, nameToInstruction := range namesToInstructions {
		validatorName, validatorInstructions, parseErr := parseValidatorNameAndInstruction(nameToInstruction)
		if parseErr != nil {
			return parseErr
		}

		restOfValidationTag := func() string {
			return strings.Join(namesToInstructions[instructionIdx+1:], ValidatorsSep)
		}

		shouldContinue, err := callback(validatorName, validatorInstructions, restOfValidationTag)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}
	}

	return nil
}

// checkValidatorsAgainstValue validates a value based on the provided validation tag.
// It returns an error if anything went wrong while validating.
func checkValidatorsAgainstValue(
	isStructValue bool,
	structValue reflect.Value,
	structFieldName string,
	fieldValue reflect.Value,
	validationTagContents string,
	violations *ViolationsError,
) error {
	iterCallback := func(name string, instruction string, rest func() string) (bool, error) {
		callbackNotCast, callbackFound := registeredValidations.Load(name)
		if !callbackFound {
			return false, fmt.Errorf("validation with name '%s' is not registered", name)
		}
		callback := callbackNotCast.(Callback)
		callbackParameters := &CallbackParameters{
			Validator:          Validator(name),
			IsStructValidation: isStructValue,
			StructValue:        structValue,
			StructFieldName:    structFieldName,
			Value:              fieldValue,
			Parameters:         instruction,
		}

		if callbackResponse := callback(callbackParameters); callbackResponse != nil {
			if callbackResponse.err != nil {
				var violation *ViolationError
				if errors.As(callbackResponse.err, &violation) {
					violations.AddViolation(violation)
					return false, nil
				}
				return false, callbackResponse.err
			}
			if callbackResponse.stop {
				return false, nil
			}
			if callbackResponse.newValues != nil {
				for _, newValue := range callbackResponse.newValues {
					newValErr := checkValidatorsAgainstValue(
						isStructValue, structValue, structFieldName, newValue, rest(), violations)
					if newValErr != nil {
						return false, newValErr
					}
				}
				return false, nil
			}
			return false, fmt.Errorf("callback response is not correctly filled for validator %s", name)
		}

		return true, nil
	}
	return forEachValidatorAndInstruction(validationTagContents, iterCallback)
}

// validateNestedStruct validates a nested struct and accumulates any violations.
func validateNestedStruct(depth int, val reflect.Value, violations *ViolationsError) error {
	err := validateStruct(val.Interface(), depth+1)
	if err == nil {
		return nil
	}
	var structViolations *ViolationsError
	if errors.As(err, &structViolations) {
		violations.AddViolations(structViolations)
		return nil
	}
	return err
}

// validateContainerElements validates elements within slices, arrays, and maps.
func validateContainerElements(depth int, val reflect.Value, violations *ViolationsError) error {
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := range val.Len() {
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
	}
	return nil
}

// validateRecursively ensures nested structs inside containers (slices, arrays, maps) are
// validated, even when the container field itself has no validate tag. For example, a field
// "Users []User" with no tag will still have each User struct validated for its own constraints.
func validateRecursively(depth int, val reflect.Value, violations *ViolationsError) error {
	const maxDepth = 32
	if depth >= maxDepth {
		return errors.New("cycle found in the validation")
	}

	val = reflection.Dereference(val)
	if reflection.IsNil(val) {
		return nil
	}

	if val.Kind() == reflect.Struct {
		return validateNestedStruct(depth, val, violations)
	}
	return validateContainerElements(depth, val, violations)
}

// Struct validates all struct fields using their validation tags, returning an error if any fail.
// In the case that the struct has tag violations, a ViolationsError error is returned.
func Struct[T any](val T) error {
	return validateStruct(val, 0)
}

// validateStruct is a helper for the Struct and validateRecursively functions.
func validateStruct[T any](val T, depth int) error {
	reflectValue, err := dereferenceAndNilCheck(reflect.ValueOf(val))
	if err != nil {
		return err
	}
	if reflectValue.Kind() != reflect.Struct {
		panic(fmt.Errorf("validation parameter must be a struct, got %s", reflectValue.Kind()))
	}

	violations := NewViolationsError()
	structMetadataMap := structs.MetadataFromType(reflectValue.Type())

	for fieldName, fieldMetadata := range structMetadataMap.All() {
		fieldValueFromStruct, _ := structs.ValueFromName(val, fieldName)

		if validationTag, hasValidationTag := fieldMetadata.Tags().Fetch(Tag); hasValidationTag {
			err = checkValidatorsAgainstValue(
				true, reflectValue, fieldName, fieldValueFromStruct, validationTag, violations)
			if err != nil {
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
// In the case that the variable has tag violations, a ViolationsError error is returned.
func Var[T any](val T, validatorInstructions string) error {
	reflectValue := reflect.ValueOf(val)
	violations := NewViolationsError()
	err := checkValidatorsAgainstValue(false, reflect.Value{}, "", reflectValue, validatorInstructions, violations)
	if err != nil {
		return err
	}
	err = validateRecursively(0, reflectValue, violations)
	if err != nil {
		return err
	}
	return violations.NilIfEmpty()
}
