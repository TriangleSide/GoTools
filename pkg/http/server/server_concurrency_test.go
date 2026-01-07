package server_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/http/endpoints"
	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/responders"
	"github.com/TriangleSide/GoTools/pkg/http/server"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type concurrentTestCase struct {
	method      string
	path        string
	body        func() io.Reader
	contentType string
	expected    int
}

func runConcurrentRequest(t *testing.T, serverAddress string, testCase concurrentTestCase) {
	t.Helper()
	var body io.Reader
	if testCase.body != nil {
		body = testCase.body()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(
		ctx, testCase.method, "http://"+serverAddress+testCase.path, body)
	assert.NoError(t, err, assert.Continue())
	if err != nil {
		return
	}
	if testCase.contentType != "" {
		request.Header.Set(headers.ContentType, testCase.contentType)
	}
	response, err := http.DefaultClient.Do(request)
	assert.NoError(t, err, assert.Continue())
	if err != nil {
		return
	}
	assert.Equals(t, response.StatusCode, testCase.expected, assert.Continue())
	assert.NoError(t, response.Body.Close(), assert.Continue())
}

func concurrentRequestEndpoints() []endpoints.EndpointHandler {
	return []endpoints.EndpointHandler{
		&testHandler{
			Path:   "/status",
			Method: http.MethodGet,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				type params struct {
					Value string `json:"-" urlQuery:"value" validate:"required"`
				}
				responders.Status[params](writer, request, func(*params) (int, error) {
					return http.StatusOK, nil
				})
			},
		},
		&testHandler{
			Path:   "/error",
			Method: http.MethodGet,
			Handler: func(writer http.ResponseWriter, _ *http.Request) {
				responders.Error(writer, errors.New("error"))
			},
		},
		&testHandler{
			Path:   "/json/{id}",
			Method: http.MethodPost,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				type requestParams struct {
					ID   string `json:"-"    urlPath:"id"        validate:"required"`
					Data string `json:"data" validate:"required"`
				}
				type response struct {
					ID string
				}
				responders.JSON(writer, request, func(params *requestParams) (*response, int, error) {
					return &response{
						ID: params.ID,
					}, http.StatusOK, nil
				})
			},
		},
		&testHandler{
			Path:   "/jsonstream",
			Method: http.MethodGet,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				type requestParams struct{}
				type response struct {
					ID string
				}
				responders.JSONStream(writer, request, func(*requestParams) (<-chan *response, int, error) {
					responseChan := make(chan *response)
					go func() {
						defer close(responseChan)
						responseChan <- &response{ID: "1"}
						responseChan <- &response{ID: "2"}
						responseChan <- &response{ID: "3"}
					}()
					return responseChan, http.StatusOK, nil
				})
			},
		},
	}
}

func TestRun_ConcurrentRequests_NoErrors(t *testing.T) {
	t.Parallel()

	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}), server.WithEndpoints(concurrentRequestEndpoints()...))

	testCases := []concurrentTestCase{
		{http.MethodGet, "/status?value=test", nil, "", http.StatusOK},
		{http.MethodGet, "/status", nil, "", http.StatusBadRequest},
		{http.MethodGet, "/error", nil, "", http.StatusInternalServerError},
		{http.MethodPost, "/json/testId",
			func() io.Reader { return bytes.NewBufferString(`{"data":"value"}`) },
			headers.ContentTypeApplicationJSON, http.StatusOK},
		{http.MethodPost, "/json/testId",
			func() io.Reader { return bytes.NewBufferString(`{"data":""}`) },
			headers.ContentTypeApplicationJSON, http.StatusBadRequest},
		{http.MethodGet, "/jsonstream", nil, "", http.StatusOK},
	}

	var waitGroup sync.WaitGroup
	waitToStart := make(chan struct{})
	const totalGoRoutinesPerOperation = 2
	const totalRequestsPerGoRoutine = 1000

	for _, testCase := range testCases {
		for range totalGoRoutinesPerOperation {
			waitGroup.Go(func() {
				<-waitToStart
				for range totalRequestsPerGoRoutine {
					runConcurrentRequest(t, serverAddress, testCase)
				}
			})
		}
	}

	close(waitToStart)
	waitGroup.Wait()
}
