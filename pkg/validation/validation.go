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
// It returns a list of field errors and an error if anything went wrong while validating.
func checkValidatorsAgainstValue(
	isStructValue bool,
	structValue reflect.Value,
	structFieldName string,
	fieldValue reflect.Value,
	validationTagContents string,
) ([]*FieldError, error) {
	var fieldErrors []*FieldError
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

		callbackResponse, callbackErr := callback(callbackParameters)
		if callbackErr != nil {
			return false, callbackErr
		}

		if callbackResponse != nil {
			if len(callbackResponse.fieldErrors) > 0 {
				fieldErrors = append(fieldErrors, callbackResponse.fieldErrors...)
				return false, nil
			}
			if callbackResponse.stop {
				return false, nil
			}
			if callbackResponse.newValues != nil {
				for _, newValue := range callbackResponse.newValues {
					newFieldErrors, newValErr := checkValidatorsAgainstValue(
						isStructValue, structValue, structFieldName, newValue, rest())
					if newValErr != nil {
						return false, newValErr
					}
					fieldErrors = append(fieldErrors, newFieldErrors...)
				}
				return false, nil
			}
			return false, fmt.Errorf("callback response is not correctly filled for validator %s", name)
		}

		return true, nil
	}
	err := forEachValidatorAndInstruction(validationTagContents, iterCallback)
	return fieldErrors, err
}

// validateContainerElements validates elements within slices, arrays, and maps.
func validateContainerElements(depth int, val reflect.Value) ([]*FieldError, error) {
	var fieldErrors []*FieldError
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := range val.Len() {
			elementFieldErrors, err := validateRecursively(depth+1, val.Index(i))
			if err != nil {
				return nil, err
			}
			fieldErrors = append(fieldErrors, elementFieldErrors...)
		}
	case reflect.Map:
		mapRange := val.MapRange()
		for mapRange.Next() {
			keyFieldErrors, err := validateRecursively(depth+1, mapRange.Key())
			if err != nil {
				return nil, err
			}
			fieldErrors = append(fieldErrors, keyFieldErrors...)
			valueFieldErrors, err := validateRecursively(depth+1, mapRange.Value())
			if err != nil {
				return nil, err
			}
			fieldErrors = append(fieldErrors, valueFieldErrors...)
		}
	default:
		// Not a container type; skipping.
	}
	return fieldErrors, nil
}

// validateRecursively ensures nested structs inside containers (slices, arrays, maps) are
// validated, even when the container field itself has no validate tag. For example, a field
// "Users []User" with no tag will still have each User struct validated for its own constraints.
func validateRecursively(depth int, val reflect.Value) ([]*FieldError, error) {
	const maxDepth = 32
	if depth >= maxDepth {
		return nil, errors.New("cycle found in the validation")
	}

	val = reflection.Dereference(val)
	if reflection.IsNil(val) {
		return nil, nil
	}

	if val.Kind() == reflect.Struct {
		return validateStructInternal(val, depth+1)
	}
	return validateContainerElements(depth, val)
}

// Struct validates all struct fields using their validation tags, returning an error if any fail.
// In the case that the struct has tag field errors, the field errors are joined with errors.Join.
func Struct[T any](val T) error {
	reflectValue, err := dereferenceAndNilCheck(reflect.ValueOf(val))
	if err != nil {
		return err
	}
	fieldErrors, err := validateStructInternal(reflectValue, 0)
	if err != nil {
		return err
	}
	return fieldErrorsToError(fieldErrors)
}

// validateStructInternal is a helper for the Struct and validateRecursively functions.
func validateStructInternal(val reflect.Value, depth int) ([]*FieldError, error) {
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("validation parameter must be a struct but got %s", val.Kind())
	}

	var fieldErrors []*FieldError
	structMetadataMap := structs.MetadataFromType(val.Type())

	for fieldName, fieldMetadata := range structMetadataMap.All() {
		fieldValueFromStruct, _ := structs.ValueFromName(val.Interface(), fieldName)

		if validationTag, hasValidationTag := fieldMetadata.Tags().Fetch(Tag); hasValidationTag {
			validationFieldErrors, err := checkValidatorsAgainstValue(
				true, val, fieldName, fieldValueFromStruct, validationTag)
			if err != nil {
				return nil, err
			}
			fieldErrors = append(fieldErrors, validationFieldErrors...)
		}

		recursiveFieldErrors, err := validateRecursively(depth, fieldValueFromStruct)
		if err != nil {
			return nil, err
		}
		fieldErrors = append(fieldErrors, recursiveFieldErrors...)
	}

	return fieldErrors, nil
}

// Var validates a single variable with the given instructions, returning an error if it fails.
// In the case that the variable has tag field errors, the field errors are joined with errors.Join.
func Var[T any](val T, validatorInstructions string) error {
	reflectValue := reflect.ValueOf(val)
	fieldErrors, err := checkValidatorsAgainstValue(false, reflect.Value{}, "", reflectValue, validatorInstructions)
	if err != nil {
		return err
	}
	recursiveFieldErrors, err := validateRecursively(0, reflectValue)
	if err != nil {
		return err
	}
	fieldErrors = append(fieldErrors, recursiveFieldErrors...)
	return fieldErrorsToError(fieldErrors)
}

// fieldErrorsToError converts a slice of field errors to an error by joining them.
func fieldErrorsToError(fieldErrors []*FieldError) error {
	if len(fieldErrors) == 0 {
		return nil
	}
	errs := make([]error, len(fieldErrors))
	for i, fieldErr := range fieldErrors {
		errs[i] = fieldErr
	}
	return errors.Join(errs...)
}
