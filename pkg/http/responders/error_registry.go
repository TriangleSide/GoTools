package responders

import (
	"net/http"
	"reflect"
	"sync"

	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

// StandardErrorResponse is the standard JSON response an API endpoint makes
// when an error occurs in the endpoint handler.
type StandardErrorResponse struct {
	Message string `json:"message"`
}

// registeredErrorResponse is used by the Error responder to format the response.
type registeredErrorResponse struct {
	// Status is the HTTP status code to return.
	Status int
	// Callback is the function that formats the error response.
	// The response will be marshaled to JSON before being sent.
	Callback func(err any) *StandardErrorResponse
}

var (
	// registeredErrorResponses is a map of reflect.Type to *registeredErrorResponse.
	registeredErrorResponses = sync.Map{}
)

// MustRegisterErrorResponse allows error types to be registered for the Error responder.
// The registered error type should always be instantiated as a pointer for this to work correctly.
func MustRegisterErrorResponse[T any](status int, callback func(err *T) *StandardErrorResponse) {
	errorType := reflect.TypeFor[T]()
	if errorType.Kind() == reflect.Ptr {
		panic("The generic for registered error types cannot be a pointer.")
	}

	errorType = normalizeErrorTypeForRegistry(errorType)
	if !errorType.Implements(reflect.TypeFor[error]()) {
		panic("The generic for registered error types must implement the error interface.")
	}

	errorResponse := &registeredErrorResponse{
		Status: status,
		Callback: func(err any) *StandardErrorResponse {
			return callback(err.(*T))
		},
	}
	_, alreadyRegistered := registeredErrorResponses.LoadOrStore(errorType, errorResponse)
	if alreadyRegistered {
		panic("The error type has already been registered.")
	}
}

// normalizeErrorTypeForRegistry ensures that the error type used for registry lookups is always the same.
func normalizeErrorTypeForRegistry(errType reflect.Type) reflect.Type {
	return reflect.PointerTo(reflection.DereferenceType(errType))
}

// init registers standard error messages for the responder.
func init() {
	MustRegisterErrorResponse[validation.Violations](
		http.StatusBadRequest,
		func(err *validation.Violations) *StandardErrorResponse {
			return &StandardErrorResponse{
				Message: err.Error(),
			}
		})
}
