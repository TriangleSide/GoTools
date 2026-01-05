package structs

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	// typeToAssignHandlers maps primitive types to their string parsing handlers.
	typeToAssignHandlers = map[reflect.Kind]func(fieldPtr reflect.Value, stringEncodedValue string) error{
		reflect.String:  setStringIntoString,
		reflect.Int:     setStringIntoInt,
		reflect.Int8:    setStringIntoInt,
		reflect.Int16:   setStringIntoInt,
		reflect.Int32:   setStringIntoInt,
		reflect.Int64:   setStringIntoInt,
		reflect.Uint:    setStringIntoUint,
		reflect.Uint8:   setStringIntoUint,
		reflect.Uint16:  setStringIntoUint,
		reflect.Uint32:  setStringIntoUint,
		reflect.Uint64:  setStringIntoUint,
		reflect.Float32: setStringIntoFloat,
		reflect.Float64: setStringIntoFloat,
		reflect.Bool:    setStringIntoBool,
		reflect.Map:     setStringIntoJSONType,
		reflect.Slice:   setStringIntoJSONType,
		reflect.Struct:  setStringIntoJSONType,
	}
)

// setStringIntoJSONType handles Map, Slice, and Struct types by JSON unmarshaling.
func setStringIntoJSONType(fieldPtr reflect.Value, stringEncodedValue string) error {
	if err := json.Unmarshal([]byte(stringEncodedValue), fieldPtr.Interface()); err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}
	return nil
}

// setStringIntoString handles string types.
func setStringIntoString(fieldPtr reflect.Value, stringEncodedValue string) error {
	fieldPtr.Elem().SetString(stringEncodedValue)
	return nil
}

// setStringIntoInt handles signed integer types (Int, Int8, Int16, Int32, Int64).
func setStringIntoInt(fieldPtr reflect.Value, stringEncodedValue string) error {
	parsed, err := strconv.ParseInt(stringEncodedValue, 10, fieldPtr.Elem().Type().Bits())
	if err != nil {
		return fmt.Errorf("int parsing error: %w", err)
	}
	fieldPtr.Elem().SetInt(parsed)
	return nil
}

// setStringIntoUint handles unsigned integer types (Uint, Uint8, Uint16, Uint32, Uint64).
func setStringIntoUint(fieldPtr reflect.Value, stringEncodedValue string) error {
	parsed, err := strconv.ParseUint(stringEncodedValue, 10, fieldPtr.Elem().Type().Bits())
	if err != nil {
		return fmt.Errorf("unsigned int parsing error: %w", err)
	}
	fieldPtr.Elem().SetUint(parsed)
	return nil
}

// setStringIntoFloat handles floating point types (Float32, Float64).
func setStringIntoFloat(fieldPtr reflect.Value, stringEncodedValue string) error {
	parsed, err := strconv.ParseFloat(stringEncodedValue, fieldPtr.Elem().Type().Bits())
	if err != nil {
		return fmt.Errorf("float parsing error: %w", err)
	}
	fieldPtr.Elem().SetFloat(parsed)
	return nil
}

// setStringIntoBool handles boolean types.
func setStringIntoBool(fieldPtr reflect.Value, stringEncodedValue string) error {
	parsed, err := strconv.ParseBool(stringEncodedValue)
	if err != nil {
		return fmt.Errorf("bool parsing error: %w", err)
	}
	fieldPtr.Elem().SetBool(parsed)
	return nil
}

func setStringIntoTextUnmarshaler(fieldPtr reflect.Value, stringEncodedValue string) (bool, error) {
	if reflect.PointerTo(fieldPtr.Elem().Type()).Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
		unmarshaler := fieldPtr.Interface().(encoding.TextUnmarshaler)
		if err := unmarshaler.UnmarshalText([]byte(stringEncodedValue)); err != nil {
			return false, fmt.Errorf("text unmarshal error: %w", err)
		}
		return true, nil
	}
	return false, nil
}

// setStringIntoField parses a string-encoded value and sets it into the provided fieldPtr based on its type.
func setStringIntoField(fieldPtr reflect.Value, fieldType reflect.Type, stringEncodedValue string) error {
	testUnmarshalHandled, textUnmarshalErr := setStringIntoTextUnmarshaler(fieldPtr, stringEncodedValue)
	if textUnmarshalErr != nil {
		return textUnmarshalErr
	}
	if testUnmarshalHandled {
		return nil
	}

	if handler, ok := typeToAssignHandlers[fieldType.Kind()]; ok {
		err := handler(fieldPtr, stringEncodedValue)
		if err != nil {
			return fmt.Errorf("failed to set value to type %q: %w", fieldType.String(), err)
		}
		return nil
	}

	return fmt.Errorf("unsupported field type: %s", fieldType)
}

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
	originalFieldType := structFieldValue.Type()

	fieldType := originalFieldType
	if originalFieldType.Kind() == reflect.Ptr {
		fieldType = originalFieldType.Elem()
	}

	fieldPtr := reflect.New(fieldType)
	if err := setStringIntoField(fieldPtr, fieldType, stringEncodedValue); err != nil {
		return err
	}

	if originalFieldType.Kind() == reflect.Ptr {
		structFieldValue.Set(fieldPtr)
	} else {
		structFieldValue.Set(fieldPtr.Elem())
	}

	return nil
}
