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

func statusErrorMessageTest(t *testing.T, jsonBody, expectedError string) {
	t.Helper()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.Status[statusRequestParams](w, r, statusHandler, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(jsonBody))
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusBadRequest)
	assert.NoError(t, writeError)

	responseBody := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(response.Body).Decode(responseBody))
	assert.Contains(t, responseBody.Message, expectedError)
	assert.NoError(t, response.Body.Close())
}

func TestStatus_CallbackSuccess_ReturnsCorrectStatusCode(t *testing.T) {
	t.Parallel()

	var writeError error
	writeErrorCallback := func(err error) {
		writeError = err
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responders.Status[statusRequestParams](w, r, statusHandler, responders.WithErrorCallback(writeErrorCallback))
	}))
	defer server.Close()

	response, err := http.Post(server.URL, headers.ContentTypeApplicationJSON, strings.NewReader(`{"id":123}`))
	t.Cleanup(func() {
		assert.NoError(t, response.Body.Close())
	})
	assert.NoError(t, err)
	assert.Equals(t, response.StatusCode, http.StatusOK)
	assert.NoError(t, writeError)
}

func TestStatus_ParameterDecoderFails_ReturnsErrorResponse(t *testing.T) {
	t.Parallel()
	statusErrorMessageTest(t, `{"id":-1}`, "validation failed on field 'ID'")
}

func TestStatus_CallbackReturnsError_ReturnsErrorResponse(t *testing.T) {
	t.Parallel()
	statusErrorMessageTest(t, `{"id":456}`, "test error")
}
