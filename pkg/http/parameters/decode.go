package parameters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"intelligence/pkg/http/headers"
	"intelligence/pkg/logger"
	reflectutils "intelligence/pkg/utils/reflect"
	"intelligence/pkg/validation"
)

// Decode populates a parameter struct with values from an HTTP request and performs validation on the struct.
func Decode[T any](request *http.Request) (*T, error) {
	logEntry := logger.LogEntry(request.Context())
	logEntry.Tracef("Parsing request parameters.")

	params := new(T)
	if reflect.ValueOf(*params).Kind() != reflect.Struct {
		return nil, fmt.Errorf("the generic must be a struct")
	}

	tagToLookupKeyToFieldName, err := ExtractAndValidateFieldTagLookupKeys[T]()
	if err != nil {
		return nil, fmt.Errorf("tags are not correctly formatted (%s)", err.Error())
	}

	if err := decodeJSONBodyParameters(params, request); err != nil {
		return nil, fmt.Errorf("failed to parse json body parameters (%s)", err.Error())
	}

	if err := decodeQueryParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse query parameters (%s)", err.Error())
	}

	if err := decodeHeaderParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse header parameters (%s)", err.Error())
	}

	if err := decodePathParameters(params, tagToLookupKeyToFieldName, request); err != nil {
		return nil, fmt.Errorf("failed to parse path parameters (%s)", err.Error())
	}

	if err := validation.Struct(params); err != nil {
		return nil, fmt.Errorf("validation failed for request parameters (%s)", err.Error())
	}

	return params, nil
}

// decodeJSONBodyParameters decodes JSON from the request body into the parameter struct.
func decodeJSONBodyParameters(params any, request *http.Request) error {
	if strings.EqualFold(request.Header.Get(headers.ContentType), headers.ContentTypeApplicationJson) {
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		defer func() {
			if err := request.Body.Close(); err != nil {
				logger.LogEntry(request.Context()).WithError(err).Error("Failed to close the request body.")
			}
		}()
		if err := decoder.Decode(&params); err != nil {
			return fmt.Errorf("failed to decode json body (%s)", err.Error())
		}
	}
	return nil
}

// decodeQueryParameters identifies fields tagged with QueryTag and maps corresponding URL query parameters to these fields.
func decodeQueryParameters(params any, tagToLookupKeyToFieldName map[Tag]map[string]reflect.StructField, request *http.Request) error {
	lookupKeyToStructField, tagFound := tagToLookupKeyToFieldName[QueryTag]
	if !tagFound {
		panic("the query tag should be present on the lookup key map")
	}

	for queryParameterName, queryParameterValues := range request.URL.Query() {
		lowerCaseQueryParameterName := strings.ToLower(queryParameterName)
		matchedFieldName, hasMatchedFieldName := lookupKeyToStructField[lowerCaseQueryParameterName]
		if !hasMatchedFieldName {
			continue
		}
		if len(queryParameterValues) != 1 {
			return fmt.Errorf("expecting one value for query parameter %s but found %v", queryParameterName, queryParameterValues)
		}
		if err := reflectutils.AssignToField(params, matchedFieldName.Name, queryParameterValues[0]); err != nil {
			return fmt.Errorf("failed to set value for query parameter %s with values of %v (%s)", queryParameterName, queryParameterValues, err.Error())
		}
	}

	return nil
}

// decodeHeaderParameters identifies fields tagged with HeaderTag and maps corresponding HTTP headers to these fields.
func decodeHeaderParameters(params any, tagToLookupKeyToFieldName map[Tag]map[string]reflect.StructField, request *http.Request) error {
	lookupKeyToStructField, tagFound := tagToLookupKeyToFieldName[HeaderTag]
	if !tagFound {
		panic("the header tag should be present on the lookup key map")
	}

	// For each header in the request, find out if the params struct has a field for it, then attempt to set it if so.
	for headerName, headerValues := range request.Header {
		lowerCaseHeaderName := strings.ToLower(headerName)
		matchedFieldName, hasMatchedFieldName := lookupKeyToStructField[lowerCaseHeaderName]
		if !hasMatchedFieldName {
			continue
		}
		if len(headerValues) != 1 {
			return fmt.Errorf("expecting one value for header parameter %s but found %v", headerName, headerValues)
		}
		if err := reflectutils.AssignToField(params, matchedFieldName.Name, headerValues[0]); err != nil {
			return fmt.Errorf("failed to set value for header parameter %s with values of %v (%s)", headerName, headerValues, err.Error())
		}
	}

	return nil
}

// decodePathParameters identifies fields tagged with PathTag and maps corresponding URL path parameters to these fields.
func decodePathParameters(params any, tagToLookupKeyToFieldName map[Tag]map[string]reflect.StructField, request *http.Request) error {
	lookupKeyToStructField, tagFound := tagToLookupKeyToFieldName[PathTag]
	if !tagFound {
		panic("the path tag should be present on the lookup key map")
	}

	for pathName, field := range lookupKeyToStructField {
		pathValue := request.PathValue(pathName)
		if pathValue == "" {
			continue
		}
		if err := reflectutils.AssignToField(params, field.Name, pathValue); err != nil {
			return err
		}
	}

	return nil
}
