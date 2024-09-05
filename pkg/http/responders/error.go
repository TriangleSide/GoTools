package responders

import (
	"encoding/json"
	goerrors "errors"
	"net/http"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
)

// registeredErrorTypeResponse is used by the Error responder to format the response.
type registeredErrorResponse struct {
	Status          int
	MessageCallback func(err any) string
}

var (
	// registeredErrorTypes holds error types and how to format their responses.
	registeredErrorResponses = make(map[reflect.Type]registeredErrorResponse)
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
	if _, found := registeredErrorResponses[typeOfError]; found {
		panic("The error type has already been registered.")
	}
	registeredErrorResponses[typeOfError] = errorResponse
}

// Error responds to an HTTP requests with an errors.Error. It tries to match it to a known error type
// so it can return its corresponding status and message. It defaults to HTTP 500 internal server error.
func Error(writer http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	errResponse := errors.Error{
		Message: http.StatusText(http.StatusInternalServerError),
	}

	if err != nil {
		errType := reflect.TypeOf(err)
		if registeredError, registeredErrorFound := registeredErrorResponses[errType]; registeredErrorFound {
			statusCode = registeredError.Status
			errResponse.Message = registeredError.MessageCallback(err)
		} else {
			var badRequestError *errors.BadRequest
			switch {
			case goerrors.As(err, &badRequestError):
				statusCode = http.StatusBadRequest
				errResponse.Message = badRequestError.Error()
			}
		}
	}

	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.WriteHeader(statusCode)

	if err := json.NewEncoder(writer).Encode(errResponse); err != nil {
		logrus.WithError(err).Error("Error encoding error response.")
	}
}
