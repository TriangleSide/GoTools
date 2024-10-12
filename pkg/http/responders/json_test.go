package responders_test

import (
	"encoding/json"
	goerrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/responders"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

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
		return nil, 0, &errors.BadRequest{Err: goerrors.New("invalid parameters")}
	}

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		responders.JSON[requestParams, responseBody](w, r, jsonHandler)
	}

	t.Run("when valid request is made it responds with JSON and correct status code", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(httpHandler))
		defer server.Close()

		response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":123}`))
		assert.NoError(t, err)
		assert.Equals(t, response.StatusCode, http.StatusOK)

		body := &responseBody{}
		assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
		assert.Equals(t, body.Message, "processed")
		assert.NoError(t, response.Body.Close())
	})

	t.Run("when the parameter decoder fails it responds with error JSON and appropriate status code", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(httpHandler))
		defer server.Close()

		response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":-1}`))
		assert.NoError(t, err)
		assert.Equals(t, response.StatusCode, http.StatusBadRequest)

		body := &errors.Error{}
		assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
		assert.Contains(t, body.Message, "validation failed on field 'ID'")
		assert.NoError(t, response.Body.Close())
	})

	t.Run("when callback function returns error it responds with error JSON and appropriate status code", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(httpHandler))
		defer server.Close()

		response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":456}`))
		assert.NoError(t, err)
		assert.Equals(t, response.StatusCode, http.StatusBadRequest)

		body := &errors.Error{}
		assert.NoError(t, json.NewDecoder(response.Body).Decode(body))
		assert.Equals(t, body.Message, "invalid parameters")
		assert.NoError(t, response.Body.Close())
	})

	t.Run("when callback function returns unencodable response it should not write body", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responders.JSON[requestParams, unmarshalableResponse](w, r, func(params *requestParams) (*unmarshalableResponse, int, error) {
				return &unmarshalableResponse{}, http.StatusOK, nil
			})
		}))
		defer server.Close()

		response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":456}`))
		assert.NoError(t, err)
		assert.Equals(t, response.StatusCode, http.StatusOK)

		body := make(map[string]interface{})
		err = json.NewDecoder(response.Body).Decode(&body)
		assert.Error(t, err)
		assert.NoError(t, response.Body.Close())
	})
}
