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

func init() {
	responders.MustRegisterErrorResponse[testError](http.StatusBadRequest, func(err *testError) string {
		return err.Error()
	})
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

	t.Run("when the error response is has a message it should be able to be marshalled", func(t *testing.T) {
		t.Parallel()
		_, err := json.Marshal(&responders.ErrorResponse{
			Message: "test",
		})
		assert.NoError(t, err)
	})

	t.Run("when the error response is empty it should be able to be marshalled", func(t *testing.T) {
		t.Parallel()
		_, err := json.Marshal(&responders.ErrorResponse{})
		assert.NoError(t, err)
	})

	t.Run("when the option to return an error is registered twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			responders.MustRegisterErrorResponse[testError](http.StatusBadRequest, func(err *testError) string {
				return "registered twice"
			})
		})
	})

	t.Run("when a pointer generic is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			responders.MustRegisterErrorResponse[*testError](http.StatusBadRequest, func(err **testError) string {
				return "pointer is registered"
			})
		}, "registered error responses must be a struct")
	})

	t.Run("when a struct that is not an error is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			responders.MustRegisterErrorResponse[struct{}](http.StatusBadRequest, func(err *struct{}) string {
				return "error"
			})
		}, "must have an error interface")
	})

	t.Run("when the error is unknown it should return internal server error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		standardError := errors.New("standard error")
		err := responders.Error(recorder, standardError)
		assert.NoError(t, err)
		assert.Equals(t, recorder.Code, http.StatusInternalServerError)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, http.StatusText(http.StatusInternalServerError))
	})

	t.Run("when the error is nil it should return internal server error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		err := responders.Error(recorder, nil)
		assert.NoError(t, err)
		assert.Equals(t, recorder.Code, http.StatusInternalServerError)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, http.StatusText(http.StatusInternalServerError))
	})

	t.Run("when the error is a custom registered type it should return its custom message and status", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		err := responders.Error(recorder, &testError{})
		assert.NoError(t, err)
		assert.Equals(t, recorder.Code, http.StatusBadRequest)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, "test error")
	})

	t.Run("when the error is joined with a a custom registered type it should return its custom message and status", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		err := responders.Error(recorder, errors.Join(&testError{}, errors.New("other error")))
		assert.NoError(t, err)
		assert.Equals(t, recorder.Code, http.StatusBadRequest)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, "test error")
	})

	t.Run("when the writer returns an error it should return an error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		ew := &errorWriter{
			WriteFailed:    false,
			ResponseWriter: recorder,
		}
		err := responders.Error(ew, errors.New("some error"))
		assert.ErrorPart(t, err, "failed to write error response (simulated write failure)")
		assert.True(t, ew.WriteFailed)
	})
}
