package parameters

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

// Decode populates a parameter struct with values from an HTTP request.
// After decoding, it validates the struct and closes the request body.
func Decode[T any](request *http.Request) (*T, error) {
	params, err := decodeAndValidate[T](request)
	if request.Body != nil {
		if closeErr := request.Body.Close(); closeErr != nil {
			return nil, errors.Join(err, fmt.Errorf("failed to close the request body: %w", closeErr))
		}
	}
	return params, err
}

// decodeAndValidate extracts parameter values from the request and validates the resulting struct.
func decodeAndValidate[T any](request *http.Request) (*T, error) {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		panic(errors.New("generic type must be a struct"))
	}

	params := new(T)
	tagToLookupKeyToFieldName, err := ExtractAndValidateFieldTagLookupKeys[T]()
	if err != nil {
		panic(fmt.Errorf("tags are not correctly formatted: %w", err))
	}

	if err := decodeJSONBodyParameters(params, request); err != nil {
		return nil, fmt.Errorf("failed to parse json body parameters: %w", err)
	}

	if err := decodeQueryParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse query parameters: %w", err)
	}

	if err := decodeHeaderParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse header parameters: %w", err)
	}

	if err := decodePathParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse path parameters: %w", err)
	}

	if err := validation.Struct(params); err != nil {
		return nil, fmt.Errorf("validation failed for request parameters: %w", err)
	}

	return params, nil
}

// decodeJSONBodyParameters decodes JSON from the request body into the parameter struct.
func decodeJSONBodyParameters[T any](params *T, request *http.Request) error {
	if strings.EqualFold(request.Header.Get(headers.ContentType), headers.ContentTypeApplicationJSON) {
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(params); err != nil {
			return fmt.Errorf("failed to decode json body: %w", err)
		}
	}
	return nil
}

// decodeQueryParameters populates struct fields tagged with QueryTag using URL query parameter values.
// For example, a field tagged with `urlQuery:"name"` is populated from the query parameter "name" in "?name=value".
func decodeQueryParameters[T any](
	params *T,
	tagToLookupKeyToFieldName map[Tag]LookupKeyToFieldName,
	request *http.Request,
) error {
	lookupKeyToFieldName := tagToLookupKeyToFieldName[QueryTag]
	normalizer := tagToLookupKeyNormalizer[QueryTag]

	for queryParameterName, queryParameterValues := range request.URL.Query() {
		normalizedQueryParameterName := normalizer(queryParameterName)
		matchedFieldName, hasMatchedFieldName := lookupKeyToFieldName[normalizedQueryParameterName]
		if !hasMatchedFieldName {
			continue
		}
		if len(queryParameterValues) != 1 {
			return fmt.Errorf(
				"expecting one value for query parameter %s but found %v",
				queryParameterName, queryParameterValues)
		}
		if err := structs.AssignToField(params, matchedFieldName, queryParameterValues[0]); err != nil {
			return fmt.Errorf(
				"failed to set value for query parameter %s with values of %v: %w",
				queryParameterName, queryParameterValues, err)
		}
	}

	return nil
}

// decodeHeaderParameters populates struct fields tagged with HeaderTag using HTTP header values.
// For example, a field tagged with `httpHeader:"X-Request-ID"` is populated from the "X-Request-ID" header.
func decodeHeaderParameters[T any](
	params *T,
	tagToLookupKeyToFieldName map[Tag]LookupKeyToFieldName,
	request *http.Request,
) error {
	lookupKeyToFieldName := tagToLookupKeyToFieldName[HeaderTag]
	normalizer := tagToLookupKeyNormalizer[HeaderTag]

	for headerName, headerValues := range request.Header {
		normalizedHeaderName := normalizer(headerName)
		matchedFieldName, hasMatchedFieldName := lookupKeyToFieldName[normalizedHeaderName]
		if !hasMatchedFieldName {
			continue
		}
		if len(headerValues) != 1 {
			return fmt.Errorf("expecting one value for header parameter %s but found %v", headerName, headerValues)
		}
		if err := structs.AssignToField(params, matchedFieldName, headerValues[0]); err != nil {
			return fmt.Errorf(
				"failed to set value for header parameter %s with values of %v: %w",
				headerName, headerValues, err)
		}
	}

	return nil
}

// decodePathParameters populates struct fields tagged with PathTag using URL path parameter values.
// For example, a field tagged with `urlPath:"id"` is populated from the path parameter "{id}" in "/users/{id}".
func decodePathParameters[T any](
	params *T,
	tagToLookupKeyToFieldName map[Tag]LookupKeyToFieldName,
	request *http.Request,
) error {
	lookupKeyToFieldName := tagToLookupKeyToFieldName[PathTag]
	normalizer := tagToLookupKeyNormalizer[PathTag]

	for pathName, field := range lookupKeyToFieldName {
		normalizedPathName := normalizer(pathName)
		pathValue := request.PathValue(normalizedPathName)
		if pathValue == "" {
			continue
		}
		if err := structs.AssignToField(params, field, pathValue); err != nil {
			return fmt.Errorf("failed to set value for path parameter %s with values of %v: %w", pathName, pathValue, err)
		}
	}

	return nil
}
