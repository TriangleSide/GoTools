package responders_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/responders"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func newJSONResponderTestServer[TRequest, TResponse any](t *testing.T, handler func(*TRequest) (*TResponse, int, error)) (string, func(), func() error) {
	t.Helper()

	var writeErr error
	writeErrorCallback := func(err error) {
		writeErr = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.JSON[TRequest, TResponse](w, r, handler, responders.WithErrorCallback(writeErrorCallback))
	}))

	return server.URL, server.Close, func() error { return writeErr }
}

func postJSON(t *testing.T, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		url,
		strings.NewReader(body),
	)
	assert.NoError(t, err)
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)
	client := &http.Client{}
	response, err := client.Do(req)
	assert.NoError(t, err)
	return response
}

type jsonRequestParams struct {
	ID int `json:"id" validate:"gt=0"`
}

type jsonResponseBody struct {
	Message string `json:"message"`
}

type jsonUnmarshalableResponse struct {
	ChanField chan int `json:"chan_field"`
}

func jsonTestHandler(params *jsonRequestParams) (*jsonResponseBody, int, error) {
	if params.ID == 123 {
		return &jsonResponseBody{Message: "processed"}, http.StatusOK, nil
	}
	return nil, 0, &testError{}
}

func TestJSON_ValidRequest_RespondsWithJSONAndCorrectStatusCode(t *testing.T) {
	t.Parallel()

	serverURL, cleanup, writeErr := newJSONResponderTestServer[jsonRequestParams, jsonResponseBody](t, jsonTestHandler)
	defer cleanup()

	response := postJSON(t, serverURL, `{"id":123}`)
	defer func() { assert.NoError(t, response.Body.Close()) }()

	assert.Equals(t, response.StatusCode, http.StatusOK)
	assert.Equals(t, response.Header.Get(headers.ContentType), headers.ContentTypeApplicationJSON)
	assert.NoError(t, writeErr())

	body := &jsonResponseBody{}
	assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
	assert.Equals(t, body.Message, "processed")
}

func TestJSON_ParameterDecoderFails_RespondsWithErrorJSONAndBadRequestStatus(t *testing.T) {
	t.Parallel()

	serverURL, cleanup, writeErr := newJSONResponderTestServer[jsonRequestParams, jsonResponseBody](t, jsonTestHandler)
	defer cleanup()

	response := postJSON(t, serverURL, `{"id":-1}`)
	defer func() { assert.NoError(t, response.Body.Close()) }()

	assert.Equals(t, response.StatusCode, http.StatusBadRequest)
	assert.NoError(t, writeErr())

	body := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
	assert.Contains(t, body.Message, "validation failed on field 'ID'")
}

func TestJSON_CallbackReturnsError_RespondsWithErrorJSONAndBadRequestStatus(t *testing.T) {
	t.Parallel()

	serverURL, cleanup, writeErr := newJSONResponderTestServer[jsonRequestParams, jsonResponseBody](t, jsonTestHandler)
	defer cleanup()

	response := postJSON(t, serverURL, `{"id":456}`)
	defer func() { assert.NoError(t, response.Body.Close()) }()

	assert.Equals(t, response.StatusCode, http.StatusBadRequest)
	assert.NoError(t, writeErr())

	body := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
	assert.Equals(t, body.Message, "test error")
}

func TestJSON_UnencodableResponse_ReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	responderFunc := func(*jsonRequestParams) (*jsonUnmarshalableResponse, int, error) {
		return &jsonUnmarshalableResponse{}, http.StatusOK, nil
	}
	serverURL, cleanup, writeErr := newJSONResponderTestServer[jsonRequestParams, jsonUnmarshalableResponse](t, responderFunc)
	defer cleanup()

	response := postJSON(t, serverURL, `{"id":456}`)
	defer func() { assert.NoError(t, response.Body.Close()) }()

	assert.Equals(t, response.StatusCode, http.StatusInternalServerError)
	assert.NoError(t, writeErr())
}

func TestJSON_WriterReturnsError_CallsWriteErrorCallback(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	errWriter := &errorWriter{
		WriteFailed:    false,
		ResponseWriter: recorder,
	}

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		responders.JSON[jsonRequestParams, jsonResponseBody](errWriter, r, jsonTestHandler, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		server.URL,
		strings.NewReader(`{"id":123}`),
	)
	assert.NoError(t, err)
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)
	client := &http.Client{}
	response, err := client.Do(req)
	assert.NoError(t, err)
	assert.True(t, errWriter.WriteFailed)
	assert.ErrorPart(t, writeError, "simulated write failure")
	assert.NoError(t, response.Body.Close())
}

func TestJSON_NilResponseBody_RespondsWithNullJSON(t *testing.T) {
	t.Parallel()

	serverURL, cleanup, writeErr := newJSONResponderTestServer[jsonRequestParams, jsonResponseBody](t, func(*jsonRequestParams) (*jsonResponseBody, int, error) {
		return nil, http.StatusOK, nil
	})
	defer cleanup()

	response := postJSON(t, serverURL, `{"id":123}`)
	defer func() { assert.NoError(t, response.Body.Close()) }()

	assert.Equals(t, response.StatusCode, http.StatusOK)
	assert.Equals(t, response.Header.Get(headers.ContentType), headers.ContentTypeApplicationJSON)
	assert.NoError(t, writeErr())

	var body *jsonResponseBody
	assert.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	assert.True(t, body == nil)
}
