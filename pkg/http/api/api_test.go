package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/api"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

func TestHTTPApi(t *testing.T) {
	t.Parallel()

	t.Run("when Handlers() is called it should have nothing", func(t *testing.T) {
		t.Parallel()
		builder := api.NewHTTPAPIBuilder()
		handlers := builder.Handlers()
		if len(handlers) != 0 {
			t.Fatalf("handlers should be empty")
		}
	})

	t.Run("when an invalid method is registered it should panic", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("the code did not panic")
			} else {
				errorMsg := r.(string)
				if !strings.Contains(errorMsg, "method") {
					t.Errorf("the error message is not correct")
				}
			}
		}()
		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister("/", "BAD_METHOD", &api.Handler{
			Middleware: nil,
			Handler:    func(writer http.ResponseWriter, request *http.Request) {},
		})
	})

	t.Run("when an invalid path is registered it should panic", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("the code did not panic")
			} else {
				errorMsg := r.(string)
				if !strings.Contains(errorMsg, "path contains invalid characters") {
					t.Errorf("the error message is not correct")
				}
			}
		}()
		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister("/!@#$%/{}", http.MethodGet, &api.Handler{
			Middleware: nil,
			Handler:    func(writer http.ResponseWriter, request *http.Request) {},
		})
	})

	t.Run("when register is called twice with a path and method it should panic", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("the code did not panic")
			} else {
				errorMsg := r.(string)
				if !strings.Contains(errorMsg, "method 'GET' already registered for path '/'") {
					t.Errorf("the error message is not correct")
				}
			}
		}()
		builder := api.NewHTTPAPIBuilder()
		for i := 0; i < 2; i++ {
			builder.MustRegister("/", http.MethodGet, &api.Handler{
				Middleware: nil,
				Handler:    func(writer http.ResponseWriter, request *http.Request) {},
			})
		}
	})

	t.Run("when a nil handler is registered it should create a handler that returns the not implemented status", func(t *testing.T) {
		t.Parallel()
		const path = "/"

		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister("/", http.MethodGet, nil)

		pathToMethodToHandler := builder.Handlers()
		if pathToMethodToHandler == nil {
			t.Fatalf("handlers should not be nil")
		}

		methodToHandler, pathFound := pathToMethodToHandler[path]
		if !pathFound {
			t.Fatalf("handler should have been registered at path '%s'", path)
		}
		if methodToHandler == nil {
			t.Fatalf("methodToHandler should not be nil")
		}

		handler, methodFound := methodToHandler[http.MethodGet]
		if !methodFound {
			t.Fatalf("method get should have been registered at path '%s'", path)
		}
		if handler == nil {
			t.Fatalf("handler wrapper should not be nil")
		}
		if handler.Handler == nil {
			t.Fatalf("handler should not be nil")
		}
		if handler.Middleware != nil {
			t.Fatalf("middleware should be nil")
		}

		request, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			t.Fatalf("failed to create request (%s)", err.Error())
		}
		recorder := httptest.NewRecorder()
		handler.Handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotImplemented {
			t.Fatalf("handler should have returned a 501 status code")
		}
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
		if pathToMethodToHandler == nil {
			t.Fatalf("pathToMethodToHandler should not be nil")
		}

		methodToHandler, pathFound := pathToMethodToHandler[path]
		if !pathFound {
			t.Fatalf("handler should have been registered at path '%s'", path)
		}
		if methodToHandler == nil {
			t.Fatalf("methodToHandler should not be nil")
		}

		handler, methodFound := methodToHandler[http.MethodGet]
		if !methodFound {
			t.Fatalf("method get should have been registered at path '%s'", path)
		}
		if handler == nil {
			t.Fatalf("handler wrapper should not be nil")
		}
		if handler.Handler == nil {
			t.Fatalf("handler should not be nil")
		}
		if handler.Middleware != nil {
			t.Fatalf("middleware should be nil")
		}

		request, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			t.Fatalf("failed to create request (%s)", err.Error())
		}
		recorder := httptest.NewRecorder()
		handler.Handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("handler should have returned a 200 status code")
		}
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
		if pathToMethodToHandler == nil {
			t.Fatalf("pathToMethodToHandler should not be nil")
		}

		methodToHandler, pathFound := pathToMethodToHandler[path]
		if !pathFound {
			t.Fatalf("handler should have been registered at path '%s'", path)
		}
		if methodToHandler == nil {
			t.Fatalf("methodToHandler should not be nil")
		}

		getHandler, getMethodFound := methodToHandler[http.MethodGet]
		if !getMethodFound {
			t.Fatalf("method get should have been registered at path '%s'", path)
		}
		if getHandler == nil {
			t.Fatalf("handler wrapper should not be nil")
		}
		if getHandler.Handler == nil {
			t.Fatalf("handler should not be nil")
		}
		if getHandler.Middleware != nil {
			t.Fatalf("middleware should be nil")
		}

		postHandler, postMethodFound := methodToHandler[http.MethodPost]
		if !postMethodFound {
			t.Fatalf("method get should have been registered at path '%s'", path)
		}
		if postHandler == nil {
			t.Fatalf("handler wrapper should not be nil")
		}
		if postHandler.Handler == nil {
			t.Fatalf("handler should not be nil")
		}
		if postHandler.Middleware != nil {
			t.Fatalf("middleware should be nil")
		}

		getRequest, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			t.Fatalf("failed to create request (%s)", err.Error())
		}
		getRecorder := httptest.NewRecorder()
		getHandler.Handler.ServeHTTP(getRecorder, getRequest)
		if getRecorder.Code != http.StatusOK {
			t.Fatalf("handler should have returned a 200 status code")
		}

		postRequest, err := http.NewRequest(http.MethodPost, path, nil)
		if err != nil {
			t.Fatalf("failed to create request (%s)", err.Error())
		}
		postRecorder := httptest.NewRecorder()
		postHandler.Handler.ServeHTTP(postRecorder, postRequest)
		if postRecorder.Code != http.StatusAccepted {
			t.Fatalf("handler should have returned a 201 status code")
		}
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
		if pathToMethodToHandler == nil {
			t.Fatalf("pathToMethodToHandler should not be nil")
		}

		methodToHandler1, pathFound1 := pathToMethodToHandler["/test1"]
		if !pathFound1 {
			t.Fatalf("handler should have been registered at path /test1")
		}
		if methodToHandler1 == nil {
			t.Fatalf("methodToHandler should not be nil")
		}

		getHandler1, getMethodFound1 := methodToHandler1[http.MethodGet]
		if !getMethodFound1 {
			t.Fatalf("method get should have been registered at path /test1")
		}
		if getHandler1 == nil {
			t.Fatalf("handler wrapper should not be nil")
		}
		if getHandler1.Handler == nil {
			t.Fatalf("handler should not be nil")
		}
		if getHandler1.Middleware != nil {
			t.Fatalf("middleware should be nil")
		}

		methodToHandler2, pathFound2 := pathToMethodToHandler["/test2"]
		if !pathFound2 {
			t.Fatalf("handler should have been registered at path /test2")
		}
		if methodToHandler2 == nil {
			t.Fatalf("methodToHandler should not be nil")
		}

		getHandler2, getMethodFound2 := methodToHandler2[http.MethodGet]
		if !getMethodFound2 {
			t.Fatalf("method get should have been registered at path /test2")
		}
		if getHandler2 == nil {
			t.Fatalf("handler wrapper should not be nil")
		}
		if getHandler2.Handler == nil {
			t.Fatalf("handler should not be nil")
		}
		if getHandler2.Middleware != nil {
			t.Fatalf("middleware should be nil")
		}

		getRequest1, err := http.NewRequest(http.MethodGet, "/test1", nil)
		if err != nil {
			t.Fatalf("failed to create request (%s)", err.Error())
		}
		getRecorder1 := httptest.NewRecorder()
		getHandler1.Handler.ServeHTTP(getRecorder1, getRequest1)
		if getRecorder1.Code != http.StatusOK {
			t.Fatalf("handler should have returned a 200 status code")
		}

		getRequest2, err := http.NewRequest(http.MethodGet, "/test2", nil)
		if err != nil {
			t.Fatalf("failed to create request (%s)", err.Error())
		}
		getRecorder2 := httptest.NewRecorder()
		getHandler2.Handler.ServeHTTP(getRecorder2, getRequest2)
		if getRecorder2.Code != http.StatusAccepted {
			t.Fatalf("handler should have returned a 201 status code")
		}
	})

	t.Run("cases for path validation", func(t *testing.T) {
		t.Parallel()

		validationFunc := func(path string, expectedErrorMsg string) {
			errCheck := func(err error) {
				if expectedErrorMsg != "" {
					if err == nil {
						t.Errorf("error should not be nil")
					}
					if !strings.Contains(err.Error(), expectedErrorMsg) {
						t.Fatalf("error should contain '%s'", expectedErrorMsg)
					}
				} else {
					if err != nil {
						t.Fatalf("error should be nil")
					}
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
		if err == nil {
			t.Fatalf("validation should have returned an error")
		}
		if !strings.Contains(err.Error(), "path must be a string") {
			t.Fatalf("error message is not correct")
		}
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
		if err == nil {
			t.Fatalf("validation should have returned an error")
		}
		if !strings.Contains(err.Error(), "path must be a string") {
			t.Fatalf("error message is not correct")
		}
	})
}
