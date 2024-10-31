package responders

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
)

// ErrorResponse is the standard JSON response an API endpoint makes when an error occurs in the endpoint handler.
type ErrorResponse struct {
	Message string `json:"message"`
}

// errJoinUnwrap unwraps errors joined by errors.Join.
type errJoinUnwrap interface {
	Unwrap() []error
}

// findRegistryMatch checks the error type to see if it matches a value in the registry.
func findRegistryMatch(err error) (error, *registeredErrorResponse) {
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
			return workErr, registeredErrorNotCast.(*registeredErrorResponse)
		}
		if matchErr, match := findRegistryMatch(errors.Unwrap(workErr)); match != nil {
			return matchErr, match
		}
	}

	return nil, nil
}

// Error responds to an HTTP requests with an ErrorResponse. It tries to match it to a known error type
// so it can return its corresponding status and message. It defaults to HTTP 500 internal server error.
// An error is returned if there was an error writing the response.
func Error(writer http.ResponseWriter, err error, opts ...Option) {
	cfg := configure(opts...)

	statusCode := http.StatusInternalServerError
	errResponse := ErrorResponse{
		Message: http.StatusText(http.StatusInternalServerError),
	}

	if matchErr, match := findRegistryMatch(err); match != nil {
		statusCode = match.Status
		errResponse.Message = match.MessageCallback(matchErr)
	}

	// The error isn't checked here because Marshal should always succeed on the ErrorResponse struct.
	jsonBytes, _ := json.Marshal(errResponse)

	writer.Header().Set(headers.ContentLength, strconv.Itoa(len(jsonBytes)))
	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.WriteHeader(statusCode)

	if _, writeErr := io.Copy(writer, bytes.NewBuffer(jsonBytes)); writeErr != nil {
		cfg.writeErrorCallback(writeErr)
	}
}
