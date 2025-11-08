package responders_test

import (
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

	response, err := http.Post(url, headers.ContentTypeApplicationJson, strings.NewReader(body))
	assert.NoError(t, err)

	return response
}

func TestJSONResponder(t *testing.T) {
	t.Parallel()

	type requestParams struct {
		ID int `json:"id" validate:"gt=0"`
	}

	type responseBody struct {
		Message string `json:"message"`
	}

	type unmarshalableResponse struct {
		ChanField chan int `json:"chan_field"`
	}

	jsonHandler := func(params *requestParams) (*responseBody, int, error) {
		if params.ID == 123 {
			return &responseBody{Message: "processed"}, http.StatusOK, nil
		}
		return nil, 0, &testError{}
	}

	t.Run("when valid request is made it responds with JSON and correct status code", func(t *testing.T) {
		t.Parallel()

		serverURL, cleanup, writeErr := newJSONResponderTestServer[requestParams, responseBody](t, jsonHandler)
		defer cleanup()

		response := postJSON(t, serverURL, `{"id":123}`)
		defer func() { assert.NoError(t, response.Body.Close()) }()

		assert.Equals(t, response.StatusCode, http.StatusOK)
		assert.Equals(t, response.Header.Get(headers.ContentType), headers.ContentTypeApplicationJson)
		assert.NoError(t, writeErr())

		body := &responseBody{}
		assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
		assert.Equals(t, body.Message, "processed")
	})

	t.Run("when the parameter decoder fails it responds with error JSON and appropriate status code", func(t *testing.T) {
		t.Parallel()

		serverURL, cleanup, writeErr := newJSONResponderTestServer[requestParams, responseBody](t, jsonHandler)
		defer cleanup()

		response := postJSON(t, serverURL, `{"id":-1}`)
		defer func() { assert.NoError(t, response.Body.Close()) }()

		assert.Equals(t, response.StatusCode, http.StatusBadRequest)
		assert.NoError(t, writeErr())

		body := &responders.StandardErrorResponse{}
		assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
		assert.Contains(t, body.Message, "validation failed on field 'ID'")
	})

	t.Run("when callback function returns error it responds with error JSON and appropriate status code", func(t *testing.T) {
		t.Parallel()

		serverURL, cleanup, writeErr := newJSONResponderTestServer[requestParams, responseBody](t, jsonHandler)
		defer cleanup()

		response := postJSON(t, serverURL, `{"id":456}`)
		defer func() { assert.NoError(t, response.Body.Close()) }()

		assert.Equals(t, response.StatusCode, http.StatusBadRequest)
		assert.NoError(t, writeErr())

		body := &responders.StandardErrorResponse{}
		assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
		assert.Equals(t, body.Message, "test error")
	})

	t.Run("when callback function returns unencodable response it should not write body", func(t *testing.T) {
		t.Parallel()

		serverURL, cleanup, writeErr := newJSONResponderTestServer[requestParams, unmarshalableResponse](t, func(params *requestParams) (*unmarshalableResponse, int, error) {
			return &unmarshalableResponse{}, http.StatusOK, nil
		})
		defer cleanup()

		response := postJSON(t, serverURL, `{"id":456}`)
		defer func() { assert.NoError(t, response.Body.Close()) }()

		assert.Equals(t, response.StatusCode, http.StatusInternalServerError)
		assert.NoError(t, writeErr())
	})

	t.Run("when the writer returns an error it should call the write error callback", func(t *testing.T) {
		t.Parallel()

		recorder := httptest.NewRecorder()
		ew := &errorWriter{
			WriteFailed:    false,
			ResponseWriter: recorder,
		}

		var writeError error
		writeErrorCallback := func(err error) {
			writeError = err
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responders.JSON[requestParams, responseBody](ew, r, jsonHandler, responders.WithErrorCallback(writeErrorCallback))
		}))
		defer server.Close()

		response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":123}`))
		assert.NoError(t, err)
		assert.True(t, ew.WriteFailed)
		assert.ErrorPart(t, writeError, "simulated write failure")
		assert.NoError(t, response.Body.Close())
	})
}
