package responders

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/parameters"
)

// JSONStream responds to an HTTP request by streaming responses as JSON objects.
// The producer is responsible for closing the response channel.
func JSONStream[RequestParameters any, ResponseBody any](
	writer http.ResponseWriter,
	request *http.Request,
	callback func(*RequestParameters) (<-chan *ResponseBody, int, error),
	opts ...Option,
) {
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
	if responseChan == nil {
		Error(writer, errors.New("response channel cannot be nil"), opts...)
		return
	}

	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJSON)
	writer.Header().Set(headers.TransferEncoding, headers.TransferEncodingChunked)
	writer.WriteHeader(status)

	streamResponses(request.Context(), writer, responseChan, cfg)
}

// streamResponses writes responses from the channel to the writer until the channel closes or context is cancelled.
func streamResponses[ResponseBody any](
	ctx context.Context,
	writer http.ResponseWriter,
	responseChan <-chan *ResponseBody,
	cfg *config,
) {
	flusher, isFlusher := writer.(http.Flusher)
	jsonEncoder := json.NewEncoder(writer)

	for {
		select { // This additional select is because of the non-deterministic nature of select below.
		case <-ctx.Done():
			return
		default:
		}
		select { // Select is non-deterministic if multiple cases are ready.
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
