package parameters

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/TriangleSide/GoBase/pkg/utils/cache"
	reflectutils "github.com/TriangleSide/GoBase/pkg/utils/reflect"
)

// Tag is a string of metadata associated at compile time with a field of a struct.
//
//	type MyStruct struct {
//	    HeaderParameter string `httpHeader:"x-my-parameter"`
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
//	    HeaderParameter string `httpHeader:"x-my-parameter" json"-"`
//	}
//
// Returns the following map:
//
//	{
//	   "x-my-parameter": "MyParameter",
//	}
type LookupKeyToFieldName map[string]string

// TagToLookupKeyToFieldName is a map of unique tag lookup keys for each field in the struct.
//
//	type MyStruct struct {
//	    HeaderParameter string `httpHeader:"x-my-parameter" json"-"`
//		PathParameter   string `urlPath:"my-id" json"-"`
//	}
//
// Returns the following map:
//
//		{
//		   "httpHeader": {
//		       "x-my-parameter": "MyParameter"
//		   },
//	    "urlPath": {
//		       "my-id": "PathParameter"
//		   }
//		}
type TagToLookupKeyToFieldName map[Tag]LookupKeyToFieldName

var (
	// tagToLookupKeyNormalizer is a map of custom encoding tags to their string normalizers.
	tagToLookupKeyNormalizer = map[Tag]func(string) string{
		QueryTag:  strings.ToLower,
		HeaderTag: strings.ToLower,
		PathTag: func(s string) string {
			return s
		},
	}

	// lookupKeyFollowsNamingConvention is used to verify that a tags lookup key follow the naming convention as defined by TagLookupKeyNamingConvention.
	lookupKeyFollowsNamingConvention func(lookupKey string) bool

	// lookupKeyExtractionCache stores the results of the ExtractAndValidateFieldTagLookupKeys function.
	lookupKeyExtractionCache = cache.New[reflect.Type, TagToLookupKeyToFieldName]()
)

// init creates the variables needed by the processor.
func init() {
	lookupKeyFollowsNamingConvention = regexp.MustCompile(TagLookupKeyNamingConvention).MatchString
}

// TagLookupKeyFollowsNamingConvention verifies if the tag value (the lookup key) follows the naming convention.
func TagLookupKeyFollowsNamingConvention(lookupKey string) bool {
	return lookupKeyFollowsNamingConvention(lookupKey)
}

// ExtractAndValidateFieldTagLookupKeys validates the struct tags and returns a map of unique tag lookup keys for each field in the struct.
// The returned map should not be written to under any circumstances since it can be shared among many threads.
func ExtractAndValidateFieldTagLookupKeys[T any]() (TagToLookupKeyToFieldName, error) {
	reflectType := reflect.TypeOf(*new(T))
	return lookupKeyExtractionCache.GetOrSet(reflectType, func(reflectType reflect.Type) (TagToLookupKeyToFieldName, time.Duration, error) {
		fieldsMetadata := reflectutils.FieldsToMetadata[T]()

		tagToLookupKeyToFieldName := make(TagToLookupKeyToFieldName)
		for customTag := range tagToLookupKeyNormalizer {
			tagToLookupKeyToFieldName[customTag] = make(LookupKeyToFieldName)
		}

		for fieldName, fieldMetadata := range fieldsMetadata.Iterator() {
			customTagFound := false
			for customTag, lookupKeyNormalizer := range tagToLookupKeyNormalizer {
				originalLookupKeyForTag, customTagFoundOnField := fieldMetadata.Tags[string(customTag)]
				if !customTagFoundOnField {
					continue
				}

				if customTagFound {
					return nil, time.Duration(0), fmt.Errorf("there can only be one encoding tag on the field '%s'", fieldName)
				}
				customTagFound = true

				normalizedLookupKeyForTag := lookupKeyNormalizer(originalLookupKeyForTag)
				if !TagLookupKeyFollowsNamingConvention(normalizedLookupKeyForTag) {
					return nil, time.Duration(0), fmt.Errorf("tag '%s' with lookup key '%s' must adhere to the naming convention", customTag, originalLookupKeyForTag)
				}

				if _, lookupKeyAlreadySeenForTag := tagToLookupKeyToFieldName[customTag][normalizedLookupKeyForTag]; lookupKeyAlreadySeenForTag {
					return nil, time.Duration(0), fmt.Errorf("tag '%s' with lookup key '%s' is not unique", customTag, originalLookupKeyForTag)
				}
				tagToLookupKeyToFieldName[customTag][normalizedLookupKeyForTag] = fieldName

				if jsonTagValue, jsonTagFound := fieldMetadata.Tags[string(JSONTag)]; !jsonTagFound || jsonTagValue != "-" {
					return nil, time.Duration(0), fmt.Errorf("struct field '%s' with tag '%s' must have accompanying tag %s:\"-\"", fieldName, customTag, JSONTag)
				}
			}
		}

		return tagToLookupKeyToFieldName, cache.DoesNotExpire, nil
	})
}
