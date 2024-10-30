package responders

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
)

// JSONStream responds to an HTTP request by streaming responses as JSON objects.
// The producer is responsible for closing the response channel.
// An error is returned if there was an error writing the response.
func JSONStream[RequestParameters any, ResponseBody any](writer http.ResponseWriter, request *http.Request, callback func(*RequestParameters) (<-chan *ResponseBody, int, error)) error {
	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		return Error(writer, err)
	}

	responseChan, status, err := callback(requestParams)
	if err != nil {
		return Error(writer, err)
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
			return nil
		case response, isOpen := <-responseChan:
			if !isOpen {
				return nil
			}
			if encoderError := jsonEncoder.Encode(response); encoderError != nil {
				return fmt.Errorf("failed to encode response (%w)", encoderError)
			}
			if isFlusher {
				flusher.Flush()
			}
		}
	}
}
