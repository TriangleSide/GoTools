package responders

import (
	"encoding/json"
	"net/http"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
)

// JSONStream responds to an HTTP request by streaming responses as JSON objects.
// The producer is responsible for closing the response channel.
// An error is returned if there was an error writing the response.
func JSONStream[RequestParameters any, ResponseBody any](writer http.ResponseWriter, request *http.Request, callback func(*RequestParameters) (<-chan *ResponseBody, int, error), opts ...Option) {
	cfg := configure(opts...)

	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		Error(writer, err, opts...)
		return
	}

	responseChan, status, err := callback(requestParams)
	if err != nil {
		Error(writer, err, opts...)
		return
	}

	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.Header().Set(headers.TransferEncoding, headers.TransferEncodingChunked)
	writer.WriteHeader(status)

	ctx := request.Context()
	flusher, isFlusher := writer.(http.Flusher)
	jsonEncoder := json.NewEncoder(writer)

	for {
		select {
		case <-ctx.Done():
			return
		case response, isOpen := <-responseChan:
			if !isOpen {
				return
			}
			if encoderError := jsonEncoder.Encode(response); encoderError != nil {
				cfg.errorCallback(encoderError)
				return
			}
			if isFlusher {
				flusher.Flush()
			}
		}
	}
}
