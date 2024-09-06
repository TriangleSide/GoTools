package responders_test

import (
	"context"
	"encoding/json"
	goerrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/responders"
)

var _ = Describe("json stream responder", func() {
	type unmarshalableResponse struct {
		ChanField chan int `json:"chan_field"`
	}

	type requestParams struct {
		ID int `json:"id" validate:"gt=0"`
	}

	type responseBody struct {
		Message string `json:"message"`
	}

	var (
		cancel <-chan struct{}
	)

	BeforeEach(func() {
		cancel = make(chan struct{})
	})

	jsonStreamHandler := func(params *requestParams, cancelChan <-chan struct{}) (<-chan *responseBody, int, error) {
		cancel = cancelChan
		if params.ID == 1 {
			ch := make(chan *responseBody)
			go func() {
				defer close(ch)
				ch <- &responseBody{Message: "first"}
				ch <- &responseBody{Message: "second"}
			}()
			return ch, http.StatusOK, nil
		}
		return nil, 0, &errors.BadRequest{Err: goerrors.New("invalid parameters")}
	}

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		responders.JSONStream[requestParams, responseBody](w, r, jsonStreamHandler)
	}

	When("When the callback function processes the request successfully", func() {
		It("should respond with the correct JSON stream response and status code", func() {
			server := httptest.NewServer(http.HandlerFunc(httpHandler))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":1}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			decoder := json.NewDecoder(response.Body)
			responseObj := &responseBody{}
			Expect(decoder.Decode(responseObj)).To(Succeed())
			Expect(responseObj.Message).To(Equal("first"))
			Expect(decoder.Decode(responseObj)).To(Succeed())
			Expect(responseObj.Message).To(Equal("second"))
			Expect(response.Body.Close()).To(Succeed())

			Eventually(cancel).WithPolling(time.Millisecond).WithTimeout(time.Minute).Should(BeClosed())
		})
	})

	When("When the parameter decoder fails", func() {
		It("should respond with an error JSON response and appropriate status code", func() {
			server := httptest.NewServer(http.HandlerFunc(httpHandler))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":-1}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))

			responseObj := &errors.Error{}
			Expect(json.NewDecoder(response.Body).Decode(responseObj)).To(Succeed())
			Expect(responseObj.Message).To(ContainSubstring("validation failed on field 'ID'"))
			Expect(response.Body.Close()).To(Succeed())

			Expect(cancel).ToNot(BeClosed())
		})
	})

	When("When the callback function returns an error", func() {
		It("should respond with an error JSON response and appropriate status code", func() {
			server := httptest.NewServer(http.HandlerFunc(httpHandler))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":2}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))

			responseObj := &errors.Error{}
			Expect(json.NewDecoder(response.Body).Decode(responseObj)).To(Succeed())
			Expect(responseObj.Message).To(Equal("invalid parameters"))
			Expect(response.Body.Close()).To(Succeed())

			Eventually(cancel).WithPolling(time.Millisecond).WithTimeout(time.Minute).Should(BeClosed())
		})
	})

	When("the callback function returns a response that cannot be encoded", func() {
		It("should not write the body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				responders.JSONStream[requestParams, unmarshalableResponse](w, r, func(params *requestParams, cancelChan <-chan struct{}) (<-chan *unmarshalableResponse, int, error) {
					cancel = cancelChan
					ch := make(chan *unmarshalableResponse, 1)
					go func() {
						defer close(ch)
						ch <- &unmarshalableResponse{}
					}()
					return ch, http.StatusOK, nil
				})
			}))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":456}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			body := make(map[string]interface{})
			err = json.NewDecoder(response.Body).Decode(&body)
			Expect(err).To(HaveOccurred())
			Expect(response.Body.Close()).To(Succeed())

			Eventually(cancel).WithPolling(time.Millisecond).WithTimeout(time.Minute).Should(BeClosed())
		})
	})

	When("the request context is cancelled when streaming json", func() {
		It("should not write the any data to the response body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx, cancelFunc := context.WithCancel(r.Context())
				r = r.WithContext(ctx)
				cancelFunc()
				responders.JSONStream[requestParams, responseBody](w, r, func(params *requestParams, cancelChan <-chan struct{}) (<-chan *responseBody, int, error) {
					cancel = cancelChan
					ch := make(chan *responseBody)
					go func() {
						defer close(ch)
						ch <- &responseBody{Message: "first"}
					}()
					return ch, http.StatusOK, nil
				}, responders.WithDeferredConsumerTimerDuration(0))
			}))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":456}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			body := make(map[string]interface{})
			err = json.NewDecoder(response.Body).Decode(&body)
			Expect(err).To(HaveOccurred())
			Expect(response.Body.Close()).To(Succeed())

			Eventually(cancel).WithPolling(time.Millisecond).WithTimeout(time.Minute).Should(BeClosed())
		})
	})
})
