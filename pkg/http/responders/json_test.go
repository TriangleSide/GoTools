package responders_test

import (
	"encoding/json"
	goerrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/http/errors"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/responders"
)

var _ = Describe("JSON responder", func() {
	type unmarshalableResponse struct {
		ChanField chan int `json:"chan_field"`
	}

	type requestParams struct {
		ID int `json:"id" validate:"gt=0"`
	}

	type responseBody struct {
		Message string `json:"message"`
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

	When("the callback function processes the request successfully", func() {
		It("should respond with the correct JSON response and status code", func() {
			server := httptest.NewServer(http.HandlerFunc(httpHandler))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":123}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			responseBody := &responseBody{}
			Expect(json.NewDecoder(response.Body).Decode(responseBody)).To(Succeed())
			Expect(responseBody.Message).To(Equal("processed"))
			Expect(response.Body.Close()).To(Succeed())
		})
	})

	When("the parameter decoder fails", func() {
		It("should respond with an error JSON response and appropriate status code", func() {
			server := httptest.NewServer(http.HandlerFunc(httpHandler))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":-1}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))

			responseBody := &errors.Error{}
			Expect(json.NewDecoder(response.Body).Decode(responseBody)).To(Succeed())
			Expect(responseBody.Message).To(ContainSubstring("validation failed on field 'ID'"))
			Expect(response.Body.Close()).To(Succeed())
		})
	})

	When("the callback function returns an error", func() {
		It("should respond with an error JSON response and appropriate status code", func() {
			server := httptest.NewServer(http.HandlerFunc(httpHandler))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":456}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusBadRequest))

			responseBody := &errors.Error{}
			Expect(json.NewDecoder(response.Body).Decode(responseBody)).To(Succeed())
			Expect(responseBody.Message).To(Equal("invalid parameters"))
			Expect(response.Body.Close()).To(Succeed())
		})
	})

	When("the callback function returns a response that cannot be encoded", func() {
		It("should not write the body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				responders.JSON[requestParams, unmarshalableResponse](w, r, func(params *requestParams) (*unmarshalableResponse, int, error) {
					return &unmarshalableResponse{}, http.StatusOK, nil
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
		})
	})
})
