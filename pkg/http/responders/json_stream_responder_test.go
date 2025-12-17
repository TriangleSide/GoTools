package responders_test

import (
	"context"
	"encoding/json"
	"errors"
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

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":1}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](
		recorder, req, func(*jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			responseChan := make(chan *jsonStreamResponseBody)
			go func() {
				defer close(responseChan)
				responseChan <- &jsonStreamResponseBody{Message: "first"}
				responseChan <- &jsonStreamResponseBody{Message: "second"}
			}()
			return responseChan, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusOK)
	assert.NoError(t, writeError)

	decoder := json.NewDecoder(recorder.Body)
	responseObj := &jsonStreamResponseBody{}
	assert.NoError(t, decoder.Decode(responseObj))
	assert.Equals(t, responseObj.Message, "first")
	assert.NoError(t, decoder.Decode(responseObj))
	assert.Equals(t, responseObj.Message, "second")
}

func TestJSONStream_ParameterDecoderFails_RespondsWithErrorJSONAndBadRequest(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":-1}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](
		recorder, req, func(*jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			return nil, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	assert.NoError(t, writeError)

	responseObj := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(responseObj))
	assert.Contains(t, responseObj.Message, "validation failed on field 'ID'")
}

func TestJSONStream_CallbackReturnsError_RespondsWithErrorJSONAndBadRequest(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":2}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](
		recorder, req, func(*jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			return nil, 0, &testError{}
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	assert.NoError(t, writeError)

	responseObj := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(responseObj))
	assert.Equals(t, responseObj.Message, "test error")
}

func TestJSONStream_CallbackReturnsNilChannel_RespondsWithInternalServerError(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":2}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](
		recorder, req, func(*jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			return nil, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusInternalServerError)
	assert.NoError(t, writeError)

	responseObj := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(responseObj))
	assert.Equals(t, responseObj.Message, http.StatusText(http.StatusInternalServerError))
}

func TestJSONStream_UnencodableResponse_DoesNotWriteBody(t *testing.T) {
	t.Parallel()

	type unmarshalableResponse struct {
		ChanField chan int `json:"chan_field"`
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":3}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, unmarshalableResponse](
		recorder, req, func(*jsonStreamRequestParams) (<-chan *unmarshalableResponse, int, error) {
			ch := make(chan *unmarshalableResponse, 1)
			go func() {
				defer close(ch)
				ch <- &unmarshalableResponse{}
			}()
			return ch, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusOK)
	assert.Error(t, writeError)

	body := make(map[string]any)
	err := json.NewDecoder(recorder.Body).Decode(&body)
	assert.Error(t, err)
}

func TestJSONStream_RequestContextCancelled_DoesNotWriteData(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(`{"id":4}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)
	cancel()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](
		recorder, req, func(*jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			<-req.Context().Done()
			ch := make(chan *jsonStreamResponseBody)
			go func() {
				defer close(ch)
				ch <- &jsonStreamResponseBody{Message: "first"}
			}()
			return ch, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusOK)
	assert.NoError(t, writeError)

	body := make(map[string]any)
	err := json.NewDecoder(recorder.Body).Decode(&body)
	assert.ErrorPart(t, err, "EOF")
}

func TestJSONStream_WriterFails_CallsErrorCallback(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	errWriter := &errorWriter{
		WriteFailed:    false,
		ResponseWriter: recorder,
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":3}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](
		errWriter, req, func(*jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			responseChan := make(chan *jsonStreamResponseBody, 1)
			go func() {
				defer close(responseChan)
				responseChan <- &jsonStreamResponseBody{}
			}()
			return responseChan, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.ErrorPart(t, writeError, "simulated write failure")
}

func TestJSONStream_ChannelClosedImmediately_RespondsWithEmptyBody(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":1}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](
		recorder, req, func(*jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			ch := make(chan *jsonStreamResponseBody)
			go func() {
				close(ch)
			}()
			return ch, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusOK)
	assert.NoError(t, writeError)

	body := make(map[string]any)
	err := json.NewDecoder(recorder.Body).Decode(&body)
	assert.ErrorPart(t, err, "EOF")
}

type contextCancellingErrorWriter struct {
	http.ResponseWriter

	CancelFunc  context.CancelFunc
	WriteFailed bool
}

func (w *contextCancellingErrorWriter) Write([]byte) (int, error) {
	w.WriteFailed = true
	w.CancelFunc()
	return 0, errors.New("simulated write failure due to context cancellation")
}

func TestJSONStream_WriterFailsWhenContextCancelled_DoesNotCallErrorCallback(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(t.Context())
	errWriter := &contextCancellingErrorWriter{
		ResponseWriter: recorder,
		CancelFunc:     cancel,
		WriteFailed:    false,
	}

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(`{"id":3}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	responders.JSONStream[jsonStreamRequestParams, jsonStreamResponseBody](
		errWriter, req, func(*jsonStreamRequestParams) (<-chan *jsonStreamResponseBody, int, error) {
			responseChan := make(chan *jsonStreamResponseBody, 1)
			go func() {
				defer close(responseChan)
				responseChan <- &jsonStreamResponseBody{}
			}()
			return responseChan, http.StatusOK, nil
		}, responders.WithErrorCallback(writeErrorCallback))

	assert.True(t, errWriter.WriteFailed)
	assert.NoError(t, writeError)
}
