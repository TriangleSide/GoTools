package responders

import (
	"encoding/json"
	"net/http"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
	"github.com/TriangleSide/GoBase/pkg/logger"
)

// JSONStream responds to an HTTP request by streaming responses as JSON objects.
// The producer is responsible for closing the response channel.
func JSONStream[RequestParameters any, ResponseBody any](writer http.ResponseWriter, request *http.Request, callback func(requestParameters *RequestParameters) (responseStream <-chan *ResponseBody, status int, err error)) {
	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		Error(writer, request, &errors.BadRequest{Err: err})
		return
	}

	responseChan, status, err := callback(requestParams)
	if err != nil {
		Error(writer, request, err)
		return
	}

	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.Header().Set(headers.TransferEncoding, headers.TransferEncodingChunked)
	writer.WriteHeader(status)

	ctx := request.Context()
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
				logger.Errorf(ctx, "Failed to encode response (%s).", encoderError.Error())
				return
			}
			if flusher, ok := writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}
