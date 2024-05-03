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
// The returned map should not be written to under any circumstances since it can be shared among many threads.
func FieldsToMetadata[T any]() map[string]*FieldMetadata {
	reflectType := reflect.TypeOf(*new(T))
	if memoData, ok := fieldsToMetadataMemo.Load(reflectType); ok {
		return memoData.(map[string]*FieldMetadata)
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
	return fieldsToMetadata
}
