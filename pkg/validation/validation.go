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

// validatorIterator is the callback signature for iterating over validators.
// It receives the validator name, its instruction parameters, and a function to get
// remaining validators in the tag (used for recursive validation like "dive").
// Returns true to continue iteration, false to stop early, or an error.
type validatorIterator func(
	name string,
	instruction string,
	remainingValidators func() string,
) (continueIteration bool, err error)

// forEachValidatorAndInstruction parses a validation tag and invokes the iterator
// for each validator. It handles alias expansion and stops iteration early if the
// iterator returns false or an error.
func forEachValidatorAndInstruction(tagContents string, iterator validatorIterator) error {
	if strings.TrimSpace(tagContents) == "" {
		return fmt.Errorf("empty %s instructions", Tag)
	}

	expandedContents := expandAliases(tagContents)
	validators := strings.Split(expandedContents, ValidatorsSep)

	for validatorIndex, validatorEntry := range validators {
		name, instruction, err := parseValidatorNameAndInstruction(validatorEntry)
		if err != nil {
			return err
		}

		remainingValidators := func() string {
			return strings.Join(validators[validatorIndex+1:], ValidatorsSep)
		}

		continueIteration, err := iterator(name, instruction, remainingValidators)
		if err != nil {
			return err
		}
		if !continueIteration {
			return nil
		}
	}

	return nil
}

// validationContext holds all the context needed to validate a value.
type validationContext struct {
	isStructValidation bool
	structValue        reflect.Value
	structFieldName    string
	fieldValue         reflect.Value
	violations         *Violations
}

// lookupValidator retrieves a registered validator by name.
func lookupValidator(name string) (Callback, error) {
	registered, found := registeredValidations.Load(name)
	if !found {
		return nil, fmt.Errorf("validation with name '%s' is not registered", name)
	}
	return registered.(Callback), nil
}

// handleValidationResult processes a validator's result and determines how to proceed.
// Returns (continueIteration, error).
func (ctx *validationContext) handleValidationResult(
	validatorName string,
	result *CallbackResult,
	remainingValidators func() string,
) (bool, error) {
	if result == nil {
		return true, nil
	}

	if result.err != nil {
		var violation *Violation
		if errors.As(result.err, &violation) {
			ctx.violations.AddViolation(violation)
			return false, nil
		}
		return false, result.err
	}

	if result.stop {
		return false, nil
	}

	if result.newValues != nil {
		remainingTag := remainingValidators()
		for _, newValue := range result.newValues {
			err := checkValidatorsAgainstValue(
				ctx.isStructValidation,
				ctx.structValue,
				ctx.structFieldName,
				newValue,
				remainingTag,
				ctx.violations,
			)
			if err != nil {
				return false, err
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("callback response is not correctly filled for validator %s", validatorName)
}

// checkValidatorsAgainstValue validates a value against a validation tag.
// It iterates through each validator in the tag and applies them in order.
func checkValidatorsAgainstValue(
	isStructValidation bool,
	structValue reflect.Value,
	structFieldName string,
	fieldValue reflect.Value,
	tagContents string,
	violations *Violations,
) error {
	ctx := &validationContext{
		isStructValidation: isStructValidation,
		structValue:        structValue,
		structFieldName:    structFieldName,
		fieldValue:         fieldValue,
		violations:         violations,
	}

	iterator := func(
		name, instruction string,
		remainingValidators func() string,
	) (bool, error) {
		validator, err := lookupValidator(name)
		if err != nil {
			return false, err
		}

		params := &CallbackParameters{
			Validator:          Validator(name),
			IsStructValidation: ctx.isStructValidation,
			StructValue:        ctx.structValue,
			StructFieldName:    ctx.structFieldName,
			Value:              ctx.fieldValue,
			Parameters:         instruction,
		}

		result := validator(params)
		return ctx.handleValidationResult(name, result, remainingValidators)
	}

	return forEachValidatorAndInstruction(tagContents, iterator)
}

// validateNestedStruct validates a nested struct and accumulates any violations.
func validateNestedStruct(depth int, val reflect.Value, violations *Violations) error {
	err := validateStruct(val.Interface(), depth+1)
	if err == nil {
		return nil
	}
	var structViolations *Violations
	if errors.As(err, &structViolations) {
		violations.AddViolations(structViolations)
		return nil
	}
	return err
}

// validateContainerElements validates elements within slices, arrays, and maps.
func validateContainerElements(depth int, val reflect.Value, violations *Violations) error {
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
func validateRecursively(depth int, val reflect.Value, violations *Violations) error {
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
// In the case that the struct has tag violations, a Violations error is returned.
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

	violations := NewViolations()
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
// In the case that the variable has tag violations, a Violations error is returned.
func Var[T any](val T, validatorInstructions string) error {
	reflectValue := reflect.ValueOf(val)
	violations := NewViolations()
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
