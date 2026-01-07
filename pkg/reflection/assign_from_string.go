package reflection

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	// kindToAssignHandlers maps kinds to their string parsing handlers.
	kindToAssignHandlers = map[reflect.Kind]func(fieldPtr reflect.Value, stringEncodedValue string) error{
		reflect.Int:        setStringIntoInt,
		reflect.Int8:       setStringIntoInt,
		reflect.Int16:      setStringIntoInt,
		reflect.Int32:      setStringIntoInt,
		reflect.Int64:      setStringIntoInt,
		reflect.Uint:       setStringIntoUint,
		reflect.Uint8:      setStringIntoUint,
		reflect.Uint16:     setStringIntoUint,
		reflect.Uint32:     setStringIntoUint,
		reflect.Uint64:     setStringIntoUint,
		reflect.Float32:    setStringIntoFloat,
		reflect.Float64:    setStringIntoFloat,
		reflect.Bool:       setStringIntoBool,
		reflect.Complex64:  setStringIntoComplex,
		reflect.Complex128: setStringIntoComplex,
		reflect.Map:        setStringIntoJSONType,
		reflect.Slice:      setStringIntoJSONType,
		reflect.Struct:     setStringIntoJSONType,
		reflect.Array:      setStringIntoJSONType,
	}
)

// setStringIntoJSONType handles Map, Slice, Array, and Struct types by JSON unmarshaling.
func setStringIntoJSONType(fieldPtr reflect.Value, stringEncodedValue string) error {
	if err := json.Unmarshal([]byte(stringEncodedValue), fieldPtr.Interface()); err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}
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

// setStringIntoComplex handles complex types (Complex64, Complex128).
func setStringIntoComplex(fieldPtr reflect.Value, stringEncodedValue string) error {
	bitSize := 64
	if fieldPtr.Elem().Type().Kind() == reflect.Complex128 {
		bitSize = 128
	}
	parsed, err := strconv.ParseComplex(stringEncodedValue, bitSize)
	if err != nil {
		return fmt.Errorf("complex parsing error: %w", err)
	}
	fieldPtr.Elem().SetComplex(parsed)
	return nil
}

// AssignFromString parses a string-encoded value and sets it into the provided reflect.Value.
// The value must be settable (typically obtained via reflect.ValueOf(&x).Elem() or a struct field).
// The function handles various data types including basic types (string, int, bool, float, complex),
// complex types (structs, slices, arrays, maps) and types implementing the encoding.TextUnmarshaler interface.
// The conversion from string to the appropriate type is performed based on the underlying type.
// JSON format is expected for complex types like maps, slices, arrays, and structs.
// If the provided value is a pointer, the function allocates a new value and assigns it.
// An error is returned for unsupported types such as Chan, Func, Interface, UnsafePointer, and Uintptr.
func AssignFromString(value reflect.Value, stringEncodedValue string) error {
	if !value.IsValid() {
		return errors.New("value is not valid")
	}

	if !value.CanSet() {
		return errors.New("value is not settable")
	}

	valueType := value.Type()

	if valueType.Kind() == reflect.Ptr {
		newValue := reflect.New(valueType.Elem())
		if err := AssignFromString(newValue.Elem(), stringEncodedValue); err != nil {
			return err
		}
		value.Set(newValue)
		return nil
	}

	valuePtr := value.Addr()
	if reflect.PointerTo(valueType).Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
		unmarshaler := valuePtr.Interface().(encoding.TextUnmarshaler)
		if err := unmarshaler.UnmarshalText([]byte(stringEncodedValue)); err != nil {
			return fmt.Errorf("text unmarshal error: %w", err)
		}
		return nil
	}

	if valueType.Kind() == reflect.String {
		value.SetString(stringEncodedValue)
		return nil
	}

	handler, ok := kindToAssignHandlers[valueType.Kind()]
	if !ok {
		return fmt.Errorf("unsupported type: %s", valueType)
	}
	if err := handler(valuePtr, stringEncodedValue); err != nil {
		return fmt.Errorf("failed to set value to type %q: %w", valueType.String(), err)
	}
	return nil
}
