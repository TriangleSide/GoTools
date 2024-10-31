package responders_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/responders"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func init() {
	responders.MustRegisterErrorResponse[testError](http.StatusBadRequest, func(err *testError) string {
		return err.Error()
	})
}

type testError struct{}

func (e *testError) Error() string {
	return "test error"
}

type errorWriter struct {
	WriteFailed bool
	http.ResponseWriter
}

func (w *errorWriter) Write([]byte) (int, error) {
	w.WriteFailed = true
	return 0, errors.New("simulated write failure")
}

func mustDeserializeError(t *testing.T, recorder *httptest.ResponseRecorder) *responders.ErrorResponse {
	t.Helper()
	httpError := &responders.ErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(httpError))
	return httpError
}

func TestErrorResponder(t *testing.T) {
	t.Parallel()

	t.Run("when the error response is empty it should be able to be marshalled", func(t *testing.T) {
		t.Parallel()
		_, err := json.Marshal(&responders.ErrorResponse{})
		assert.NoError(t, err)
	})

	t.Run("when the error response is has a message it should be able to be marshalled", func(t *testing.T) {
		t.Parallel()
		_, err := json.Marshal(&responders.ErrorResponse{
			Message: "test",
		})
		assert.NoError(t, err)
	})

	t.Run("when the error is unknown it should return internal server error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		var writeError error
		writeErrorCallback := func(err error) { writeError = err }
		standardError := errors.New("standard error")
		responders.Error(recorder, standardError, responders.WithWriteErrorCallback(writeErrorCallback))
		assert.Equals(t, recorder.Code, http.StatusInternalServerError)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, http.StatusText(http.StatusInternalServerError))
		assert.NoError(t, writeError)
	})

	t.Run("when the error is nil it should return internal server error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		var writeError error
		writeErrorCallback := func(err error) { writeError = err }
		responders.Error(recorder, nil, responders.WithWriteErrorCallback(writeErrorCallback))
		assert.Equals(t, recorder.Code, http.StatusInternalServerError)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, http.StatusText(http.StatusInternalServerError))
		assert.NoError(t, writeError)
	})

	t.Run("when the error is a custom registered type it should return its custom message and status", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		var writeError error
		writeErrorCallback := func(err error) { writeError = err }
		responders.Error(recorder, &testError{}, responders.WithWriteErrorCallback(writeErrorCallback))
		assert.Equals(t, recorder.Code, http.StatusBadRequest)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, "test error")
		assert.NoError(t, writeError)
	})

	t.Run("when the error is joined with a a custom registered type it should return its custom message and status", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		var writeError error
		writeErrorCallback := func(err error) { writeError = err }
		responders.Error(recorder, errors.Join(&testError{}, errors.New("other error")), responders.WithWriteErrorCallback(writeErrorCallback))
		assert.Equals(t, recorder.Code, http.StatusBadRequest)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, "test error")
		assert.NoError(t, writeError)
	})

	t.Run("when the writer returns an error it should return an error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		ew := &errorWriter{
			WriteFailed:    false,
			ResponseWriter: recorder,
		}
		var writeError error
		writeErrorCallback := func(err error) { writeError = err }
		responders.Error(ew, errors.New("some error"), responders.WithWriteErrorCallback(writeErrorCallback))
		assert.True(t, ew.WriteFailed)
		assert.ErrorPart(t, writeError, "simulated write failure")
	})
}
