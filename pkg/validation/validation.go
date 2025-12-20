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

// ValidatorWithInstruction holds a parsed validator name along with its optional instruction string.
// For example, given the validation tag segment "oneof=ACTIVE INACTIVE", this struct would contain:
//   - Name: "oneof"
//   - Instruction: "ACTIVE INACTIVE"
//
// For a validator without instructions like "required", this struct would contain:
//   - Name: "required"
//   - Instruction: "" (empty string)
type ValidatorWithInstruction struct {
	Name        Validator
	Instruction string
}

// parseValidatorWithInstruction takes a single validator segment from a validation tag and parses it
// into a ValidatorWithInstruction struct.
//
// Examples:
//   - "required" -> ValidatorWithInstruction{Name: "required", Instruction: ""}
//   - "oneof=A B C" -> ValidatorWithInstruction{Name: "oneof", Instruction: "A B C"}
//   - "gt=10" -> ValidatorWithInstruction{Name: "gt", Instruction: "10"}
func parseValidatorWithInstruction(segment string) (ValidatorWithInstruction, error) {
	const maxParts = 2
	parts := strings.Split(segment, NameAndInstructionsSep)
	if len(parts) > maxParts {
		return ValidatorWithInstruction{}, errors.New("malformed validator and instruction")
	}
	instruction := ""
	if len(parts) == maxParts {
		instruction = parts[1]
	}
	return ValidatorWithInstruction{
		Name:        Validator(parts[0]),
		Instruction: instruction,
	}, nil
}

// expandAliases takes a validation tag string and expands any validator aliases into their full
// definitions. Aliases are shorthand names that expand into one or more validators.
//
// For example, if "nonempty" is registered as an alias for "required,min=1":
//   - Input: "nonempty,max=100"
//   - Output: "required,min=1,max=100"
func expandAliases(validateTagContents string) string {
	segments := strings.Split(validateTagContents, ValidatorsSep)
	var result []string

	for _, segment := range segments {
		parsed, parseErr := parseValidatorWithInstruction(segment)
		if parseErr != nil {
			result = append(result, segment)
			continue
		}

		if expansion, isAlias := lookupAlias(string(parsed.Name)); isAlias {
			result = append(result, expansion)
		} else {
			result = append(result, segment)
		}
	}

	return strings.Join(result, ValidatorsSep)
}

// parseValidationTag takes a complete validation tag string and parses it into a slice of
// ValidatorWithInstruction structs. The tag string is first expanded to resolve any aliases,
// then each segment is parsed into its name and instruction components.
//
// Example:
//   - Input: "required,oneof=A B,gt=0"
//   - Output: []ValidatorWithInstruction{
//     {Name: "required", Instruction: ""},
//     {Name: "oneof", Instruction: "A B"},
//     {Name: "gt", Instruction: "0"},
//     }
func parseValidationTag(validateTagContents string) ([]ValidatorWithInstruction, error) {
	if strings.TrimSpace(validateTagContents) == "" {
		return nil, fmt.Errorf("empty %s instructions", Tag)
	}

	expandedTag := expandAliases(validateTagContents)
	segments := strings.Split(expandedTag, ValidatorsSep)
	validators := make([]ValidatorWithInstruction, 0, len(segments))

	for _, segment := range segments {
		parsed, err := parseValidatorWithInstruction(segment)
		if err != nil {
			return nil, err
		}
		validators = append(validators, parsed)
	}

	return validators, nil
}

// iterateValidators iterates over a slice of parsed validators and invokes the callback for each one.
// The callback receives the current validator along with a slice of the remaining validators that
// have not yet been processed. This remaining slice is useful for validators like "dive" that need
// to apply subsequent validators to nested elements.
//
// The callback returns two values:
//   - shouldContinue: if true, continue to the next validator; if false, stop iteration early
//   - error: if non-nil, stop iteration and return this error
//
// Example iteration for validators [required, dive, gt=0]:
//   - First call: current={required, ""}, remaining=[{dive, ""}, {gt, "0"}]
//   - Second call: current={dive, ""}, remaining=[{gt, "0"}]
//   - Third call: current={gt, "0"}, remaining=[]
func iterateValidators(
	validators []ValidatorWithInstruction,
	callback func(current ValidatorWithInstruction, remaining []ValidatorWithInstruction) (bool, error),
) error {
	for i, validator := range validators {
		remaining := validators[i+1:]
		shouldContinue, err := callback(validator, remaining)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}
	}
	return nil
}

