package responders

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"sync"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

// init registers error messages for the responder.
func init() {
	MustRegisterErrorResponse[validation.Violations](http.StatusBadRequest, func(err *validation.Violations) string {
		return err.Error()
	})
}

// registeredErrorTypeResponse is used by the Error responder to format the response.
type registeredErrorResponse struct {
	Status          int
	MessageCallback func(err any) string
}

var (
	// registeredErrorResponses is a map of reflect.Type to *registeredErrorResponse.
	registeredErrorResponses = sync.Map{}
)

// MustRegisterErrorResponse allows error types to be registered for the Error responder.
// The registered error type should always be instantiated as a pointer for this to work correctly.
func MustRegisterErrorResponse[T any](status int, callback func(err *T) string) {
	typeOfError := reflect.TypeFor[T]()
	if typeOfError.Kind() != reflect.Struct {
		panic("The generic for registered error responses must be a struct.")
	}
	typeOfError = reflect.PointerTo(typeOfError)
	if !typeOfError.Implements(reflect.TypeFor[error]()) {
		panic("The generic for registered error types must have an error interface.")
	}
	errorResponse := &registeredErrorResponse{
		Status: status,
		MessageCallback: func(err any) string {
			return callback(err.(*T))
		},
	}
	_, alreadyRegistered := registeredErrorResponses.LoadOrStore(typeOfError, errorResponse)
	if alreadyRegistered {
		panic("The error type has already been registered.")
	}
}

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
func Error(writer http.ResponseWriter, err error) error {
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
		return fmt.Errorf("failed to write error response (%w)", writeErr)
	}

	return nil
}
