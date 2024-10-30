package parameters

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/TriangleSide/GoBase/pkg/datastructures/readonly"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/utils/assign"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

// Decode populates a parameter struct with values from an HTTP request and performs validation on the struct.
func Decode[T any](request *http.Request) (returnParams *T, returnErr error) {
	defer func() {
		if request.Body != nil {
			if err := request.Body.Close(); err != nil {
				returnErr = errors.Join(returnErr, fmt.Errorf("failed to close the response body (%w)", err))
				returnParams = nil
			}
		}
	}()

	params := new(T)
	if reflect.ValueOf(*params).Kind() != reflect.Struct {
		panic("the generic must be a struct")
	}

	tagToLookupKeyToFieldName, err := ExtractAndValidateFieldTagLookupKeys[T]()
	if err != nil {
		panic(fmt.Sprintf("tags are not correctly formatted (%s)", err.Error()))
	}

	if err := decodeJSONBodyParameters(params, request); err != nil {
		return nil, fmt.Errorf("failed to parse json body parameters (%w)", err)
	}

	if err := decodeQueryParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse query parameters (%w)", err)
	}

	if err := decodeHeaderParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse header parameters (%w)", err)
	}

	if err := decodePathParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse path parameters (%w)", err)
	}

	if err := validation.Struct(params); err != nil {
		return nil, fmt.Errorf("validation failed for request parameters (%w)", err)
	}

	return params, nil
}

// decodeJSONBodyParameters decodes JSON from the request body into the parameter struct.
func decodeJSONBodyParameters[T any](params *T, request *http.Request) error {
	if strings.EqualFold(request.Header.Get(headers.ContentType), headers.ContentTypeApplicationJson) {
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&params); err != nil {
			return fmt.Errorf("failed to decode json body (%w)", err)
		}
	}
	return nil
}

// decodeQueryParameters identifies fields tagged with QueryTag and maps corresponding URL query parameters to these fields.
func decodeQueryParameters[T any](params *T, tagToLookupKeyToFieldName *readonly.Map[Tag, LookupKeyToFieldName], request *http.Request) error {
	lookupKeyToFieldName := tagToLookupKeyToFieldName.Get(QueryTag)
	normalizer := tagToLookupKeyNormalizer[QueryTag]

	for queryParameterName, queryParameterValues := range request.URL.Query() {
		normalizedQueryParameterName := normalizer(queryParameterName)
		matchedFieldName, hasMatchedFieldName := lookupKeyToFieldName[normalizedQueryParameterName]
		if !hasMatchedFieldName {
			continue
		}
		if len(queryParameterValues) != 1 {
			return fmt.Errorf("expecting one value for query parameter %s but found %v", queryParameterName, queryParameterValues)
		}
		if err := assign.StructField(params, matchedFieldName, queryParameterValues[0]); err != nil {
			return fmt.Errorf("failed to set value for query parameter %s with values of %v (%w)", queryParameterName, queryParameterValues, err)
		}
	}

	return nil
}

// decodeHeaderParameters identifies fields tagged with HeaderTag and maps corresponding HTTP headers to these fields.
func decodeHeaderParameters[T any](params *T, tagToLookupKeyToFieldName *readonly.Map[Tag, LookupKeyToFieldName], request *http.Request) error {
	lookupKeyToFieldName := tagToLookupKeyToFieldName.Get(HeaderTag)
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
		if err := assign.StructField(params, matchedFieldName, headerValues[0]); err != nil {
			return fmt.Errorf("failed to set value for header parameter %s with values of %v (%w)", headerName, headerValues, err)
		}
	}

	return nil
}

// decodePathParameters identifies fields tagged with PathTag and maps corresponding URL path parameters to these fields.
func decodePathParameters[T any](params *T, tagToLookupKeyToFieldName *readonly.Map[Tag, LookupKeyToFieldName], request *http.Request) error {
	lookupKeyToFieldName := tagToLookupKeyToFieldName.Get(PathTag)
	normalizer := tagToLookupKeyNormalizer[PathTag]

	for pathName, field := range lookupKeyToFieldName {
		normalizedPathName := normalizer(pathName)
		pathValue := request.PathValue(normalizedPathName)
		if pathValue == "" {
			continue
		}
		if err := assign.StructField(params, field, pathValue); err != nil {
			return fmt.Errorf("failed to set value for path parameter %s with values of %v (%w)", pathName, pathValue, err)
		}
	}

	return nil
}
