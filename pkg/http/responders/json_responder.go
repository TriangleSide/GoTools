package responders

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
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

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		Error(writer, err, opts...)
		return
	}

	writer.Header().Set(headers.ContentLength, strconv.Itoa(len(jsonBytes)))
	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.WriteHeader(status)

	if _, writeErr := io.Copy(writer, bytes.NewBuffer(jsonBytes)); writeErr != nil {
		cfg.writeErrorCallback(writeErr)
	}
}
