package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/middleware"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
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
}
