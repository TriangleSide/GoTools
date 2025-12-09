package responders

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/TriangleSide/GoTools/pkg/http/headers"
)

// errJoinUnwrap unwraps errors joined by errors.Join.
type errJoinUnwrap interface {
	Unwrap() []error
}

// findRegistryMatch checks the error type to see if it matches a value in the registry.
func findRegistryMatch(err error) (*registeredErrorResponse, error) {
	if err == nil {
		return nil, nil
	}

	var allErrs []error
	if uw, ok := err.(errJoinUnwrap); ok {
		allErrs = uw.Unwrap()
	} else {
		allErrs = []error{err}
	}

	for _, workErr := range allErrs {
		errType := reflect.TypeOf(workErr)
		if registeredErrorNotCast, registeredErrorFound := registeredErrorResponses.Load(errType); registeredErrorFound {
			return registeredErrorNotCast.(*registeredErrorResponse), workErr
		}
		if match, matchErr := findRegistryMatch(errors.Unwrap(workErr)); match != nil {
			return match, matchErr
		}
	}

	return nil, nil
}

// Error responds to an HTTP requests with an ErrorResponse. It tries to match it to a known error type
// so it can return its corresponding status and message. It defaults to HTTP 500 internal server error.
func Error(writer http.ResponseWriter, err error, opts ...Option) {
	cfg := configure(opts...)

	var statusCode int
	var errResponse any
	if match, matchErr := findRegistryMatch(err); match != nil {
		statusCode = match.Status
		errResponse = match.Callback(matchErr)
	} else {
		statusCode = http.StatusInternalServerError
		errResponse = StandardErrorResponse{
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	jsonBytes, err := cfg.jsonMarshal(errResponse)
	if err != nil {
		cfg.errorCallback(err)
		statusCode = http.StatusInternalServerError
		errResponse = StandardErrorResponse{
			Message: http.StatusText(http.StatusInternalServerError),
		}
		jsonBytes, err = cfg.jsonMarshal(errResponse)
		if err != nil {
			cfg.errorCallback(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	writer.Header().Set(headers.ContentLength, strconv.Itoa(len(jsonBytes)))
	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJSON)
	writer.WriteHeader(statusCode)

	if _, writeErr := writer.Write(jsonBytes); writeErr != nil {
		cfg.errorCallback(writeErr)
		return
	}
}
