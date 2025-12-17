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

type jsonRequestParams struct {
	ID int `json:"id" validate:"gt=0"`
}

type jsonResponseBody struct {
	Message string `json:"message"`
}

type jsonUnmarshalable struct {
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

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":123}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	responders.JSON[jsonRequestParams, jsonResponseBody](
		recorder, req, jsonTestHandler, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusOK)
	assert.Equals(t, recorder.Header().Get(headers.ContentType), headers.ContentTypeApplicationJSON)
	assert.NoError(t, writeErr)

	body := &jsonResponseBody{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(body))
	assert.Equals(t, body.Message, "processed")
}

func TestJSON_ParameterDecoderFails_RespondsWithErrorJSONAndBadRequestStatus(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":-1}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	responders.JSON[jsonRequestParams, jsonResponseBody](
		recorder, req, jsonTestHandler, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	assert.NoError(t, writeErr)

	body := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(body))
	assert.Contains(t, body.Message, "validation failed on field 'ID'")
}

func TestJSON_CallbackReturnsError_RespondsWithErrorJSONAndBadRequestStatus(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":456}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	responders.JSON[jsonRequestParams, jsonResponseBody](
		recorder, req, jsonTestHandler, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	assert.NoError(t, writeErr)

	body := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(body))
	assert.Equals(t, body.Message, "test error")
}

func TestJSON_UnencodableResponse_ReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":456}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	responderFunc := func(*jsonRequestParams) (*jsonUnmarshalable, int, error) {
		return &jsonUnmarshalable{}, http.StatusOK, nil
	}

	responders.JSON[jsonRequestParams, jsonUnmarshalable](
		recorder, req, responderFunc, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusInternalServerError)
	assert.NoError(t, writeErr)
}

func TestJSON_WriterReturnsError_CallsWriteErrorCallback(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	errWriter := &errorWriter{
		WriteFailed:    false,
		ResponseWriter: recorder,
	}

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":123}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	responders.JSON[jsonRequestParams, jsonResponseBody](
		errWriter, req, jsonTestHandler, responders.WithErrorCallback(writeErrorCallback))

	assert.True(t, errWriter.WriteFailed)
	assert.ErrorPart(t, writeErr, "simulated write failure")
}

func TestJSON_NilResponseBody_RespondsWithNullJSON(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":123}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	responderFunc := func(*jsonRequestParams) (*jsonResponseBody, int, error) {
		return nil, http.StatusOK, nil
	}

	responders.JSON[jsonRequestParams, jsonResponseBody](
		recorder, req, responderFunc, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusOK)
	assert.Equals(t, recorder.Header().Get(headers.ContentType), headers.ContentTypeApplicationJSON)
	assert.NoError(t, writeErr)

	var body *jsonResponseBody
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))
	assert.True(t, body == nil)
}
