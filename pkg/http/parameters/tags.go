package parameters

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/TriangleSide/GoTools/pkg/datastructures/cache"
	"github.com/TriangleSide/GoTools/pkg/datastructures/readonly"
	"github.com/TriangleSide/GoTools/pkg/structs"
)

// Tag is a string of metadata associated at compile time with a field of a struct.
//
//	type MyStruct struct {
//		HeaderParameter string `httpHeader:"x-my-parameter"`
//	}
//
// In this case, the tag would be "httpHeader".
type Tag string

const (
	// QueryTag is a struct field tag used to specify that the field's value should be sourced from URL query parameters.
	QueryTag Tag = "urlQuery"

	// HeaderTag is a struct field tag used to specify that the field's value should be sourced from the HTTP headers.
	HeaderTag Tag = "httpHeader"

	// PathTag is a struct field tag used to specify that the field's value should be sourced from the URL path parameters.
	PathTag Tag = "urlPath"

	// JSONTag is a struct field tag used to specify that the field's value should be sourced from the request JSON body.
	JSONTag Tag = "json"

	// TagLookupKeyNamingConvention is the naming convention a tags lookup key must adhere to.
	TagLookupKeyNamingConvention = `^[a-zA-Z][a-zA-Z0-9_-]*$`
)

// LookupKeyToFieldName is the tag's lookup key to the name of the field on the struct.
//
//	type MyStruct struct {
//		HeaderParameter string `httpHeader:"x-my-parameter" json:"-"`
//	}
//
// Returns the following map:
//
//	{
//		"x-my-parameter": "MyParameter",
//	}
type LookupKeyToFieldName map[string]string

var (
	// tagToLookupKeyNormalizer is a map of custom encoding tags to their string normalizers.
	tagToLookupKeyNormalizer = map[Tag]func(string) string{
		QueryTag:  strings.ToLower,
		HeaderTag: strings.ToLower,
		PathTag: func(s string) string {
			return s
		},
	}

	// lookupKeyFollowsNamingConvention verifies that a tag's lookup key follows
	// the naming convention defined by TagLookupKeyNamingConvention.
	lookupKeyFollowsNamingConvention func(lookupKey string) bool

	// lookupKeyExtractionCache stores the results of the ExtractAndValidateFieldTagLookupKeys function.
	lookupKeyExtractionCache = cache.New[reflect.Type, *readonly.Map[Tag, LookupKeyToFieldName]]()
)

// init creates the variables needed by the processor.
func init() {
	lookupKeyFollowsNamingConvention = regexp.MustCompile(TagLookupKeyNamingConvention).MatchString
}

// TagLookupKeyFollowsNamingConvention verifies if the tag value (the lookup key) follows the naming convention.
func TagLookupKeyFollowsNamingConvention(lookupKey string) bool {
	return lookupKeyFollowsNamingConvention(lookupKey)
}

// buildFieldTagLookupKeys extracts and validates field tag lookup keys for type T.
func buildFieldTagLookupKeys[T any](reflect.Type) (*readonly.Map[Tag, LookupKeyToFieldName], *time.Duration, error) {
	fieldsMetadata := structs.Metadata[T]()

	tagToLookupKeyToFieldName := make(map[Tag]LookupKeyToFieldName)
	for customTag := range tagToLookupKeyNormalizer {
		tagToLookupKeyToFieldName[customTag] = make(LookupKeyToFieldName)
	}

	for fieldName, fieldMetadata := range fieldsMetadata.All() {
		customTagFound := false
		for customTag, lookupKeyNormalizer := range tagToLookupKeyNormalizer {
			originalLookupKeyForTag, customTagFoundOnField := fieldMetadata.Tags().Fetch(string(customTag))
			if !customTagFoundOnField {
				continue
			}

			if customTagFound {
				return nil, nil, fmt.Errorf("there can only be one encoding tag on the field '%s'", fieldName)
			}
			customTagFound = true

			normalizedLookupKeyForTag := lookupKeyNormalizer(originalLookupKeyForTag)
			if !TagLookupKeyFollowsNamingConvention(normalizedLookupKeyForTag) {
				return nil, nil, fmt.Errorf(
					"tag '%s' with lookup key '%s' must adhere to the naming convention",
					customTag, originalLookupKeyForTag)
			}

			_, lookupKeyAlreadySeenForTag := tagToLookupKeyToFieldName[customTag][normalizedLookupKeyForTag]
			if lookupKeyAlreadySeenForTag {
				return nil, nil, fmt.Errorf(
					"tag '%s' with lookup key '%s' is not unique",
					customTag, originalLookupKeyForTag)
			}
			tagToLookupKeyToFieldName[customTag][normalizedLookupKeyForTag] = fieldName

			jsonTagValue, jsonTagFound := fieldMetadata.Tags().Fetch(string(JSONTag))
			if !jsonTagFound || jsonTagValue != "-" {
				return nil, nil, fmt.Errorf(
					"struct field '%s' with tag '%s' must have accompanying tag %s:\"-\"",
					fieldName, customTag, JSONTag)
			}
		}
	}

	return readonly.NewMapBuilder[Tag, LookupKeyToFieldName]().SetMap(tagToLookupKeyToFieldName).Build(), nil, nil
}

// ExtractAndValidateFieldTagLookupKeys validates the struct tags and returns a map
// of unique tag lookup keys for each field in the struct.
//
//	type MyStruct struct {
//		HeaderParameter string `httpHeader:"x-my-parameter" json:"-"`
//		PathParameter   string `urlPath:"my-id" json:"-"`
//	}
//
// Returns the following map:
//
//	{
//		"httpHeader": {
//			"x-my-parameter": "HeaderParameter"
//		},
//		"urlPath": {
//			"my-id": "PathParameter"
//		}
//	}
func ExtractAndValidateFieldTagLookupKeys[T any]() (*readonly.Map[Tag, LookupKeyToFieldName], error) {
	reflectType := reflect.TypeFor[T]()
	result, err := lookupKeyExtractionCache.GetOrSet(reflectType, buildFieldTagLookupKeys[T])
	if err != nil {
		return nil, fmt.Errorf("failed to extract field tag lookup keys (%w)", err)
	}
	return result, nil
}
