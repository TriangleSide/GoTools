package reflect

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/TriangleSide/GoBase/pkg/datastructures"
	"github.com/TriangleSide/GoBase/pkg/utils/cache"
)

var (
	// tagMatchRegex matches all tag entries on a struct field.
	tagMatchRegex = regexp.MustCompile(`(\w+):"([^"]*)"`)

	// typeToMetadataCache is used to cache the result of the FieldsToMetadata function.
	typeToMetadataCache = cache.New[reflect.Type, datastructures.ReadOnlyMap[string, *FieldMetadata]]()
)

// FieldMetadata is the metadata extracted from struct fields.
type FieldMetadata struct {
	Type      reflect.Type
	Tags      map[string]string
	Anonymous []string
}

// FieldsToMetadata returns a map of a structs field names to their respective metadata.
func FieldsToMetadata[T any]() datastructures.ReadOnlyMap[string, *FieldMetadata] {
	reflectType := reflect.TypeOf(*new(T))
	fieldsToMetadata, _ := typeToMetadataCache.GetOrSet(reflectType, func(reflectType reflect.Type) (datastructures.ReadOnlyMap[string, *FieldMetadata], time.Duration, error) {
		fieldsToMetadata := make(map[string]*FieldMetadata)
		processType(reflectType, fieldsToMetadata, make([]string, 0))
		readOnlyMap := datastructures.NewReadOnlyMapBuilder[string, *FieldMetadata]().SetMap(fieldsToMetadata).Build()
		return readOnlyMap, cache.DoesNotExpire, nil
	})
	return fieldsToMetadata
}

// processType takes a struct type, lists all of its fields, and builds the metadata for it.
// If the struct contains an embedded anonymous struct, it appends its name to the anonymous name chain.
// If a field name is not unique, a panic occurs. This includes field names of the anonymous structs.
func processType(reflectType reflect.Type, fieldsToMetadata map[string]*FieldMetadata, anonymousChain []string) {
	if reflectType.Kind() != reflect.Struct {
		panic("type must be a struct")
	}

	for fieldIndex := 0; fieldIndex < reflectType.NumField(); fieldIndex++ {
		field := reflectType.Field(fieldIndex)

		anonymousChainCopy := make([]string, len(anonymousChain))
		copy(anonymousChainCopy, anonymousChain)

		if field.Anonymous {
			anonymousChainCopy = append(anonymousChainCopy, field.Name)
			processType(field.Type, fieldsToMetadata, anonymousChainCopy)
			continue
		}

		if _, alreadyHasFieldName := fieldsToMetadata[field.Name]; alreadyHasFieldName {
			panic(fmt.Sprintf("field %s is ambiguous", field.Name))
		}

		metadata := &FieldMetadata{}
		metadata.Type = field.Type
		metadata.Tags = make(map[string]string)
		metadata.Anonymous = anonymousChainCopy

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
}
