package structs

import (
	"errors"
	"fmt"
	"reflect"
)

// ValueFromName returns the fields value if it exists.
func ValueFromName[T any](structInstance T, fieldName string) (reflect.Value, error) {
	reflectVal := reflect.ValueOf(structInstance)
	if reflectVal.Kind() == reflect.Ptr {
		if reflectVal.IsNil() {
			return reflect.Value{}, errors.New("struct instance cannot be nil")
		}
		reflectVal = reflectVal.Elem()
	}
	if reflectVal.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("type must be a struct or a pointer to a struct")
	}
	field := reflectVal.FieldByName(fieldName)
	if !field.IsValid() {
		return reflect.Value{}, fmt.Errorf("field %s does not exist in the struct", fieldName)
	}
	return field, nil
}
