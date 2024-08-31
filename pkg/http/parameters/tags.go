package parameters

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	reflectutils "intelligence/pkg/utils/reflect"
)

// Tag is a string of metadata associated at compile time with a field of a struct.
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

	// lookupKeyTypeLocks ensures unique types are processed one at a time.
	lookupKeyTypeLocks = sync.Map{}

	// lookupKeyExtractMemo stores the results of the ExtractAndValidateFieldTagLookupKeys function.
	lookupKeyExtractMemo = make(map[reflect.Type]map[Tag]map[string]string)
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
//
//	type MyStruct struct {
//	    MyParameter string `httpHeader:x-my-parameter`
//	}
//
// Returns the following map:
//
//	{
//	   "httpHeader": {
//	       "x-my-parameter": "MyParameter"
//	   }
//	}
func ExtractAndValidateFieldTagLookupKeys[T any]() (map[Tag]map[string]string, error) {
	reflectType := reflect.TypeOf(*new(T))

	lock, _ := lookupKeyTypeLocks.LoadOrStore(reflectType, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	if memoData, isMemoized := lookupKeyExtractMemo[reflectType]; isMemoized {
		return memoData, nil
	}

	fieldsMetadata := reflectutils.FieldsToMetadata[T]()

	tagToLookupKeyToFieldName := make(map[Tag]map[string]string)
	for customTag := range tagToLookupKeyNormalizer {
		tagToLookupKeyToFieldName[customTag] = make(map[string]string)
	}

	for fieldName, fieldMetadata := range fieldsMetadata {
		customTagFound := false
		for customTag, lookupKeyNormalizer := range tagToLookupKeyNormalizer {
			originalLookupKeyForTag, customTagFoundOnField := fieldMetadata.Tags[string(customTag)]
			if !customTagFoundOnField {
				continue
			}

			if customTagFound {
				return nil, fmt.Errorf("there can only be one encoding tag on the field '%s'", fieldName)
			}
			customTagFound = true

			normalizedLookupKeyForTag := lookupKeyNormalizer(originalLookupKeyForTag)
			if !TagLookupKeyFollowsNamingConvention(normalizedLookupKeyForTag) {
				return nil, fmt.Errorf("tag '%s' with lookup key '%s' must adhere to the naming convention", customTag, originalLookupKeyForTag)
			}

			if _, lookupKeyAlreadySeenForTag := tagToLookupKeyToFieldName[customTag][normalizedLookupKeyForTag]; lookupKeyAlreadySeenForTag {
				return nil, fmt.Errorf("tag '%s' with lookup key '%s' is not unique", customTag, originalLookupKeyForTag)
			}
			tagToLookupKeyToFieldName[customTag][normalizedLookupKeyForTag] = fieldName

			if jsonTagValue, jsonTagFound := fieldMetadata.Tags[string(JSONTag)]; !jsonTagFound || jsonTagValue != "-" {
				return nil, fmt.Errorf("struct field '%s' with tag '%s' must have accompanying tag %s:\"-\"", fieldName, customTag, JSONTag)
			}
		}
	}

	lookupKeyExtractMemo[reflectType] = tagToLookupKeyToFieldName
	return tagToLookupKeyToFieldName, nil
}
