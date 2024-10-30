package responders

import (
	"net/http"

	"github.com/TriangleSide/GoBase/pkg/http/parameters"
)

// Status responds to an HTTP request with a status but no response body.
// An error is returned if there was an error writing the response.
func Status[RequestParameters any](writer http.ResponseWriter, request *http.Request, callback func(*RequestParameters) (int, error)) error {
	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		return Error(writer, err)
	}

	status, err := callback(requestParams)
	if err != nil {
		return Error(writer, err)
	}

	writer.WriteHeader(status)
	return nil
}
