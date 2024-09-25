package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/middleware"
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
		if called == false {
			t.Fatalf("the handler should have been called but it was not")
		}
	})

	t.Run("when the middleware chain is created with an empty middleware list it should only call the handler", func(t *testing.T) {
		t.Parallel()

		called := false
		handler := func(w http.ResponseWriter, req *http.Request) {
			called = true
		}
		mwChain := middleware.CreateChain(make([]middleware.Middleware, 0), handler)
		mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		if called == false {
			t.Fatalf("the handler should have been called but it was not")
		}
	})

	t.Run("when the middleware chain is created it should invoke them in order", func(t *testing.T) {
		invocations := make([]string, 0)
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

		if len(invocations) != 3 {
			t.Fatalf("excepting 3 invocations")
		}
		if invocations[0] != "first" {
			t.Fatalf("the first mw should have been invoked but it was not")
		}
		if invocations[1] != "second" {
			t.Fatalf("the second mw should have been invoked but it was not")
		}
		if invocations[2] != "handler" {
			t.Fatalf("the handler should have been invoked but it was not")
		}
	})
}
