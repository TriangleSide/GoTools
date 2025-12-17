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

type statusRequestParams struct {
	ID int `json:"id" validate:"gt=0"`
}

func statusHandler(params *statusRequestParams) (int, error) {
	if params.ID == 123 {
		return http.StatusOK, nil
	}
	return 0, &testError{}
}

func TestStatus_CallbackSuccess_ReturnsCorrectStatusCode(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":123}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	responders.Status[statusRequestParams](
		recorder, req, statusHandler, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusOK)
	assert.NoError(t, writeErr)
}

func TestStatus_ParameterDecoderFails_ReturnsErrorResponse(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":-1}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	responders.Status[statusRequestParams](
		recorder, req, statusHandler, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	assert.NoError(t, writeErr)

	body := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(body))
	assert.Contains(t, body.Message, "validation failed on field 'ID'")
}

func TestStatus_CallbackReturnsError_ReturnsErrorResponse(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{"id":456}`))
	req.Header.Set(headers.ContentType, headers.ContentTypeApplicationJSON)

	var writeErr error
	writeErrorCallback := func(err error) { writeErr = err }

	responders.Status[statusRequestParams](
		recorder, req, statusHandler, responders.WithErrorCallback(writeErrorCallback))

	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	assert.NoError(t, writeErr)

	body := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(body))
	assert.Equals(t, body.Message, "test error")
}
