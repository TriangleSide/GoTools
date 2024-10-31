package responders

import (
	"net/http"
	"reflect"
	"sync"

	"github.com/TriangleSide/GoBase/pkg/validation"
)

// registeredErrorTypeResponse is used by the Error responder to format the response.
type registeredErrorResponse struct {
	Status   int
	Callback func(err any) any
}

var (
	// registeredErrorResponses is a map of reflect.Type to *registeredErrorResponse.
	registeredErrorResponses = sync.Map{}
)

// MustRegisterErrorResponse allows error types to be registered for the Error responder.
// The registered error type should always be instantiated as a pointer for this to work correctly.
func MustRegisterErrorResponse[T any, R any](status int, callback func(err *T) *R) {
	typeOfError := reflect.TypeFor[T]()
	if typeOfError.Kind() != reflect.Struct {
		panic("The generic for registered error responses must be a struct.")
	}

	ptrToTypeOfError := reflect.PointerTo(typeOfError)
	if !ptrToTypeOfError.Implements(reflect.TypeFor[error]()) {
		panic("The generic for registered error types must have an error interface.")
	}

	responseType := reflect.TypeFor[R]()
	if responseType.Kind() != reflect.Struct {
		panic("The response type must be a struct.")
	}

	errorResponse := &registeredErrorResponse{
		Status: status,
		Callback: func(err any) any {
			return callback(err.(*T))
		},
	}
	_, alreadyRegistered := registeredErrorResponses.LoadOrStore(ptrToTypeOfError, errorResponse)
	if alreadyRegistered {
		panic("The error type has already been registered.")
	}
}

// StandardErrorResponse is the standard JSON response an API endpoint makes when an unknown error occurs in the endpoint handler.
type StandardErrorResponse struct {
	Message string `json:"message"`
}

// init registers standard error messages for the responder.
func init() {
	MustRegisterErrorResponse[validation.Violations, StandardErrorResponse](http.StatusBadRequest, func(err *validation.Violations) *StandardErrorResponse {
		return &StandardErrorResponse{
			Message: err.Error(),
		}
	})
}
