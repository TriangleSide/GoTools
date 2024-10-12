package responders

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sync"

	httperrors "github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/logger"
)

// registeredErrorTypeResponse is used by the Error responder to format the response.
type registeredErrorResponse struct {
	Status          int
	MessageCallback func(err any) string
}

var (
	// registeredErrorTypes is a map of reflect.Type to registeredErrorResponse
	registeredErrorResponses = sync.Map{}
)

// MustRegisterErrorResponse allows error types to be registered for the Error responder.
// The registered error type should always be instantiated as a pointer for this to work correctly.
func MustRegisterErrorResponse[T error](status int, callback func(err *T) string) {
	typeOfError := reflect.TypeOf((*T)(nil))
	if typeOfError.Elem().Kind() == reflect.Pointer {
		panic("The generic for registered error types cannot be a pointer.")
	}
	convertedCallback := func(err any) string {
		return callback(err.(*T))
	}
	errorResponse := registeredErrorResponse{
		Status:          status,
		MessageCallback: convertedCallback,
	}
	_, alreadyRegistered := registeredErrorResponses.LoadOrStore(typeOfError, errorResponse)
	if alreadyRegistered {
		panic("The error type has already been registered.")
	}
}

// Error responds to an HTTP requests with an errors.Error. It tries to match it to a known error type
// so it can return its corresponding status and message. It defaults to HTTP 500 internal server error.
func Error(request *http.Request, writer http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	errResponse := httperrors.Error{
		Message: http.StatusText(http.StatusInternalServerError),
	}

	if err != nil {
		errType := reflect.TypeOf(err)
		if registeredErrorNotCast, registeredErrorFound := registeredErrorResponses.Load(errType); registeredErrorFound {
			registeredError := registeredErrorNotCast.(registeredErrorResponse)
			statusCode = registeredError.Status
			errResponse.Message = registeredError.MessageCallback(err)
		} else {
			var badRequestError *httperrors.BadRequest
			switch {
			case errors.As(err, &badRequestError):
				statusCode = http.StatusBadRequest
				errResponse.Message = badRequestError.Error()
			}
		}
	}

	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.WriteHeader(statusCode)

	if err := json.NewEncoder(writer).Encode(errResponse); err != nil {
		logger.Errorf(request.Context(), "Error encoding error response (%s).", err)
	}
}
