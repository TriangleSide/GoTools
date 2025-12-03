package responders

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/parameters"
)

// JSON responds to an HTTP request by encoding the response as JSON.
// An error is returned if there was an error writing the response.
func JSON[RequestParameters any, ResponseBody any](writer http.ResponseWriter, request *http.Request, callback func(*RequestParameters) (*ResponseBody, int, error), opts ...Option) {
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
	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.WriteHeader(status)

	if _, writeErr := io.Copy(writer, bytes.NewBuffer(jsonBytes)); writeErr != nil {
		cfg.errorCallback(writeErr)
		return
	}
}
