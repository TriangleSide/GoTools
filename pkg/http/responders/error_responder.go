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

// findRegistryMatchAndPerformCallback checks the error type to see if it matches a value in the registry.
// If it does, it returns the status code, the result of the callback, and true.
// If not, it returns zero values and false.
func findRegistryMatchAndPerformCallback(err error) (int, any, bool) {
	if err == nil {
		return 0, nil, false
	}

	errType := normalizeErrorTypeForRegistry(reflect.TypeOf(err))
	if registeredErrorNotCast, found := registeredErrorResponses.Load(errType); found {
		registeredErr := registeredErrorNotCast.(*registeredErrorResponse)
		return registeredErr.Status, registeredErr.Callback(err), true
	}

	if uw, ok := err.(errJoinUnwrap); ok {
		for _, joinErr := range uw.Unwrap() {
			if status, result, found := findRegistryMatchAndPerformCallback(joinErr); found {
				return status, result, true
			}
		}
	}

	if status, result, found := findRegistryMatchAndPerformCallback(errors.Unwrap(err)); found {
		return status, result, true
	}

	return 0, nil, false
}

// Error responds to an HTTP requests with an ErrorResponse. It tries to match it to a known error type
// so it can return its corresponding status and message. It defaults to HTTP 500 internal server error.
func Error(writer http.ResponseWriter, err error, opts ...Option) {
	cfg := configure(opts...)

	var statusCode int
	var errResponse any
	if callbackStatus, callbackResult, matched := findRegistryMatchAndPerformCallback(err); matched {
		statusCode = callbackStatus
		errResponse = callbackResult
	} else {
		statusCode = http.StatusInternalServerError
		errResponse = StandardErrorResponse{
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	jsonBytes, marshalErr := cfg.jsonMarshal(errResponse)
	if marshalErr != nil {
		cfg.errorCallback(marshalErr)
		statusCode = http.StatusInternalServerError
		errResponse = StandardErrorResponse{
			Message: http.StatusText(http.StatusInternalServerError),
		}
		jsonBytes, marshalErr = cfg.jsonMarshal(errResponse)
		if marshalErr != nil {
			cfg.errorCallback(marshalErr)
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
