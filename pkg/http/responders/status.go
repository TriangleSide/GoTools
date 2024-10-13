package responders

import (
	"net/http"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
)

// Status responds to an HTTP request with a status but no response body.
func Status[RequestParameters any](writer http.ResponseWriter, request *http.Request, callback func(*RequestParameters) (int, error)) {
	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		Error(writer, request, &errors.BadRequest{Err: err})
		return
	}

	status, err := callback(requestParams)
	if err != nil {
		Error(writer, request, err)
		return
	}

	writer.WriteHeader(status)
}
