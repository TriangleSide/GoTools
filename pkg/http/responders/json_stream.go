package responders

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
	"github.com/TriangleSide/GoBase/pkg/logger"
)

// jsonStreamConfig is used to configure the JSON stream utility.
type jsonStreamConfig struct {
	deferredConsumerTimerDuration time.Duration
}

// JSONStream responds to an HTTP request by streaming responses as JSON objects.
//
// When this method exits, it launches a go routine to continue consuming the responses
// to ensure the producer closes the channel appropriately. This is done in the
// case that the producer is blocked writing on the channel.
//
// The producer routine must check for the cancel channel and stop producing.
// Example producer go routine:
//
//	for {
//	  select {
//	  case <-cancel:
//	    return
//	  default:
//	    // Producer work here.
//	  }
//	}
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
		case response, isResponseChannelOpen := <-responseChan:
			if !isResponseChannelOpen {
				return
			}
			if err := jsonEncoder.Encode(response); err != nil {
				logger.Errorf(ctx, "Failed to encode response (%s).", err.Error())
				return
			}
			if flusher, ok := writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}
