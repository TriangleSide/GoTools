package responders

import (
	"net/http"
	"reflect"
	"sync"

	"github.com/TriangleSide/GoBase/pkg/validation"
)

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

// init registers standard error messages for the responder.
func init() {
	MustRegisterErrorResponse[validation.Violations](http.StatusBadRequest, func(err *validation.Violations) string {
		return err.Error()
	})
}
