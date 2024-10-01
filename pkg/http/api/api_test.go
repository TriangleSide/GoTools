package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/api"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

func TestHTTPApi(t *testing.T) {
	t.Parallel()

	t.Run("when Handlers() is called it should have nothing", func(t *testing.T) {
		t.Parallel()
		builder := api.NewHTTPAPIBuilder()
		handlers := builder.Handlers()
		assert.Equals(t, len(handlers), 0)
	})

	t.Run("when an invalid method is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			builder := api.NewHTTPAPIBuilder()
			builder.MustRegister("/", "BAD_METHOD", &api.Handler{
				Middleware: nil,
				Handler:    func(writer http.ResponseWriter, request *http.Request) {},
			})
		}, "method")
	})

	t.Run("when an invalid path is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			builder := api.NewHTTPAPIBuilder()
			builder.MustRegister("/!@#$%/{}", http.MethodGet, &api.Handler{
				Middleware: nil,
				Handler:    func(writer http.ResponseWriter, request *http.Request) {},
			})
		}, "path contains invalid characters")
	})

	t.Run("when register is called twice with a path and method it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			builder := api.NewHTTPAPIBuilder()
			builder.MustRegister("/", http.MethodGet, &api.Handler{
				Middleware: nil,
				Handler:    func(writer http.ResponseWriter, request *http.Request) {},
			})
			builder.MustRegister("/", http.MethodGet, &api.Handler{
				Middleware: nil,
				Handler:    func(writer http.ResponseWriter, request *http.Request) {},
			})
		}, "method 'GET' already registered for path '/'")
	})

	t.Run("when a nil handler is registered it should create a handler that returns the not implemented status", func(t *testing.T) {
		t.Parallel()
		const path = "/"

		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister("/", http.MethodGet, nil)

		pathToMethodToHandler := builder.Handlers()
		assert.NotNil(t, pathToMethodToHandler)

		methodToHandler, pathFound := pathToMethodToHandler[path]
		assert.True(t, pathFound)
		assert.NotNil(t, methodToHandler)

		handler, methodFound := methodToHandler[http.MethodGet]
		assert.True(t, methodFound)
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.Handler)
		assert.Nil(t, handler.Middleware)

		request, err := http.NewRequest(http.MethodGet, path, nil)
		assert.NoError(t, err)
		recorder := httptest.NewRecorder()
		handler.Handler.ServeHTTP(recorder, request)
		assert.Equals(t, recorder.Code, http.StatusNotImplemented)
	})

	t.Run("when a path of / with a method of GET is registered it should be present when calling Handlers()", func(t *testing.T) {
		t.Parallel()
		const path = "/"

		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister(path, http.MethodGet, &api.Handler{
			Middleware: nil,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
			},
		})

		pathToMethodToHandler := builder.Handlers()
		assert.NotNil(t, pathToMethodToHandler)

		methodToHandler, pathFound := pathToMethodToHandler[path]
		assert.True(t, pathFound)
		assert.NotNil(t, methodToHandler)

		handler, methodFound := methodToHandler[http.MethodGet]
		assert.True(t, methodFound)
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.Handler)
		assert.Nil(t, handler.Middleware)

		request, err := http.NewRequest(http.MethodGet, path, nil)
		assert.NoError(t, err)
		recorder := httptest.NewRecorder()
		handler.Handler.ServeHTTP(recorder, request)
		assert.Equals(t, recorder.Code, http.StatusOK)
	})

	t.Run("when two methods are registered for the same path it should be present when calling Handlers()", func(t *testing.T) {
		t.Parallel()
		const path = "/"

		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister(path, http.MethodGet, &api.Handler{
			Middleware: nil,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
			},
		})
		builder.MustRegister(path, http.MethodPost, &api.Handler{
			Middleware: nil,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusAccepted)
			},
		})

		pathToMethodToHandler := builder.Handlers()
		assert.NotNil(t, pathToMethodToHandler)

		methodToHandler, pathFound := pathToMethodToHandler[path]
		assert.True(t, pathFound)
		assert.NotNil(t, methodToHandler)

		getHandler, getMethodFound := methodToHandler[http.MethodGet]
		assert.True(t, getMethodFound)
		assert.NotNil(t, getHandler)
		assert.NotNil(t, getHandler.Handler)
		assert.Nil(t, getHandler.Middleware)

		postHandler, postMethodFound := methodToHandler[http.MethodPost]
		assert.True(t, postMethodFound)
		assert.NotNil(t, postHandler)
		assert.NotNil(t, postHandler.Handler)
		assert.Nil(t, postHandler.Middleware)

		getRequest, err := http.NewRequest(http.MethodGet, path, nil)
		assert.NoError(t, err)
		getRecorder := httptest.NewRecorder()
		getHandler.Handler.ServeHTTP(getRecorder, getRequest)
		assert.Equals(t, getRecorder.Code, http.StatusOK)

		postRequest, err := http.NewRequest(http.MethodPost, path, nil)
		assert.NoError(t, err)
		postRecorder := httptest.NewRecorder()
		postHandler.Handler.ServeHTTP(postRecorder, postRequest)
		assert.Equals(t, postRecorder.Code, http.StatusAccepted)
	})

	t.Run("when two paths are registered for the same method it should be present when calling Handlers()", func(t *testing.T) {
		t.Parallel()

		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister("/test1", http.MethodGet, &api.Handler{
			Middleware: nil,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
			},
		})
		builder.MustRegister("/test2", http.MethodGet, &api.Handler{
			Middleware: nil,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusAccepted)
			},
		})

		pathToMethodToHandler := builder.Handlers()
		assert.NotNil(t, pathToMethodToHandler)

		methodToHandler1, pathFound1 := pathToMethodToHandler["/test1"]
		assert.True(t, pathFound1)
		assert.NotNil(t, methodToHandler1)

		getHandler1, getMethodFound1 := methodToHandler1[http.MethodGet]
		assert.True(t, getMethodFound1)
		assert.NotNil(t, getHandler1)
		assert.NotNil(t, getHandler1.Handler)
		assert.Nil(t, getHandler1.Middleware)

		methodToHandler2, pathFound2 := pathToMethodToHandler["/test2"]
		assert.True(t, pathFound2)
		assert.NotNil(t, methodToHandler2)

		getHandler2, getMethodFound2 := methodToHandler2[http.MethodGet]
		assert.True(t, getMethodFound2)
		assert.NotNil(t, getHandler2)
		assert.NotNil(t, getHandler2.Handler)
		assert.Nil(t, getHandler2.Middleware)

		getRequest1, err := http.NewRequest(http.MethodGet, "/test1", nil)
		assert.NoError(t, err)
		getRecorder1 := httptest.NewRecorder()
		getHandler1.Handler.ServeHTTP(getRecorder1, getRequest1)
		assert.Equals(t, getRecorder1.Code, http.StatusOK)

		getRequest2, err := http.NewRequest(http.MethodGet, "/test2", nil)
		assert.NoError(t, err)
		getRecorder2 := httptest.NewRecorder()
		getHandler2.Handler.ServeHTTP(getRecorder2, getRequest2)
		assert.Equals(t, getRecorder2.Code, http.StatusAccepted)
	})

	t.Run("cases for path validation", func(t *testing.T) {
		t.Parallel()

		validationFunc := func(path string, expectedErrorMsg string) {
			errCheck := func(err error) {
				if expectedErrorMsg != "" {
					assert.ErrorPart(t, err, expectedErrorMsg)
				} else {
					assert.NoError(t, err)
				}
			}

			type testStructRef struct {
				Path string `validate:"api_path"`
			}
			errCheck(validation.Struct(&testStructRef{Path: path}))

			type testStructPtr struct {
				Path *string `validate:"api_path"`
			}
			errCheck(validation.Struct(&testStructPtr{Path: &path}))
		}

		validationFunc("/", "")
		validationFunc("/a/b/c/1/2/3", "")
		validationFunc("/a/{b}/c", "")
		validationFunc("", "path cannot be empty")
		validationFunc("/+", "path contains invalid characters")
		validationFunc(" /a", "path contains invalid characters")
		validationFunc("/a ", "path contains invalid characters")
		validationFunc("/a/", "path cannot end with '/'")
		validationFunc("a/b", "path must start with '/'")
		validationFunc("a", "path must start with '/'")
		validationFunc("/a//b", "path parts cannot be empty")
		validationFunc("//a", "path parts cannot be empty")
		validationFunc("/a/{b", "path parameters must start with '{' and end with '}'")
		validationFunc("/a/b}", "path parameters must start with '{' and end with '}'")
		validationFunc("/a/{{b}", "path parameters have only one '{' and '}'")
		validationFunc("/a/{b}}", "path parameters have only one '{' and '}'")
		validationFunc("/a/{}", "path parameters cannot be empty")
		validationFunc("/a/{b}/{b}", "path part must be unique")
		validationFunc("/a/a", "path part must be unique")
		validationFunc("/a/b/a", "path part must be unique")
	})

	t.Run("path validation is done on a reference field that it not a string it should return an error", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Path int `validate:"api_path"`
		}
		test := testStruct{
			Path: 1,
		}
		err := validation.Struct(&test)
		assert.ErrorPart(t, err, "path must be a string")
	})

	t.Run("when path validation is done on a pointer field that it not a string it should return an error", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Path *int `validate:"api_path"`
		}
		i := 0
		test := testStruct{
			Path: &i,
		}
		err := validation.Struct(&test)
		assert.ErrorPart(t, err, "path must be a string")
	})
}
