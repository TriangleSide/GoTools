package structs

import (
	"fmt"
	"reflect"
	"regexp"

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
	getOrSetFn := func(reflectType reflect.Type) (*readonly.Map[string, *FieldMetadata], error) {
		fieldsToMetadata := make(map[string]*FieldMetadata)
		processType(reflectType, fieldsToMetadata, []string{})
		readOnlyMap := readonly.NewMapBuilder[string, *FieldMetadata]().SetMap(fieldsToMetadata).Build()
		return readOnlyMap, nil
	}
	fieldsToMetadata, _ := typeToMetadataCache.GetOrSet(reflectType, getOrSetFn)
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
		panic(fmt.Errorf("type must be a struct or a pointer to a struct, got %s", reflectType.Kind().String()))
	}

	for fieldIndex := range reflectType.NumField() {
		field := reflectType.Field(fieldIndex)

		anonymousChainCopy := append([]string{}, anonymousChain...)
		if field.Anonymous {
			anonymousChainCopy = append(anonymousChainCopy, field.Name)
			processType(field.Type, fieldsToMetadata, anonymousChainCopy)
			continue
		}

		if _, alreadyHasFieldName := fieldsToMetadata[field.Name]; alreadyHasFieldName {
			panic(fmt.Errorf("field %s is ambiguous", field.Name))
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

		fieldsToMetadata[field.Name] = &FieldMetadata{
			reflectType: field.Type,
			anonymous:   append([]string{}, anonymousChain...),
			tags:        tagBuilder.Build(),
		}
	}
}
