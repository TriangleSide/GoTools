package parameters

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
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
		QueryTag: func(s string) string {
			return strings.ToLower(s)
		},
		HeaderTag: func(s string) string {
			return strings.ToLower(s)
		},
		PathTag: func(s string) string {
			return s
		},
	}

	// lookupKeyFollowsNamingConvention is used to verify that a tags lookup key follow the naming convention as defined by TagLookupKeyNamingConvention.
	lookupKeyFollowsNamingConvention func(lookupKey string) bool
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
func ExtractAndValidateFieldTagLookupKeys[T any]() (map[Tag]map[string]reflect.StructField, error) {
	paramsType := reflect.TypeOf(new(T)).Elem()
	if paramsType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("the generic must be a struct")
	}

	tagToLookupKeyToFieldName := map[Tag]map[string]reflect.StructField{}
	for customTag := range tagToLookupKeyNormalizer {
		tagToLookupKeyToFieldName[customTag] = map[string]reflect.StructField{}
	}

	for fieldIndex := 0; fieldIndex < paramsType.NumField(); fieldIndex++ {
		field := paramsType.Field(fieldIndex)

		customTagFound := false
		for customTag, lookupKeyNormalizer := range tagToLookupKeyNormalizer {
			originalLookupKeyForTag, customTagFoundOnField := field.Tag.Lookup(string(customTag))
			if !customTagFoundOnField {
				continue
			}
			normalizedLookupKeyForTag := lookupKeyNormalizer(originalLookupKeyForTag)

			if customTagFound {
				return nil, fmt.Errorf("there can only be one encoding tag on the field '%s'", field.Name)
			}
			customTagFound = true

			if !TagLookupKeyFollowsNamingConvention(normalizedLookupKeyForTag) {
				return nil, fmt.Errorf("tag '%s' with lookup key '%s' must adhere to the naming convention", customTag, originalLookupKeyForTag)
			}

			if _, lookupKeyAlreadySeenForTag := tagToLookupKeyToFieldName[customTag][normalizedLookupKeyForTag]; lookupKeyAlreadySeenForTag {
				return nil, fmt.Errorf("tag '%s' with lookup key '%s' is not unique", customTag, originalLookupKeyForTag)
			}
			tagToLookupKeyToFieldName[customTag][normalizedLookupKeyForTag] = field

			if jsonTagValue, jsonTagFound := field.Tag.Lookup(string(JSONTag)); !jsonTagFound || jsonTagValue != "-" {
				return nil, fmt.Errorf("struct field '%s' with tag '%s' must have accompanying tag %s:\"-\"", field.Name, customTag, JSONTag)
			}
		}
	}

	return tagToLookupKeyToFieldName, nil
}
