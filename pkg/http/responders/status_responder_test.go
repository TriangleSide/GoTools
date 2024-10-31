package responders_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/responders"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestStatus(t *testing.T) {
	t.Parallel()

	type requestParams struct {
		ID int `json:"id" validate:"gt=0"`
	}

	statusHandler := func(params *requestParams) (int, error) {
		if params.ID == 123 {
			return http.StatusOK, nil
		}
		return 0, &testError{}
	}

	t.Run("when the callback function processes the request successfully it should respond with the correct status code", func(t *testing.T) {
		t.Parallel()

		var writeError error
		writeErrorCallback := func(err error) {
			writeError = err
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responders.Status[requestParams](w, r, statusHandler, responders.WithWriteErrorCallback(writeErrorCallback))
		}))
		defer server.Close()

		response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":123}`))
		t.Cleanup(func() {
			assert.NoError(t, response.Body.Close())
		})
		assert.NoError(t, err)
		assert.Equals(t, response.StatusCode, http.StatusOK)
		assert.NoError(t, writeError)
	})

	t.Run("when the parameter decoder fails it should respond with an error JSON response and appropriate status code", func(t *testing.T) {
		t.Parallel()

		var writeError error
		writeErrorCallback := func(err error) {
			writeError = err
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responders.Status[requestParams](w, r, statusHandler, responders.WithWriteErrorCallback(writeErrorCallback))
		}))
		defer server.Close()

		response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":-1}`))
		assert.NoError(t, err)
		assert.Equals(t, response.StatusCode, http.StatusBadRequest)
		assert.NoError(t, writeError)

		responseBody := &responders.ErrorResponse{}
		assert.NoError(t, json.NewDecoder(response.Body).Decode(responseBody))
		assert.Contains(t, responseBody.Message, "validation failed on field 'ID'")
		assert.NoError(t, response.Body.Close())
	})

	t.Run("when the callback function returns an error it should respond with an error JSON response and appropriate status code", func(t *testing.T) {
		t.Parallel()

		var writeError error
		writeErrorCallback := func(err error) {
			writeError = err
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responders.Status[requestParams](w, r, statusHandler, responders.WithWriteErrorCallback(writeErrorCallback))
		}))
		defer server.Close()

		response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":456}`))
		assert.NoError(t, err)
		assert.Equals(t, response.StatusCode, http.StatusBadRequest)
		assert.NoError(t, writeError)

		responseBody := &responders.ErrorResponse{}
		assert.NoError(t, json.NewDecoder(response.Body).Decode(responseBody))
		assert.Equals(t, responseBody.Message, "test error")
		assert.NoError(t, response.Body.Close())
	})
}
