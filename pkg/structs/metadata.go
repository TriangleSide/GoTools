package structs

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/TriangleSide/GoTools/pkg/datastructures/cache"
	"github.com/TriangleSide/GoTools/pkg/datastructures/readonly"
)

var (
	// tagMatchRegex matches all tag entries on a struct field.
	tagMatchRegex = regexp.MustCompile(`(\w+):"([^"]*)"`)

	// typeToMetadataCache is used to cache the result of the Metadata function.
	typeToMetadataCache = cache.New[reflect.Type, *readonly.Map[string, *FieldMetadata]]()
)

// Metadata returns a map of a struct's field names to their respective metadata.
func Metadata[T any]() *readonly.Map[string, *FieldMetadata] {
	return MetadataFromType(reflect.TypeFor[T]())
}

// MetadataFromType returns a map of a struct's field names to their respective metadata.
func MetadataFromType(reflectType reflect.Type) *readonly.Map[string, *FieldMetadata] {
	fieldsToMetadata, _ := typeToMetadataCache.GetOrSet(reflectType, func(reflectType reflect.Type) (*readonly.Map[string, *FieldMetadata], *time.Duration, error) {
		fieldsToMetadata := make(map[string]*FieldMetadata)
		processType(reflectType, fieldsToMetadata, make([]string, 0))
		readOnlyMap := readonly.NewMapBuilder[string, *FieldMetadata]().SetMap(fieldsToMetadata).Build()
		return readOnlyMap, nil, nil
	})
	return fieldsToMetadata
}

// processType takes a struct type, lists all of its fields, and builds the metadata for it.
// If the struct contains an embedded anonymous struct, it appends its name to the anonymous name chain.
// If a field name is not unique, a panic occurs. This includes field names of the anonymous structs.
func processType(reflectType reflect.Type, fieldsToMetadata map[string]*FieldMetadata, anonymousChain []string) {
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	if reflectType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Type must be a struct or a pointer to a struct but got %s.", reflectType.Kind().String()))
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

		tagBuilder := readonly.NewMapBuilder[string, string]()
		if len(string(field.Tag)) != 0 {
			matches := tagMatchRegex.FindAllStringSubmatch(string(field.Tag), -1)
			for _, match := range matches {
				tagBuilder.Set(readonly.MapEntry[string, string]{
					Key:   match[1],
					Value: match[2],
				})
			}
		}

		anonymousChainBuilder := readonly.NewSliceBuilder[string]()
		anonymousChainBuilder.Append(anonymousChain...)

		fieldsToMetadata[field.Name] = &FieldMetadata{
			reflectType: field.Type,
			anonymous:   anonymousChainBuilder.Build(),
			tags:        tagBuilder.Build(),
		}
	}
}
