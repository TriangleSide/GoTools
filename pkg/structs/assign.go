package structs

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/TriangleSide/go-toolkit/pkg/reflection"
)

// getStructFieldValue retrieves the reflect.Value of a field, accounting for embedded anonymous structs.
func getStructFieldValue(structValue reflect.Value, fieldName string, fieldMetadata *FieldMetadata) reflect.Value {
	if len(fieldMetadata.Anonymous()) != 0 {
		anonValue := structValue.Elem()
		for _, anonymousName := range fieldMetadata.Anonymous() {
			anonValue = anonValue.FieldByName(anonymousName)
		}
		return anonValue.FieldByName(fieldName)
	}
	return structValue.Elem().FieldByName(fieldName)
}

// AssignToField sets a struct field specified by its name to a provided value encoded as a string.
// The function handles various data types including basic types (string, int, etc.),
// complex types (structs, slices, maps) and types implementing the encoding.TextUnmarshaler interface.
// The conversion from string to the appropriate type is performed based on the field's underlying type.
// JSON format is expected for complex types. This function supports setting both direct values
// and pointers to the values.
func AssignToField[T any](obj *T, fieldName string, stringEncodedValue string) error {
	structValue := reflect.ValueOf(obj)
	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		return errors.New("obj must be a pointer to a struct")
	}

	fieldsToMetadata := Metadata[T]()

	fieldMetadata, foundFieldMetadata := fieldsToMetadata[fieldName]
	if !foundFieldMetadata {
		return fmt.Errorf("no field '%s' in struct '%s'", fieldName, structValue.Type().String())
	}

	structFieldValue := getStructFieldValue(structValue, fieldName, fieldMetadata)

	if err := reflection.AssignFromString(structFieldValue, stringEncodedValue); err != nil {
		return fmt.Errorf("failed to assign value '%s' to field '%s': %w", stringEncodedValue, fieldName, err)
	}

	return nil
}
