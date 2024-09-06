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

var _ = Describe("status responder", func() {
	type requestParams struct {
		ID int `json:"id" validate:"gt=0"`
	}

	statusHandler := func(params *requestParams) (int, error) {
		if params.ID == 123 {
			return http.StatusOK, nil
		}
		return 0, &errors.BadRequest{Err: goerrors.New("invalid parameters")}
	}

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		responders.Status[requestParams](w, r, statusHandler)
	}

	When("the callback function processes the request successfully", func() {
		It("should respond with the correct status code", func() {
			server := httptest.NewServer(http.HandlerFunc(httpHandler))
			defer server.Close()

			response, err := http.Post(server.URL, headers.ContentTypeApplicationJson, strings.NewReader(`{"id":123}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
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
})
