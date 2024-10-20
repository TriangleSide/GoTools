package responders

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sync"

	httperrors "github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/logger"
)

// init registers error messages for the responder.
func init() {
	MustRegisterErrorResponse[httperrors.BadRequest](http.StatusBadRequest, func(err *httperrors.BadRequest) string {
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

// Error responds to an HTTP requests with an errors.Error. It tries to match it to a known error type
// so it can return its corresponding status and message. It defaults to HTTP 500 internal server error.
func Error(writer http.ResponseWriter, request *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	errResponse := httperrors.Error{
		Message: http.StatusText(http.StatusInternalServerError),
	}

	if err != nil {
		errType := reflect.TypeOf(err)
		if registeredErrorNotCast, registeredErrorFound := registeredErrorResponses.Load(errType); registeredErrorFound {
			registeredError := registeredErrorNotCast.(*registeredErrorResponse)
			statusCode = registeredError.Status
			errResponse.Message = registeredError.MessageCallback(err)
		}
	}

	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.WriteHeader(statusCode)
	if encodingError := json.NewEncoder(writer).Encode(errResponse); encodingError != nil {
		logger.Errorf(request.Context(), "Error encoding error response (%s).", encodingError.Error())
	}
}
