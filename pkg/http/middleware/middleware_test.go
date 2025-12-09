package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/http/middleware"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestCreateChain_NilMiddlewareList_OnlyCallsHandler(t *testing.T) {
	t.Parallel()

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
	}
	mwChain := middleware.CreateChain(nil, handler)
	mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	assert.True(t, called)
}

func TestCreateChain_EmptyMiddlewareList_OnlyCallsHandler(t *testing.T) {
	t.Parallel()

	called := false
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
	}
	mwChain := middleware.CreateChain([]middleware.Middleware{}, handler)
	mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	assert.True(t, called)
}

func TestCreateChain_SingleMiddleware_InvokesMiddlewareAndHandler(t *testing.T) {
	t.Parallel()

	invocations := []string{}
	mwList := []middleware.Middleware{
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				invocations = append(invocations, "middleware")
				next(writer, request)
			}
		},
	}
	handler := func(w http.ResponseWriter, req *http.Request) {
		invocations = append(invocations, "handler")
	}
	mwChain := middleware.CreateChain(mwList, handler)
	mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equals(t, len(invocations), 2)
	assert.Equals(t, invocations[0], "middleware")
	assert.Equals(t, invocations[1], "handler")
}

func TestCreateChain_MultipleMiddlewares_InvokesInOrder(t *testing.T) {
	t.Parallel()

	invocations := []string{}
	mwList := []middleware.Middleware{
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				invocations = append(invocations, "first")
				next(writer, request)
			}
		},
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				invocations = append(invocations, "second")
				next(writer, request)
			}
		},
	}
	handler := func(w http.ResponseWriter, req *http.Request) {
		invocations = append(invocations, "handler")
	}
	mwChain := middleware.CreateChain(mwList, handler)
	mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equals(t, len(invocations), 3)
	assert.Equals(t, invocations[0], "first")
	assert.Equals(t, invocations[1], "second")
	assert.Equals(t, invocations[2], "handler")
}

func TestCreateChain_NilFinalHandler_Panics(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Equals(t, r, "the final handler cannot be nil")
	}()
	middleware.CreateChain(nil, nil)
}

func TestCreateChain_AllNilMiddlewareEntries_OnlyCallsHandler(t *testing.T) {
	t.Parallel()

	called := false
	mwList := []middleware.Middleware{nil, nil, nil}
	handler := func(w http.ResponseWriter, req *http.Request) {
		called = true
	}
	mwChain := middleware.CreateChain(mwList, handler)
	mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	assert.True(t, called)
}

func TestCreateChain_NilMiddlewareEntries_SkipsThem(t *testing.T) {
	t.Parallel()

	invocations := []string{}
	mwList := []middleware.Middleware{
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				invocations = append(invocations, "first")
				next(writer, request)
			}
		},
		nil,
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				invocations = append(invocations, "second")
				next(writer, request)
			}
		},
	}
	handler := func(w http.ResponseWriter, req *http.Request) {
		invocations = append(invocations, "handler")
	}
	mwChain := middleware.CreateChain(mwList, handler)
	mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equals(t, len(invocations), 3)
	assert.Equals(t, invocations[0], "first")
	assert.Equals(t, invocations[1], "second")
	assert.Equals(t, invocations[2], "handler")
}

func TestCreateChain_MiddlewareDoesNotCallNext_ShortCircuitsChain(t *testing.T) {
	t.Parallel()

	invocations := []string{}
	mwList := []middleware.Middleware{
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				invocations = append(invocations, "first")
				next(writer, request)
			}
		},
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				invocations = append(invocations, "second")
			}
		},
	}
	handler := func(w http.ResponseWriter, req *http.Request) {
		invocations = append(invocations, "handler")
	}
	mwChain := middleware.CreateChain(mwList, handler)
	mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equals(t, len(invocations), 2)
	assert.Equals(t, invocations[0], "first")
	assert.Equals(t, invocations[1], "second")
}
