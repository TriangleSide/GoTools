package structs

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/TriangleSide/go-toolkit/pkg/datastructures/cache"
	"github.com/TriangleSide/go-toolkit/pkg/reflection"
)

var (
	// tagMatchRegex matches all tag entries on a struct field.
	tagMatchRegex = regexp.MustCompile(`(\w+):"([^"]*)"`)

	// typeToMetadataCache is used to cache the result of the Metadata function.
	typeToMetadataCache = cache.New[reflect.Type, map[string]*FieldMetadata]()
)

// Metadata returns a map of a struct's field names to their respective metadata.
func Metadata[T any]() map[string]*FieldMetadata {
	return MetadataFromType(reflect.TypeFor[T]())
}

// MetadataFromType returns a map of a struct's field names to their respective metadata.
func MetadataFromType(reflectType reflect.Type) map[string]*FieldMetadata {
	getOrSetFn := func(reflectType reflect.Type) (map[string]*FieldMetadata, error) {
		fieldsToMetadata := make(map[string]*FieldMetadata)
		processType(reflectType, fieldsToMetadata, []string{})
		return fieldsToMetadata, nil
	}
	fieldsToMetadata, _ := typeToMetadataCache.GetOrSet(reflectType, getOrSetFn)
	return fieldsToMetadata
}

// processType takes a struct type, lists all of its fields, and builds the metadata for it.
// If the struct contains an embedded anonymous struct, it appends its name to the anonymous name chain.
// If a field name is not unique, a panic occurs. This includes field names of the anonymous structs.
func processType(reflectType reflect.Type, fieldsToMetadata map[string]*FieldMetadata, anonymousChain []string) {
	reflectType = reflection.DereferenceType(reflectType)
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

		tags := make(map[string]string)

		if len(string(field.Tag)) != 0 {
			matches := tagMatchRegex.FindAllStringSubmatch(string(field.Tag), -1)
			for _, match := range matches {
				tags[match[1]] = match[2]
			}
		}

		fieldsToMetadata[field.Name] = &FieldMetadata{
			reflectType: field.Type,
			anonymous:   append([]string{}, anonymousChain...),
			tags:        tags,
		}
	}
}
