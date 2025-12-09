package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/http/api"
	"github.com/TriangleSide/GoTools/pkg/http/middleware"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestHandlers_EmptyBuilder_ReturnsNothing(t *testing.T) {
	t.Parallel()
	builder := api.NewHTTPAPIBuilder()
	handlers := builder.Handlers()
	assert.Equals(t, len(handlers), 0)
}

func TestMustRegister_InvalidMethod_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister("/", "BAD_METHOD", &api.Handler{
			Middleware: nil,
			Handler:    func(writer http.ResponseWriter, request *http.Request) {},
		})
	}, "method")
}

func TestMustRegister_InvalidPath_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		builder := api.NewHTTPAPIBuilder()
		builder.MustRegister("/!@#$%/{}", http.MethodGet, &api.Handler{
			Middleware: nil,
			Handler:    func(writer http.ResponseWriter, request *http.Request) {},
		})
	}, "path contains invalid characters")
}

func TestMustRegister_DuplicatePathAndMethod_Panics(t *testing.T) {
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
}

func TestMustRegister_NilHandler_ReturnsNotImplemented(t *testing.T) {
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

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	handler.Handler.ServeHTTP(recorder, request)
	assert.Equals(t, recorder.Code, http.StatusNotImplemented)
}

func TestMustRegister_SinglePathAndMethod_IsRetrievable(t *testing.T) {
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

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	handler.Handler.ServeHTTP(recorder, request)
	assert.Equals(t, recorder.Code, http.StatusOK)
}

func TestMustRegister_TwoMethodsSamePath_BothRetrievable(t *testing.T) {
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

	getRequest, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	assert.NoError(t, err)
	getRecorder := httptest.NewRecorder()
	getHandler.Handler.ServeHTTP(getRecorder, getRequest)
	assert.Equals(t, getRecorder.Code, http.StatusOK)

	postRequest, err := http.NewRequestWithContext(t.Context(), http.MethodPost, path, nil)
	assert.NoError(t, err)
	postRecorder := httptest.NewRecorder()
	postHandler.Handler.ServeHTTP(postRecorder, postRequest)
	assert.Equals(t, postRecorder.Code, http.StatusAccepted)
}

func TestMustRegister_TwoPathsSameMethod_BothRetrievable(t *testing.T) {
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

	getRequest1, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/test1", nil)
	assert.NoError(t, err)
	getRecorder1 := httptest.NewRecorder()
	getHandler1.Handler.ServeHTTP(getRecorder1, getRequest1)
	assert.Equals(t, getRecorder1.Code, http.StatusOK)

	getRequest2, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/test2", nil)
	assert.NoError(t, err)
	getRecorder2 := httptest.NewRecorder()
	getHandler2.Handler.ServeHTTP(getRecorder2, getRequest2)
	assert.Equals(t, getRecorder2.Code, http.StatusAccepted)
}

func TestPathValidation_VariousPaths_ValidatesCorrectly(t *testing.T) {
	t.Parallel()

	validationFunc := func(t *testing.T, path string, expectedErrorMsg string) {
		t.Helper()

		errCheck := func(t *testing.T, err error) {
			t.Helper()
			if expectedErrorMsg != "" {
				assert.ErrorPart(t, err, expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		}

		type testStructRef struct {
			Path string `validate:"api_path"`
		}
		errCheck(t, validation.Struct(&testStructRef{Path: path}))

		type testStructPtr struct {
			Path *string `validate:"api_path"`
		}
		errCheck(t, validation.Struct(&testStructPtr{Path: &path}))
	}

	validationFunc(t, "/", "")
	validationFunc(t, "/a/b/c/1/2/3", "")
	validationFunc(t, "/a/{b}/c", "")
	validationFunc(t, "", "path cannot be empty")
	validationFunc(t, "/+", "path contains invalid characters")
	validationFunc(t, " /a", "path contains invalid characters")
	validationFunc(t, "/a ", "path contains invalid characters")
	validationFunc(t, "/a/", "path cannot end with '/'")
	validationFunc(t, "a/b", "path must start with '/'")
	validationFunc(t, "a", "path must start with '/'")
	validationFunc(t, "/a//b", "path parts cannot be empty")
	validationFunc(t, "//a", "path parts cannot be empty")
	validationFunc(t, "/a/{b", "path parameters must start with '{' and end with '}'")
	validationFunc(t, "/a/b}", "path parameters must start with '{' and end with '}'")
	validationFunc(t, "/a/{{b}", "path parameters must have only one '{' and '}'")
	validationFunc(t, "/a/{b}}", "path parameters must have only one '{' and '}'")
	validationFunc(t, "/a/{}", "path parameters cannot be empty")
	validationFunc(t, "/a/{b}/{b}", "path parts must be unique")
	validationFunc(t, "/a/a", "path parts must be unique")
	validationFunc(t, "/a/b/a", "path parts must be unique")
}

func TestPathValidation_NonStringReferenceField_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path int `validate:"api_path"`
	}
	test := testStruct{
		Path: 1,
	}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "must be a string")
}

func TestPathValidation_NonStringPointerField_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path *int `validate:"api_path"`
	}
	i := 0
	test := testStruct{
		Path: &i,
	}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "must be a string")
}

func TestPathValidation_NilPointerString_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path *string `validate:"api_path"`
	}
	test := testStruct{
		Path: nil,
	}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "value is nil")
}

func TestMustRegister_HandlerWithMiddleware_StoresMiddleware(t *testing.T) {
	t.Parallel()
	const path = "/"

	testMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)
		}
	}

	builder := api.NewHTTPAPIBuilder()
	builder.MustRegister(path, http.MethodGet, &api.Handler{
		Middleware: []middleware.Middleware{testMiddleware},
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		},
	})

	pathToMethodToHandler := builder.Handlers()
	assert.NotNil(t, pathToMethodToHandler)

	methodToHandler, pathFound := pathToMethodToHandler[path]
	assert.True(t, pathFound)

	handler, methodFound := methodToHandler[http.MethodGet]
	assert.True(t, methodFound)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.Middleware)
	assert.Equals(t, len(handler.Middleware), 1)
}

func TestMustRegister_HandlerWithNilHandlerFunc_ReturnsNotImplemented(t *testing.T) {
	t.Parallel()
	const path = "/"

	builder := api.NewHTTPAPIBuilder()
	builder.MustRegister(path, http.MethodGet, &api.Handler{
		Middleware: nil,
		Handler:    nil,
	})

	pathToMethodToHandler := builder.Handlers()
	methodToHandler := pathToMethodToHandler[path]
	handler := methodToHandler[http.MethodGet]

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	handler.Handler.ServeHTTP(recorder, request)
	assert.Equals(t, recorder.Code, http.StatusNotImplemented)
}
