package responders_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/http/responders"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
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
	responders.MustRegisterErrorResponse(http.StatusBadRequest, func(*testUnmarshalableError) *struct{ C chan int } {
		return &struct{ C chan int }{}
	})
}

func TestError_UnknownError_ReturnsInternalServerError(t *testing.T) {
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
}

func TestError_NilError_ReturnsInternalServerError(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	var writeError error
	writeErrorCallback := func(err error) { writeError = err }
	responders.Error(recorder, nil, responders.WithErrorCallback(writeErrorCallback))
	assert.Equals(t, recorder.Code, http.StatusInternalServerError)
	httpError := mustDeserializeError(t, recorder)
	assert.Equals(t, httpError.Message, http.StatusText(http.StatusInternalServerError))
	assert.NoError(t, writeError)
}

func TestError_CustomRegisteredType_ReturnsCustomMessageAndStatus(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	var writeError error
	writeErrorCallback := func(err error) { writeError = err }
	responders.Error(recorder, &testError{}, responders.WithErrorCallback(writeErrorCallback))
	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	httpError := mustDeserializeError(t, recorder)
	assert.Equals(t, httpError.Message, "test error")
	assert.NoError(t, writeError)
}

func TestError_JoinedWithCustomRegisteredType_ReturnsCustomMessageAndStatus(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	var writeError error
	writeErrorCallback := func(err error) { writeError = err }
	responders.Error(recorder, errors.Join(&testError{}, errors.New("other error")), responders.WithErrorCallback(writeErrorCallback))
	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	httpError := mustDeserializeError(t, recorder)
	assert.Equals(t, httpError.Message, "test error")
	assert.NoError(t, writeError)
}

func TestError_WrappedWithCustomRegisteredType_ReturnsCustomMessageAndStatus(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	var writeError error
	writeErrorCallback := func(err error) { writeError = err }
	wrappedErr := fmt.Errorf("wrapped: %w", &testError{})
	responders.Error(recorder, wrappedErr, responders.WithErrorCallback(writeErrorCallback))
	assert.Equals(t, recorder.Code, http.StatusBadRequest)
	httpError := mustDeserializeError(t, recorder)
	assert.Equals(t, httpError.Message, "test error")
	assert.NoError(t, writeError)
}

func TestError_WriterError_InvokesCallback(t *testing.T) {
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
}

func TestError_UnmarshalableErrorResponse_InvokesCallback(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	var writeError error
	writeErrorCallback := func(err error) { writeError = err }
	responders.Error(recorder, &testUnmarshalableError{}, responders.WithErrorCallback(writeErrorCallback))
	assert.Equals(t, recorder.Code, http.StatusInternalServerError)
	httpError := mustDeserializeError(t, recorder)
	assert.Equals(t, httpError.Message, http.StatusText(http.StatusInternalServerError))
	assert.ErrorPart(t, writeError, "json")
}

func TestError_FallbackMarshallerFails_InvokesCallback(t *testing.T) {
	t.Parallel()
	marshalErr := errors.New("marshal failure")
	var marshalErrors []error
	writeErrorCallback := func(err error) {
		marshalErrors = append(marshalErrors, err)
	}
	marshal := func(any) ([]byte, error) {
		return nil, marshalErr
	}
	recorder := httptest.NewRecorder()
	responders.Error(recorder, errors.New("initial error"), responders.WithJSONMarshal(marshal), responders.WithErrorCallback(writeErrorCallback))
	assert.Equals(t, len(marshalErrors), 2)
	assert.Equals(t, recorder.Code, http.StatusInternalServerError)
	assert.Equals(t, marshalErrors[0], marshalErr)
	assert.Equals(t, marshalErrors[1], marshalErr)
}
