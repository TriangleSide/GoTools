package responders_test

import (
	"encoding/json"
	goerrors "errors"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/responders"
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

func mustDeserializeError(recorder *httptest.ResponseRecorder) *errors.Error {
	httpError := &errors.Error{}
	Expect(json.NewDecoder(recorder.Body).Decode(httpError)).To(Succeed())
	return httpError
}

var _ = Describe("error responder", Ordered, func() {
	var (
		standardError error
		recorder      *httptest.ResponseRecorder
	)

	BeforeAll(func() {
		responders.MustRegisterErrorResponse[testError](http.StatusFound, func(err *testError) string {
			return "custom message"
		})
	})

	BeforeEach(func() {
		standardError = goerrors.New("standard error")
		recorder = httptest.NewRecorder()
	})

	When("an error type is registered twice", func() {
		It("should panic", func() {
			Expect(func() {
				responders.MustRegisterErrorResponse[testError](http.StatusFound, func(err *testError) string {
					return "registered twice"
				})
			}).To(Panic())
		})
	})

	When("a pointer generic is registered", func() {
		It("should panic", func() {
			Expect(func() {
				responders.MustRegisterErrorResponse[*testError](http.StatusFound, func(err **testError) string {
					return "pointer is registered"
				})
			}).To(Panic())
		})
	})

	When("the error is unknown", func() {
		It("should return an internal server error", func() {
			responders.Error(recorder, standardError)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			httpError := mustDeserializeError(recorder)
			Expect(httpError.Message).To(Equal(http.StatusText(http.StatusInternalServerError)))
		})
	})

	When("the error is known", func() {
		It("should return the correct status and message", func() {
			badRequestErr := &errors.BadRequest{
				Err: goerrors.New("bad request"),
			}
			responders.Error(recorder, badRequestErr)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			httpError := mustDeserializeError(recorder)
			Expect(httpError.Message).To(Equal(badRequestErr.Error()))
		})
	})

	When("the error is nil", func() {
		It("should return an internal server error", func() {
			responders.Error(recorder, nil)
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			httpError := mustDeserializeError(recorder)
			Expect(httpError.Message).To(Equal(http.StatusText(http.StatusInternalServerError)))
		})
	})

	When("the error is a custom registered type", func() {
		It("should return its custom message and status", func() {
			responders.Error(recorder, &testError{})
			Expect(recorder.Code).To(Equal(http.StatusFound))
			httpError := mustDeserializeError(recorder)
			Expect(httpError.Message).To(Equal("custom message"))
		})
	})

	When("the JSON encoding fails", func() {
		It("not write a response", func() {
			recorder := httptest.NewRecorder()
			failingWriter := &failingWriter{
				WriteFailed:    false,
				ResponseWriter: recorder,
			}
			responders.Error(failingWriter, goerrors.New("some error"))
			Expect(failingWriter.WriteFailed).To(BeTrue())
		})
	})
})
