package responders

import (
	"net/http"

	"github.com/TriangleSide/go-toolkit/pkg/http/parameters"
)

// Status responds to an HTTP request with a status but no response body.
func Status[RequestParameters any](
	writer http.ResponseWriter,
	request *http.Request,
	callback func(*RequestParameters) (int, error),
	opts ...Option,
) {
	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		Error(writer, err, opts...)
		return
	}

	status, err := callback(requestParams)
	if err != nil {
		Error(writer, err, opts...)
		return
	}

	writer.WriteHeader(status)
}
