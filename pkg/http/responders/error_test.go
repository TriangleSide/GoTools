package responders_test

import (
	"encoding/json"
	goerrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/responders"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

type testError struct{}

func (e testError) Error() string {
	return "test error"
}

type failingWriter struct {
	WriteFailed bool
	http.ResponseWriter
}

func (fw *failingWriter) Write([]byte) (int, error) {
	fw.WriteFailed = true
	return 0, goerrors.New("simulated write failure")
}

func mustDeserializeError(t *testing.T, recorder *httptest.ResponseRecorder) *errors.Error {
	t.Helper()
	httpError := &errors.Error{}
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(httpError))
	return httpError
}

func TestErrorResponder(t *testing.T) {
	t.Parallel()

	responders.MustRegisterErrorResponse[testError](http.StatusFound, func(err *testError) string {
		return "custom message"
	})

	t.Run("when the option to return an error is registered twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			responders.MustRegisterErrorResponse[testError](http.StatusFound, func(err *testError) string {
				return "registered twice"
			})
		})
	})

	t.Run("when a pointer generic is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			responders.MustRegisterErrorResponse[*testError](http.StatusFound, func(err **testError) string {
				return "pointer is registered"
			})
		}, "cannot be a pointer")
	})

	t.Run("when the error is unknown it should return internal server error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		standardError := goerrors.New("standard error")
		responders.Error(recorder, standardError)
		assert.Equals(t, recorder.Code, http.StatusInternalServerError)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, http.StatusText(http.StatusInternalServerError))
	})

	t.Run("when the error is known it should return the correct status and message", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		badRequestErr := &errors.BadRequest{
			Err: goerrors.New("bad request"),
		}
		responders.Error(recorder, badRequestErr)
		assert.Equals(t, recorder.Code, http.StatusBadRequest)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, badRequestErr.Error())
	})

	t.Run("when the error is nil it should return internal server error", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		responders.Error(recorder, nil)
		assert.Equals(t, recorder.Code, http.StatusInternalServerError)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, http.StatusText(http.StatusInternalServerError))
	})

	t.Run("when the error is a custom registered type it should return its custom message and status", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		responders.Error(recorder, &testError{})
		assert.Equals(t, recorder.Code, http.StatusFound)
		httpError := mustDeserializeError(t, recorder)
		assert.Equals(t, httpError.Message, "custom message")
	})

	t.Run("when the JSON encoding fails it should not write a response", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		fw := &failingWriter{
			WriteFailed:    false,
			ResponseWriter: recorder,
		}
		responders.Error(fw, goerrors.New("some error"))
		assert.True(t, fw.WriteFailed)
	})
}
