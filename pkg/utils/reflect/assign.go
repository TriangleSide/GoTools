package reflect

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// AssignToField sets a struct field specified by its name to a provided value encoded as a string.
// The function handles various data types including basic types (string, int, etc.),
// complex types (structs, slices, maps) and types implementing the encoding.TextUnmarshaler interface.
// The conversion from string to the appropriate type is performed based on the field's underlying type.
// JSON format is expected for complex types. This function supports setting both direct values and pointers to the values.
func AssignToField[T any](obj *T, fieldName string, stringEncodedValue string) error {
	structValue := reflect.ValueOf(obj)
	if structValue.Kind() != reflect.Ptr || structValue.Elem().Kind() != reflect.Struct {
		panic("obj must be a pointer to a struct")
	}

	// Get the field metadata for all the structs fields.
	fieldsToMetadata := FieldsToMetadata[T]()
	fieldMetadata, foundFieldMetadata := fieldsToMetadata.Fetch(fieldName)
	if !foundFieldMetadata {
		panic(fmt.Sprintf("no field '%s' in struct '%s'", fieldName, structValue.Type().String()))
	}

	// Get the value of the specified field in the struct.
	// This accounts for fields in embedded anonymous structs.
	var structFieldValue reflect.Value
	if len(fieldMetadata.Anonymous) != 0 {
		anonValue := structValue.Elem()
		for _, anonymousName := range fieldMetadata.Anonymous {
			anonValue = anonValue.FieldByName(anonymousName)
		}
		structFieldValue = anonValue.FieldByName(fieldName)
	} else {
		structFieldValue = structValue.Elem().FieldByName(fieldName)
	}

	// Get the struct field type. This is needed to determine how to set the value.
	originalFieldType := structFieldValue.Type()
	var fieldType reflect.Type
	if originalFieldType.Kind() == reflect.Ptr {
		fieldType = originalFieldType.Elem()
	} else {
		fieldType = originalFieldType
	}

	// fieldPtr is an allocated ptr to the raw type of the field to set the encoded value into.
	fieldPtr := reflect.New(fieldType)

	// Switch on how to set the value.
	if reflect.PointerTo(fieldType).Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
		// If the field type implements encoding.TextUnmarshaler, the interface is used parse the value.
		unmarshaler := fieldPtr.Interface().(encoding.TextUnmarshaler)
		if err := unmarshaler.UnmarshalText([]byte(stringEncodedValue)); err != nil {
			return fmt.Errorf("text unmarshall error (%s)", err.Error())
		}
	} else {
		// If the field type is basic, the value is set directly.
		// If the field type is map, slice, or struct, it is assumed that the value is a json object.
		switch fieldType.Kind() {
		case reflect.Map, reflect.Slice, reflect.Struct:
			if err := json.Unmarshal([]byte(stringEncodedValue), fieldPtr.Interface()); err != nil {
				return fmt.Errorf("json unmarshal error (%s)", err.Error())
			}
		case reflect.String:
			fieldPtr.Elem().SetString(stringEncodedValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			parsed, err := strconv.ParseInt(stringEncodedValue, 10, fieldType.Bits())
			if err != nil {
				return fmt.Errorf("int parsing error (%s)", err.Error())
			}
			fieldPtr.Elem().SetInt(parsed)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			parsed, err := strconv.ParseUint(stringEncodedValue, 10, fieldType.Bits())
			if err != nil {
				return fmt.Errorf("unsigned int parsing error (%s)", err.Error())
			}
			fieldPtr.Elem().SetUint(parsed)
		case reflect.Float32, reflect.Float64:
			parsed, err := strconv.ParseFloat(stringEncodedValue, fieldType.Bits())
			if err != nil {
				return fmt.Errorf("float parsing error (%s)", err.Error())
			}
			fieldPtr.Elem().SetFloat(parsed)
		case reflect.Bool:
			parsed, err := strconv.ParseBool(stringEncodedValue)
			if err != nil {
				return fmt.Errorf("bool parsing error (%s)", err.Error())
			}
			fieldPtr.Elem().SetBool(parsed)
		default:
			return fmt.Errorf("unsupported field type: %s", fieldType)
		}
	}

	// If the field is a ptr, set the ptr to the newly allocated value in fieldPtr.
	// If the field it not a ptr, copy the contents of fieldPtr into it.
	if originalFieldType.Kind() == reflect.Ptr {
		structFieldValue.Set(fieldPtr)
	} else {
		structFieldValue.Set(fieldPtr.Elem())
	}

	return nil
}
