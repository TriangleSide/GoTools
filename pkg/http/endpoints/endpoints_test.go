package endpoints_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/http/endpoints"
	"github.com/TriangleSide/go-toolkit/pkg/http/middleware"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
)

func TestHandlers_EmptyBuilder_ReturnsNothing(t *testing.T) {
	t.Parallel()
	builder := endpoints.NewBuilder()
	handlers := builder.API()
	assert.Equals(t, len(handlers), 0)
}

func TestMustRegister_InvalidMethod_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		builder := endpoints.NewBuilder()
		builder.MustRegister("/", "BAD_METHOD", &endpoints.Endpoint{
			Middleware: nil,
			Handler:    func(http.ResponseWriter, *http.Request) {},
		})
	}, "method")
}

func TestMustRegister_InvalidPath_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		builder := endpoints.NewBuilder()
		builder.MustRegister("/!@#$%/{}", http.MethodGet, &endpoints.Endpoint{
			Middleware: nil,
			Handler:    func(http.ResponseWriter, *http.Request) {},
		})
	}, "path contains invalid characters")
}

func TestMustRegister_DuplicatePathAndMethod_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		builder := endpoints.NewBuilder()
		builder.MustRegister("/", http.MethodGet, &endpoints.Endpoint{
			Middleware: nil,
			Handler:    func(http.ResponseWriter, *http.Request) {},
		})
		builder.MustRegister("/", http.MethodGet, &endpoints.Endpoint{
			Middleware: nil,
			Handler:    func(http.ResponseWriter, *http.Request) {},
		})
	}, "method \"GET\" already registered for path \"/\"")
}

func TestMustRegister_NilRoute_ReturnsNotImplemented(t *testing.T) {
	t.Parallel()
	const path = "/"

	builder := endpoints.NewBuilder()
	builder.MustRegister("/", http.MethodGet, nil)

	pathToMethodToRoute := builder.API()
	assert.NotNil(t, pathToMethodToRoute)

	methodToRoute, pathFound := pathToMethodToRoute[path]
	assert.True(t, pathFound)
	assert.NotNil(t, methodToRoute)

	route, methodFound := methodToRoute[http.MethodGet]
	assert.True(t, methodFound)
	assert.NotNil(t, route)
	assert.NotNil(t, route.Handler)
	assert.Nil(t, route.Middleware)

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	route.Handler.ServeHTTP(recorder, request)
	assert.Equals(t, recorder.Code, http.StatusNotImplemented)
}

func TestMustRegister_SinglePathAndMethod_IsRetrievable(t *testing.T) {
	t.Parallel()
	const path = "/"

	builder := endpoints.NewBuilder()
	builder.MustRegister(path, http.MethodGet, &endpoints.Endpoint{
		Middleware: nil,
		Handler: func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	})

	pathToMethodToRoute := builder.API()
	assert.NotNil(t, pathToMethodToRoute)

	methodToRoute, pathFound := pathToMethodToRoute[path]
	assert.True(t, pathFound)
	assert.NotNil(t, methodToRoute)

	route, methodFound := methodToRoute[http.MethodGet]
	assert.True(t, methodFound)
	assert.NotNil(t, route)
	assert.NotNil(t, route.Handler)
	assert.Nil(t, route.Middleware)

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	route.Handler.ServeHTTP(recorder, request)
	assert.Equals(t, recorder.Code, http.StatusOK)
}

func TestMustRegister_TwoMethodsSamePath_BothRetrievable(t *testing.T) {
	t.Parallel()
	const path = "/"

	builder := endpoints.NewBuilder()
	builder.MustRegister(path, http.MethodGet, &endpoints.Endpoint{
		Middleware: nil,
		Handler: func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	})
	builder.MustRegister(path, http.MethodPost, &endpoints.Endpoint{
		Middleware: nil,
		Handler: func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusAccepted)
		},
	})

	pathToMethodToRoute := builder.API()
	assert.NotNil(t, pathToMethodToRoute)

	methodToRoute, pathFound := pathToMethodToRoute[path]
	assert.True(t, pathFound)
	assert.NotNil(t, methodToRoute)

	getRoute, getMethodFound := methodToRoute[http.MethodGet]
	assert.True(t, getMethodFound)
	assert.NotNil(t, getRoute)
	assert.NotNil(t, getRoute.Handler)
	assert.Nil(t, getRoute.Middleware)

	postRoute, postMethodFound := methodToRoute[http.MethodPost]
	assert.True(t, postMethodFound)
	assert.NotNil(t, postRoute)
	assert.NotNil(t, postRoute.Handler)
	assert.Nil(t, postRoute.Middleware)

	getRequest, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	assert.NoError(t, err)
	getRecorder := httptest.NewRecorder()
	getRoute.Handler.ServeHTTP(getRecorder, getRequest)
	assert.Equals(t, getRecorder.Code, http.StatusOK)

	postRequest, err := http.NewRequestWithContext(t.Context(), http.MethodPost, path, nil)
	assert.NoError(t, err)
	postRecorder := httptest.NewRecorder()
	postRoute.Handler.ServeHTTP(postRecorder, postRequest)
	assert.Equals(t, postRecorder.Code, http.StatusAccepted)
}

