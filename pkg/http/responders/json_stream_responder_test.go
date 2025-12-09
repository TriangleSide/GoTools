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

type jsonStreamRequestParams struct {
	ID int `json:"id" validate:"gt=0"`
}

type jsonStreamResponseBody struct {
	Message string `json:"message"`
}

func TestJSONStream_SuccessfulCallback_RespondsWithCorrectJSONStreamAndStatusCode(t *testing.T) {
	t.Parallel()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](w, r, func(params *jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			ch := make(chan *jsonStreamResponseBody)
			go func() {
				defer close(ch)
				ch <- &jsonStreamResponseBody{Message: "first"}
				ch <- &jsonStreamResponseBody{Message: "second"}
			}()
			return ch, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":1}`))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusOK)
	assert.NoError(t, writeError)

	decoder := json.NewDecoder(response.Body)
	responseObj := &jsonStreamResponseBody{}
	assert.NoError(t, decoder.Decode(responseObj))
	assert.Equals(t, responseObj.Message, "first")
	assert.NoError(t, decoder.Decode(responseObj))
	assert.Equals(t, responseObj.Message, "second")
	assert.NoError(t, response.Body.Close())
}

func TestJSONStream_ParameterDecoderFails_RespondsWithErrorJSONAndBadRequest(t *testing.T) {
	t.Parallel()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](w, r, func(params *jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			return nil, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":-1}`))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusBadRequest)
	assert.NoError(t, writeError)

	responseObj := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(response.Body).Decode(responseObj))
	assert.Contains(t, responseObj.Message, "validation failed on field 'ID'")
	assert.NoError(t, response.Body.Close())
}

func TestJSONStream_CallbackReturnsError_RespondsWithErrorJSONAndBadRequest(t *testing.T) {
	t.Parallel()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](w, r, func(params *jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			return nil, 0, &testError{}
		}, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":2}`))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusBadRequest)
	assert.NoError(t, writeError)

	responseObj := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(response.Body).Decode(responseObj))
	assert.Equals(t, responseObj.Message, "test error")
	assert.NoError(t, response.Body.Close())
}

func TestJSONStream_CallbackReturnsNilChannel_RespondsWithInternalServerError(t *testing.T) {
	t.Parallel()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](w, r, func(params *jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			return nil, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":2}`))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusInternalServerError)
	assert.NoError(t, writeError)

	responseObj := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(response.Body).Decode(responseObj))
	assert.Equals(t, responseObj.Message, http.StatusText(http.StatusInternalServerError))
	assert.NoError(t, response.Body.Close())
}

func TestJSONStream_UnencodableResponse_DoesNotWriteBody(t *testing.T) {
	t.Parallel()

	type unmarshalableResponse struct {
		ChanField chan int `json:"chan_field"`
	}

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.JSONStream[jsonStreamRequestParams, unmarshalableResponse](w, r, func(params *jsonStreamRequestParams) (<-chan *unmarshalableResponse, int, error) {
			ch := make(chan *unmarshalableResponse, 1)
			go func() {
				defer close(ch)
				ch <- &unmarshalableResponse{}
			}()
			return ch, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":3}`))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusOK)
	assert.Error(t, writeError)

	body := make(map[string]any)
	err = json.NewDecoder(response.Body).Decode(&body)
	assert.Error(t, err)
	assert.NoError(t, response.Body.Close())
}

func TestJSONStream_RequestContextCancelled_DoesNotWriteData(t *testing.T) {
	t.Parallel()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		r = r.WithContext(ctx)
		cancel()
		responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](w, r, func(params *jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			<-r.Context().Done()
			ch := make(chan *jsonStreamResponseBody)
			go func() {
				defer close(ch)
				ch <- &jsonStreamResponseBody{Message: "first"}
			}()
			return ch, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":4}`))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusOK)
	assert.NoError(t, writeError)

	body := make(map[string]any)
	err = json.NewDecoder(response.Body).Decode(&body)
	assert.ErrorPart(t, err, "EOF")
	assert.NoError(t, response.Body.Close())
}

func TestJSONStream_WriterFails_CallsErrorCallback(t *testing.T) {
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
		responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](ew, r, func(params *jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			ch := make(chan *jsonStreamResponseBody, 1)
			go func() {
				defer close(ch)
				ch <- &jsonStreamResponseBody{}
			}()
			return ch, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":3}`))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusOK)
	assert.ErrorPart(t, writeError, "simulated write failure")
	assert.NoError(t, response.Body.Close())
}

func TestJSONStream_ChannelClosedImmediately_RespondsWithEmptyBody(t *testing.T) {
	t.Parallel()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](w, r, func(params *jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			ch := make(chan *jsonStreamResponseBody)
			go func() {
				close(ch)
			}()
			return ch, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":1}`))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusOK)
	assert.NoError(t, writeError)

	body := make(map[string]any)
	err = json.NewDecoder(response.Body).Decode(&body)
	assert.ErrorPart(t, err, "EOF")
	assert.NoError(t, response.Body.Close())
}
