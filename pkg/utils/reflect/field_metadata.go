package reflect

import (
	"reflect"
	"regexp"
	"sync"
)

var (
	// fieldsToMetadataMemo is used to memoize the result of the FieldsToMetadata function.
	fieldsToMetadataMemo = sync.Map{}

	// tagMatchRegex matches all tag entries on a struct field.
	tagMatchRegex = regexp.MustCompile(`(\w+):"([^"]*)"`)
)

// FieldMetadata is the metadata extracted from struct fields.
type FieldMetadata struct {
	Type reflect.Type
	Tags map[string]string
}

// FieldsToMetadata returns a map of a structs field names to their respective metadata.
func FieldsToMetadata[T any]() (map[string]*FieldMetadata, error) {
	reflectType := reflect.TypeOf(*new(T))
	if fieldsToMetadata, ok := fieldsToMetadataMemo.Load(reflectType); ok {
		return fieldsToMetadata.(map[string]*FieldMetadata), nil
	}

	if reflectType.Kind() != reflect.Struct {
		panic("type must be a struct")
	}
	fieldsToMetadata := make(map[string]*FieldMetadata)

	for fieldIndex := 0; fieldIndex < reflectType.NumField(); fieldIndex++ {
		field := reflectType.Field(fieldIndex)
		metadata := &FieldMetadata{}

		metadata.Type = field.Type
		metadata.Tags = make(map[string]string)

		if len(string(field.Tag)) != 0 {
			matches := tagMatchRegex.FindAllStringSubmatch(string(field.Tag), -1)
			for _, match := range matches {
				tagKey := match[1]
				tagValue := match[2]
				metadata.Tags[tagKey] = tagValue
			}
		}

		fieldsToMetadata[field.Name] = metadata
	}

	fieldsToMetadataMemo.Store(reflectType, fieldsToMetadata)
	return fieldsToMetadata, nil
}
