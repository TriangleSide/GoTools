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

type testError struct{}

func (e *testError) Error() string {
	return "test error"
}

type testUnmarshalableError struct{}

func (e *testUnmarshalableError) Error() string {
	return "unmarshalable error"
}

type errorWriter struct {
	WriteFailed bool
	http.ResponseWriter
}

func (w *errorWriter) Write([]byte) (int, error) {
	w.WriteFailed = true
	return 0, errors.New("simulated write failure")
}

func mustDeserializeError(t *testing.T, recorder *httptest.ResponseRecorder) *responders.StandardErrorResponse {
	t.Helper()
	httpError := &responders.StandardErrorResponse{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(httpError))
	return httpError
}

func init() {
	responders.MustRegisterErrorResponse(http.StatusBadRequest, func(err *testError) *responders.StandardErrorResponse {
		return &responders.StandardErrorResponse{
			Message: err.Error(),
		}
	})
	responders.MustRegisterErrorResponse(http.StatusBadRequest, func(err *testUnmarshalableError) *struct{ C chan int } {
		return &struct{ C chan int }{}
	})
}

func TestErrorResponder(t *testing.T) {
	t.Parallel()

	t.Run("when the error is unknown it should return internal server error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		var writeError error
		writeErrorCallback := func(err error) { writeError = err }
		standardError := errors.New("standard error")
		responders.Error(recorder, standardError, responders.WithErrorCallback(writeErrorCallback))
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
		responders.Error(recorder, nil, responders.WithErrorCallback(writeErrorCallback))
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
		responders.Error(recorder, &testError{}, responders.WithErrorCallback(writeErrorCallback))
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
		responders.Error(recorder, errors.Join(&testError{}, errors.New("other error")), responders.WithErrorCallback(writeErrorCallback))
		assert.Equals(t, recorder.Code, http.StatusBadRequest)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, "test error")
		assert.NoError(t, writeError)
	})

	t.Run("when the writer returns an error it should invoke to the callback", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		ew := &errorWriter{
			WriteFailed:    false,
			ResponseWriter: recorder,
		}
		var writeError error
		writeErrorCallback := func(err error) { writeError = err }
		responders.Error(ew, errors.New("some error"), responders.WithErrorCallback(writeErrorCallback))
		assert.True(t, ew.WriteFailed)
		assert.ErrorPart(t, writeError, "simulated write failure")
	})

	t.Run("when the error response cannot be marshalled it should invoke the callback", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		var writeError error
		writeErrorCallback := func(err error) { writeError = err }
		responders.Error(recorder, &testUnmarshalableError{}, responders.WithErrorCallback(writeErrorCallback))
		assert.ErrorPart(t, writeError, "json")
	})
}
