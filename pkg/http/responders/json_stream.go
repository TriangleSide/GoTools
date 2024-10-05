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

// JSONStreamOption is used to set values on the stream configuration.
type JSONStreamOption func(config *jsonStreamConfig)

// WithDeferredConsumerTimerDuration configures how long to wait before printing
// an error log on the deferred consumer.
func WithDeferredConsumerTimerDuration(duration time.Duration) JSONStreamOption {
	return func(config *jsonStreamConfig) {
		config.deferredConsumerTimerDuration = duration
	}
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
func JSONStream[RequestParameters any, ResponseBody any](writer http.ResponseWriter, request *http.Request, callback func(requestParameters *RequestParameters, cancelChan <-chan struct{}) (responseStream <-chan *ResponseBody, status int, err error), options ...JSONStreamOption) {
	cfg := &jsonStreamConfig{
		deferredConsumerTimerDuration: time.Minute,
	}
	for _, option := range options {
		option(cfg)
	}

	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		Error(request, writer, &errors.BadRequest{Err: err})
		return
	}

	cancelChan := make(chan struct{})
	defer close(cancelChan)

	responseChan, status, err := callback(requestParams, cancelChan)
	if err != nil {
		Error(request, writer, err)
		return
	}

	defer func() {
		go func() {
			timer := time.After(cfg.deferredConsumerTimerDuration)
			for {
				select {
				case <-timer:
					logger.Errorf(request.Context(), "Potential leak detected: JSON stream producer did not close its channel after %s.", cfg.deferredConsumerTimerDuration.String())
				case _, isResponseChannelOpen := <-responseChan:
					if !isResponseChannelOpen {
						return
					}
				}
			}
		}()
	}()

	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.Header().Set(headers.TransferEncoding, headers.TransferEncodingChunked)
	writer.WriteHeader(status)

	ctx := request.Context()
	jsonEncoder := json.NewEncoder(writer)
	for {
		select {
		case <-ctx.Done():
			logger.Errorf(ctx, "Request cancelled (%s).", ctx.Err())
			return
		case response, isResponseChannelOpen := <-responseChan:
			if !isResponseChannelOpen {
				return
			}
			if err := jsonEncoder.Encode(response); err != nil {
				logger.Errorf(ctx, "Failed to encode response (%s).", err)
				return
			}
			if flusher, ok := writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}