// executeValidator runs a single validator callback and processes its result. It looks up the
// registered callback for the validator, invokes it, and handles the response which may indicate
// a field error, a request to stop validation, new values to validate recursively, or a pass.
//
// Example:
//   - Validator: {gt, "5"}
//   - Looks up the "gt" callback and invokes it with instruction "5" to check if value > 5
func executeValidator(
	validator ValidatorWithInstruction,
	remainingValidators []ValidatorWithInstruction,
	isStructValue bool,
	structValue reflect.Value,
	structFieldName string,
	fieldValue reflect.Value,
) ([]*FieldError, bool, error) {
	callbackNotCast, callbackFound := registeredValidations.Load(string(validator.Name))
	if !callbackFound {
		return nil, false, fmt.Errorf("validation with name '%s' is not registered", validator.Name)
	}

	callback := callbackNotCast.(Callback)
	callbackParameters := &CallbackParameters{
		Validator:          validator.Name,
		IsStructValidation: isStructValue,
		StructValue:        structValue,
		StructFieldName:    structFieldName,
		Value:              fieldValue,
		Parameters:         validator.Instruction,
	}

	callbackResponse, callbackErr := callback(callbackParameters)
	if callbackErr != nil {
		return nil, false, callbackErr
	}

	if callbackResponse == nil {
		return nil, false, fmt.Errorf("callback returned nil result for validator %s", validator.Name)
	}

	if len(callbackResponse.fieldErrors) > 0 {
		return callbackResponse.fieldErrors, false, nil
	}

	if callbackResponse.stop {
		return nil, false, nil
	}

	if callbackResponse.newValues != nil {
		var fieldErrors []*FieldError
		for _, newValue := range callbackResponse.newValues {
			newFieldErrors, err := runValidatorsAgainstValue(
				remainingValidators, isStructValue, structValue, structFieldName, newValue)
			if err != nil {
				return nil, false, err
			}
			fieldErrors = append(fieldErrors, newFieldErrors...)
		}
		return fieldErrors, false, nil
	}

	if callbackResponse.pass {
		return nil, true, nil
	}

	return nil, false, fmt.Errorf("callback response is not correctly filled for validator %s", validator.Name)
}

// runValidatorsAgainstValue takes a pre-parsed slice of validators and runs them against a value.
// It iterates through each validator, executes it, and accumulates any field errors.
//
// Example:
//   - Validators: [{required, ""}, {gt, "0"}]
//   - Checks that the value is non-zero, then checks that it is greater than 0
func runValidatorsAgainstValue(
	validators []ValidatorWithInstruction,
	isStructValue bool,
	structValue reflect.Value,
	structFieldName string,
	fieldValue reflect.Value,
) ([]*FieldError, error) {
	if len(validators) == 0 {
		return nil, fmt.Errorf("empty %s instructions", Tag)
	}

	var allFieldErrors []*FieldError

	iterCallback := func(current ValidatorWithInstruction, remaining []ValidatorWithInstruction) (bool, error) {
		fieldErrors, shouldContinue, execErr := executeValidator(
			current, remaining, isStructValue, structValue, structFieldName, fieldValue)
		if execErr != nil {
			return false, execErr
		}
		allFieldErrors = append(allFieldErrors, fieldErrors...)
		return shouldContinue, nil
	}
	err := iterateValidators(validators, iterCallback)

	return allFieldErrors, err
}

// checkValidatorsAgainstValue parses a validation tag string and runs the validators against the
// provided value. This is the main entry point for validating a value using a tag string.
//
// Example:
//   - Tag: "required,gt=0,lt=100"
//   - Validates that the value is non-zero, greater than 0, and less than 100
func checkValidatorsAgainstValue(
	isStructValue bool,
	structValue reflect.Value,
	structFieldName string,
	fieldValue reflect.Value,
	validationTagContents string,
) ([]*FieldError, error) {
	validators, err := parseValidationTag(validationTagContents)
	if err != nil {
		return nil, err
	}
	return runValidatorsAgainstValue(validators, isStructValue, structValue, structFieldName, fieldValue)
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
