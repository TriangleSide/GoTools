package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/http/middleware"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestHTTPMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("when the middleware chain is created with a nil middleware list it should only call the handler", func(t *testing.T) {
		t.Parallel()

		called := false
		handler := func(w http.ResponseWriter, req *http.Request) {
			called = true
		}
		mwChain := middleware.CreateChain(nil, handler)
		mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		assert.True(t, called)
	})

	t.Run("when the middleware chain is created with an empty middleware list it should only call the handler", func(t *testing.T) {
		t.Parallel()

		called := false
		handler := func(w http.ResponseWriter, req *http.Request) {
			called = true
		}
		mwChain := middleware.CreateChain([]middleware.Middleware{}, handler)
		mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		assert.True(t, called)
	})

	t.Run("when the middleware chain is created it should invoke them in order", func(t *testing.T) {
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
	})

	t.Run("when the final handler is nil it should panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Equals(t, r, "the final handler cannot be nil")
		}()
		middleware.CreateChain(nil, nil)
	})

	t.Run("when the middleware list contains nil entries it should skip them", func(t *testing.T) {
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
	})

	t.Run("when middleware does not call next it should short circuit the chain", func(t *testing.T) {
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
	})
}