func TestMustRegister_TwoPathsSameMethod_BothRetrievable(t *testing.T) {
	t.Parallel()

	builder := endpoints.NewBuilder()
	builder.MustRegister("/test1", http.MethodGet, &endpoints.Endpoint{
		Middleware: nil,
		Handler: func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	})
	builder.MustRegister("/test2", http.MethodGet, &endpoints.Endpoint{
		Middleware: nil,
		Handler: func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusAccepted)
		},
	})

	pathToMethodToRoute := builder.API()
	assert.NotNil(t, pathToMethodToRoute)

	methodToRoute1, pathFound1 := pathToMethodToRoute["/test1"]
	assert.True(t, pathFound1)
	assert.NotNil(t, methodToRoute1)

	getRoute1, getMethodFound1 := methodToRoute1[http.MethodGet]
	assert.True(t, getMethodFound1)
	assert.NotNil(t, getRoute1)
	assert.NotNil(t, getRoute1.Handler)
	assert.Nil(t, getRoute1.Middleware)

	methodToRoute2, pathFound2 := pathToMethodToRoute["/test2"]
	assert.True(t, pathFound2)
	assert.NotNil(t, methodToRoute2)

	getRoute2, getMethodFound2 := methodToRoute2[http.MethodGet]
	assert.True(t, getMethodFound2)
	assert.NotNil(t, getRoute2)
	assert.NotNil(t, getRoute2.Handler)
	assert.Nil(t, getRoute2.Middleware)

	getRequest1, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/test1", nil)
	assert.NoError(t, err)
	getRecorder1 := httptest.NewRecorder()
	getRoute1.Handler.ServeHTTP(getRecorder1, getRequest1)
	assert.Equals(t, getRecorder1.Code, http.StatusOK)

	getRequest2, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/test2", nil)
	assert.NoError(t, err)
	getRecorder2 := httptest.NewRecorder()
	getRoute2.Handler.ServeHTTP(getRecorder2, getRequest2)
	assert.Equals(t, getRecorder2.Code, http.StatusAccepted)
}

func TestMustRegister_RouteWithMiddleware_StoresMiddleware(t *testing.T) {
	t.Parallel()
	const path = "/"

	testMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}
	}

	builder := endpoints.NewBuilder()
	builder.MustRegister(path, http.MethodGet, &endpoints.Endpoint{
		Middleware: []middleware.Middleware{testMiddleware},
		Handler: func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	})

	pathToMethodToRoute := builder.API()
	assert.NotNil(t, pathToMethodToRoute)

	methodToRoute, pathFound := pathToMethodToRoute[path]
	assert.True(t, pathFound)

	route, methodFound := methodToRoute[http.MethodGet]
	assert.True(t, methodFound)
	assert.NotNil(t, route)
	assert.NotNil(t, route.Middleware)
	assert.Equals(t, len(route.Middleware), 1)
}

func TestMustRegister_RouteWithNilHandlerFunc_ReturnsNotImplemented(t *testing.T) {
	t.Parallel()
	const path = "/"

	builder := endpoints.NewBuilder()
	builder.MustRegister(path, http.MethodGet, &endpoints.Endpoint{
		Middleware: nil,
		Handler:    nil,
	})

	pathToMethodToRoute := builder.API()
	methodToRoute := pathToMethodToRoute[path]
	route := methodToRoute[http.MethodGet]

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	route.Handler.ServeHTTP(recorder, request)
	assert.Equals(t, recorder.Code, http.StatusNotImplemented)
}
