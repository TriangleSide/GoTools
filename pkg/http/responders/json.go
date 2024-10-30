package responders

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
)

// JSON responds to an HTTP request by encoding the response as JSON.
// An error is returned if there was an error writing the response.
func JSON[RequestParameters any, ResponseBody any](writer http.ResponseWriter, request *http.Request, callback func(*RequestParameters) (*ResponseBody, int, error)) error {
	requestParams, err := parameters.Decode[RequestParameters](request)
	if err != nil {
		return Error(writer, err)
	}

	response, status, err := callback(requestParams)
	if err != nil {
		return Error(writer, err)
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to marshal json response (%w)", err), Error(writer, err))
	}

	writer.Header().Set(headers.ContentLength, strconv.Itoa(len(jsonBytes)))
	writer.Header().Set(headers.ContentType, headers.ContentTypeApplicationJson)
	writer.WriteHeader(status)

	if _, writeErr := io.Copy(writer, bytes.NewBuffer(jsonBytes)); writeErr != nil {
		return fmt.Errorf("failed to write json response (%w)", writeErr)
	}

	return nil
}
