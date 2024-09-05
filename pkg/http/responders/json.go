package responders

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
)

// JSON responds to an HTTP request by encoding the response as JSON.
func JSON[RequestParameters any, ResponseBody any](writer http.ResponseWriter, request *http.Request, callback func(*RequestParameters) (*ResponseBody, int, error)) {
	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		Error(writer, &errors.BadRequest{Err: err})
		return
	}

	response, status, err := callback(requestParams)
	if err != nil {
		Error(writer, err)
		return
	}

	writer.Header().Add(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.WriteHeader(status)

	if err := json.NewEncoder(writer).Encode(response); err != nil {
		logrus.WithError(err).Error("Failed to encode response.")
		return
	}
}
