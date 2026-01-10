package responders

import (
	"net/http"
	"strconv"

	"github.com/TriangleSide/go-toolkit/pkg/http/headers"
	"github.com/TriangleSide/go-toolkit/pkg/http/parameters"
)

// JSON responds to an HTTP request by encoding the response as JSON.
func JSON[RequestParameters any, ResponseBody any](
	writer http.ResponseWriter,
	request *http.Request,
	callback func(*RequestParameters) (*ResponseBody, int, error),
	opts ...Option,
) {
	cfg := configure(opts...)

	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		Error(writer, err, opts...)
		return
	}

	response, status, err := callback(requestParams)
	if err != nil {
		Error(writer, err, opts...)
		return
	}

	jsonBytes, err := cfg.jsonMarshal(response)
	if err != nil {
		Error(writer, err, opts...)
		return
	}

	writer.Header().Set(headers.ContentLength, strconv.Itoa(len(jsonBytes)))
	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJSON)
	writer.WriteHeader(status)

	if _, writeErr := writer.Write(jsonBytes); writeErr != nil {
		cfg.errorCallback(writeErr)
		return
	}
}
