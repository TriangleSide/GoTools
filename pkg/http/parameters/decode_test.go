package parameters_test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

type testJsonReadCloser struct {
	ReturnedError error
	Closed        bool
}

func (j *testJsonReadCloser) Read(p []byte) (int, error) {
	jsonData := `{"message": "generic json response"}`
	return copy(p, jsonData), nil
}

func (j *testJsonReadCloser) Close() error {
	j.Closed = true
	return j.ReturnedError
}

func TestDecodeHTTPParameters(t *testing.T) {
	t.Parallel()

	t.Run("when decoding a struct that fails the tag validation it should panic", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		assert.PanicPart(t, func() {
			_, _ = parameters.Decode[struct {
				Field string `urlQuery:"a*" json:"-"`
			}](request)
		}, "lookup key 'a*' must adhere to the naming convention")
	})

	t.Run("when json is sent with an unknown field it should fail to decode", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fieldThatDoesNotExist":"value"}`))
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		request.Header.Set(headers.ContentType, headers.ContentTypeApplicationJson)
		_, err = parameters.Decode[struct {
			Field string `json:"myJsonField"`
		}](request)
		assert.ErrorPart(t, err, `unknown field "fieldThatDoesNotExist"`)
	})

	t.Run("when json is not properly formatted it should fail to decode", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"myJsonField":"value"`))
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		request.Header.Set(headers.ContentType, headers.ContentTypeApplicationJson)
		_, err = parameters.Decode[struct {
			Field string `json:"myJsonField"`
		}](request)
		assert.ErrorPart(t, err, `failed to decode json body`)
	})

	t.Run("when there are multiple values for a query parameter it should fail to decode", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodGet, "/?TestQuery=value1&TestQuery=value2", nil)
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		_, err = parameters.Decode[struct {
			Field string `urlQuery:"TestQuery" json:"-"`
		}](request)
		assert.ErrorPart(t, err, `expecting one value for query parameter TestQuery`)
	})

	t.Run("when there is a query parameter field that can't be set it should fail to decode", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodGet, "/?TestQuery=NotAnInt", nil)
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		_, err = parameters.Decode[struct {
			Field int `urlQuery:"TestQuery" json:"-"`
		}](request)
		assert.ErrorPart(t, err, `failed to set value for query parameter TestQuery`)
	})

	t.Run("when there are multiple values for a header it should fail to decode", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		request.Header["TestHeader"] = []string{"value1", "value2"}
		_, err = parameters.Decode[struct {
			Field string `httpHeader:"TestHeader" json:"-"`
		}](request)
		assert.ErrorPart(t, err, `expecting one value for header parameter TestHeader`)
	})

	t.Run("when there is a header field that can't be set it should fail to decode", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		request.Header["TestHeader"] = []string{"NotAndInt"}
		_, err = parameters.Decode[struct {
			Field int `httpHeader:"TestHeader" json:"-"`
		}](request)
		assert.ErrorPart(t, err, `failed to set value for header parameter TestHeader`)
	})

	t.Run("when there is a path field that can't be set it should fail to decode", func(t *testing.T) {
		t.Parallel()
		var decodeErr error
		mux := http.NewServeMux()
		mux.HandleFunc("/{urlTestPath}", func(_ http.ResponseWriter, request *http.Request) {
			_, decodeErr = parameters.Decode[struct {
				Field int `urlPath:"urlTestPath" json:"-"`
			}](request)
		})
		server := &http.Server{Handler: mux}
		defer func() {
			err := server.Close()
			assert.NoError(t, err, assert.Continue())
		}()
		listener, err := net.Listen("tcp", "[::1]:0")
		assert.NoError(t, err)
		go func() { _ = server.Serve(listener) }()
		response, err := http.Get("http://" + listener.Addr().String() + "/NotAnInt")
		t.Cleanup(func() {
			assert.NoError(t, response.Body.Close())
		})
		assert.NoError(t, err)
		assert.ErrorPart(t, decodeErr, `failed to set value for path parameter urlTestPath`)
	})

	t.Run("when the validation fails it should fail to decode", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		_, err = parameters.Decode[struct {
			Field string `httpHeader:"TestHeader" json:"-" validate:"required"`
		}](request)
		assert.ErrorPart(t, err, `validation failed on field 'Field' with validator 'required'`)
	})

	t.Run("when the generic is not a struct it should panic", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		assert.PanicPart(t, func() {
			_, _ = parameters.Decode[string](request)
		}, "the generic must be a struct")
	})

	t.Run("when the generic is a struct pointer it should panic", func(t *testing.T) {
		t.Parallel()
		type parameterParams struct {
			Field string `httpHeader:"TestHeader" json:"-" validate:"required"`
		}
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoError(t, err)
		request = request.WithContext(context.Background())
		assert.PanicPart(t, func() {
			_, _ = parameters.Decode[*parameterParams](request)
		}, "the generic must be a struct")
	})

	t.Run("when the body fails to close it should still succeed", func(t *testing.T) {
		t.Parallel()
		request, err := http.NewRequest(http.MethodPost, "/", nil)
		assert.NoError(t, err)
		readCloser := &testJsonReadCloser{
			ReturnedError: errors.New("test error"),
		}
		request.Body = readCloser
		request.Header.Set(headers.ContentType, headers.ContentTypeApplicationJson)
		request = request.WithContext(context.Background())
		decoded, err := parameters.Decode[struct {
			Field string `json:"message"`
		}](request)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)
		assert.NotEquals(t, decoded.Field, "")
		assert.True(t, readCloser.Closed)
	})

	t.Run("when decoding a struct with many different fields it should succeed", func(t *testing.T) {
		t.Parallel()
		type embeddedStruct struct {
			HeaderEmbeddedField string `httpHeader:"HeaderEmbeddedField" json:"-" validate:"required"`
		}

		type internalStruct struct {
			SubField1 string `json:"SubField1" validate:"required"`
			SubField2 int    `json:"SubField2" validate:"required"`
		}

		type parameterFields struct {
			embeddedStruct

			QueryStringField string            `urlQuery:"QueryStringField" json:"-" validate:"required"`
			QueryIntField    int               `urlQuery:"QueryIntField" json:"-" validate:"required"`
			QueryFloatField  float64           `urlQuery:"QueryFloatField" json:"-" validate:"required"`
			QueryBoolField   bool              `urlQuery:"QueryBoolField" json:"-" validate:"required"`
			QueryStructField internalStruct    `urlQuery:"QueryStructField" json:"-" validate:"required"`
			QueryMapField    map[string]string `urlQuery:"QueryMapField" json:"-" validate:"required"`
			QueryListField   []string          `urlQuery:"QueryListField" json:"-" validate:"required"`
			QueryNotSet      string            `urlQuery:"QueryNotSet" json:"-"`

			QueryPtrStringField *string            `urlQuery:"QueryPtrStringField" json:"-" validate:"required"`
			QueryPtrIntField    *int               `urlQuery:"QueryPtrIntField" json:"-" validate:"required"`
			QueryPtrFloatField  *float64           `urlQuery:"QueryPtrFloatField" json:"-" validate:"required"`
			QueryPtrBoolField   *bool              `urlQuery:"QueryPtrBoolField" json:"-" validate:"required"`
			QueryPtrStructField *internalStruct    `urlQuery:"QueryPtrStructField" json:"-" validate:"required"`
			QueryPtrMapField    *map[string]string `urlQuery:"QueryPtrMapField" json:"-" validate:"required"`
			QueryPtrListField   *[]string          `urlQuery:"QueryPtrListField" json:"-" validate:"required"`

			HeaderStringField string            `httpHeader:"HeaderStringField" json:"-" validate:"required"`
			HeaderIntField    int               `httpHeader:"HeaderIntField" json:"-" validate:"required"`
			HeaderFloatField  float64           `httpHeader:"HeaderFloatField" json:"-" validate:"required"`
			HeaderBoolField   bool              `httpHeader:"HeaderBoolField" json:"-" validate:"required"`
			HeaderStructField internalStruct    `httpHeader:"HeaderStructField" json:"-" validate:"required"`
			HeaderMapField    map[string]string `httpHeader:"HeaderMapField" json:"-" validate:"required"`
			HeaderListField   []string          `httpHeader:"HeaderListField" json:"-" validate:"required"`
			HeaderNotSet      string            `httpHeader:"HeaderNotSet" json:"-"`

			HeaderPtrStringField *string            `httpHeader:"HeaderPtrStringField" json:"-" validate:"required"`
			HeaderPtrIntField    *int               `httpHeader:"HeaderPtrIntField" json:"-" validate:"required"`
			HeaderPtrFloatField  *float64           `httpHeader:"HeaderPtrFloatField" json:"-" validate:"required"`
			HeaderPtrBoolField   *bool              `httpHeader:"HeaderPtrBoolField" json:"-" validate:"required"`
			HeaderPtrStructField *internalStruct    `httpHeader:"HeaderPtrStructField" json:"-" validate:"required"`
			HeaderPtrMapField    *map[string]string `httpHeader:"HeaderPtrMapField" json:"-" validate:"required"`
			HeaderPtrListField   *[]string          `httpHeader:"HeaderPtrListField" json:"-" validate:"required"`

			PathStringField string `urlPath:"PathStringField" json:"-" validate:"required"`
			PathNotSet      string `urlPath:"PathNotSet" json:"-"`

			PathPtrStringField *string `urlPath:"PathPtrStringField" json:"-" validate:"required"`

			JSONStringField string            `json:"JSONStringField,omitempty" validate:"required"`
			JSONIntField    int               `json:"JSONIntField,omitempty" validate:"required"`
			JSONFloatField  float64           `json:"JSONFloatField,omitempty" validate:"required"`
			JSONBoolField   bool              `json:"JSONBoolField,omitempty" validate:"required"`
			JSONStructField internalStruct    `json:"JSONStructField,omitempty" validate:"required"`
			JSONMapField    map[string]string `json:"JSONMapField,omitempty" validate:"required"`
			JSONListField   []string          `json:"JSONListField,omitempty" validate:"required"`
			JSONNotSet      string            `json:"JSONNotSet,omitempty"`

			JSONPtrStringField *string            `json:"JSONPtrStringField" validate:"required"`
			JSONPtrIntField    *int               `json:"JSONPtrIntField" validate:"required"`
			JSONPtrFloatField  *float64           `json:"JSONPtrFloatField" validate:"required"`
			JSONPtrBoolField   *bool              `json:"JSONPtrBoolField" validate:"required"`
			JSONPtrStructField *internalStruct    `json:"JSONPtrStructField" validate:"required"`
			JSONPtrMapField    *map[string]string `json:"JSONPtrMapField" validate:"required"`
			JSONPtrListField   *[]string          `json:"JSONPtrListField" validate:"required"`
		}

		params := &parameterFields{}
		assert.Error(t, validation.Struct(params))

		mux := http.NewServeMux()
		mux.HandleFunc("/{PathStringField}/{PathPtrStringField}/{doesNoExistInTheStruct}", func(_ http.ResponseWriter, request *http.Request) {
			params, _ = parameters.Decode[parameterFields](request)
		})

		server := &http.Server{Handler: mux}
		defer func() {
			err := server.Close()
			assert.NoError(t, err, assert.Continue())
		}()
		listener, err := net.Listen("tcp", "[::1]:0")
		assert.NoError(t, err)
		go func() { _ = server.Serve(listener) }()

		clientPath := "/pathStringField/pathPtrStringField/doesNotExistInTheStruct"
		queryParams := "?" +
			"QueryParamDoesNotExistInTheStruct=value" +
			"&QueryStringField=value" +
			"&QueryIntField=123" +
			"&QueryFloatField=1.23" +
			"&QueryBoolField=true" +
			"&QueryStructField=" + url.QueryEscape(`{"SubField1":"subValue1","SubField2":2}`) +
			"&QueryMapField=" + url.QueryEscape(`{"key1":"value1","key2":"value2"}`) +
			"&QueryListField=" + url.QueryEscape(`["item1","item2"]`) +
			"&QueryPtrStringField=value" +
			"&QueryPtrIntField=123" +
			"&QueryPtrFloatField=1.23" +
			"&QueryPtrBoolField=true" +
			"&QueryPtrStructField=" + url.QueryEscape(`{"SubField1":"subValue1","SubField2":2}`) +
			"&QueryPtrMapField=" + url.QueryEscape(`{"key1":"value1","key2":"value2"}`) +
			"&QueryPtrListField=" + url.QueryEscape(`["item1","item2"]`)

		jsonBody := `{
				"JSONStringField": "value",
				"JSONIntField": 123,
				"JSONFloatField": 1.23,
				"JSONBoolField": true,
				"JSONStructField": {"SubField1": "subValue1", "SubField2": 2},
				"JSONMapField": {"key": "value"},
				"JSONListField": ["item1", "item2"],
				"JSONPtrStringField": "value",
				"JSONPtrIntField": 123,
				"JSONPtrFloatField": 1.23,
				"JSONPtrBoolField": true,
				"JSONPtrStructField": {"SubField1": "subValue1", "SubField2": 2},
				"JSONPtrMapField": {"key": "value"},
				"JSONPtrListField": ["item1", "item2"]
			}`

		request, err := http.NewRequest(http.MethodPost, "http://"+listener.Addr().String()+clientPath+queryParams, bytes.NewBufferString(jsonBody))
		assert.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("HeaderDoesNotExistInTheStruct", "value")
		request.Header.Set("HeaderEmbeddedField", "value")
		request.Header.Set("HeaderStringField", "value")
		request.Header.Set("HeaderIntField", "123")
		request.Header.Set("HeaderFloatField", "1.23")
		request.Header.Set("HeaderBoolField", "1")
		request.Header.Set("HeaderStructField", `{"SubField1": "subValue1", "SubField2": 2}`)
		request.Header.Set("HeaderMapField", `{"key": "value"}`)
		request.Header.Set("HeaderListField", `["item1","item2"]`)
		request.Header.Set("HeaderPtrStringField", "value")
		request.Header.Set("HeaderPtrIntField", "123")
		request.Header.Set("HeaderPtrFloatField", "1.23")
		request.Header.Set("HeaderPtrBoolField", "true")
		request.Header.Set("HeaderPtrStructField", `{"SubField1": "subValue1", "SubField2": 2}`)
		request.Header.Set("HeaderPtrMapField", `{"key": "value"}`)
		request.Header.Set("HeaderPtrListField", `["item1","item2"]`)

		client := &http.Client{}
		response, err := client.Do(request)
		t.Cleanup(func() {
			assert.NoError(t, response.Body.Close())
		})
		assert.NoError(t, err)
		assert.Equals(t, response.StatusCode, http.StatusOK)
		assert.NoError(t, validation.Struct(params))

		assert.Equals(t, params.QueryStringField, "value")
		assert.Equals(t, params.QueryIntField, 123)
		assert.Equals(t, params.QueryFloatField, 1.23)
		assert.True(t, params.QueryBoolField)
		assert.Equals(t, params.QueryStructField.SubField1, "subValue1")
		assert.Equals(t, params.QueryStructField.SubField2, 2)
		assert.Equals(t, params.QueryMapField["key1"], "value1")
		assert.Equals(t, params.QueryMapField["key2"], "value2")
		assert.Equals(t, params.QueryListField[0], "item1")
		assert.Equals(t, params.QueryListField[1], "item2")

		assert.Equals(t, *params.QueryPtrStringField, "value")
		assert.Equals(t, *params.QueryPtrIntField, 123)
		assert.Equals(t, *params.QueryPtrFloatField, 1.23)
		assert.True(t, *params.QueryPtrBoolField)
		assert.Equals(t, *params.QueryPtrStructField, internalStruct{SubField1: "subValue1", SubField2: 2})
		assert.Equals(t, (*params.QueryPtrMapField)["key1"], "value1")
		assert.Equals(t, (*params.QueryPtrMapField)["key2"], "value2")
		assert.Equals(t, (*params.QueryPtrListField)[0], "item1")
		assert.Equals(t, (*params.QueryPtrListField)[1], "item2")

		assert.Equals(t, params.HeaderEmbeddedField, "value")
		assert.Equals(t, params.HeaderStringField, "value")
		assert.Equals(t, params.HeaderIntField, 123)
		assert.Equals(t, params.HeaderFloatField, 1.23)
		assert.True(t, params.HeaderBoolField)
		assert.Equals(t, params.HeaderStructField, internalStruct{SubField1: "subValue1", SubField2: 2})
		assert.Equals(t, params.HeaderMapField["key"], "value")
		assert.Equals(t, params.HeaderListField[0], "item1")
		assert.Equals(t, params.HeaderListField[1], "item2")

		assert.Equals(t, *params.HeaderPtrStringField, "value")
		assert.Equals(t, *params.HeaderPtrIntField, 123)
		assert.Equals(t, *params.HeaderPtrFloatField, 1.23)
		assert.True(t, *params.HeaderPtrBoolField)
		assert.Equals(t, *params.HeaderPtrStructField, internalStruct{SubField1: "subValue1", SubField2: 2})
		assert.Equals(t, (*params.HeaderPtrMapField)["key"], "value")
		assert.Equals(t, (*params.HeaderPtrListField)[0], "item1")
		assert.Equals(t, (*params.HeaderPtrListField)[1], "item2")

		assert.Equals(t, params.PathStringField, "pathStringField")
		assert.Equals(t, *params.PathPtrStringField, "pathPtrStringField")

		assert.Equals(t, params.JSONStringField, "value")
		assert.Equals(t, params.JSONIntField, 123)
		assert.Equals(t, params.JSONFloatField, 1.23)
		assert.True(t, params.JSONBoolField)
		assert.Equals(t, params.JSONStructField, internalStruct{SubField1: "subValue1", SubField2: 2})
		assert.Equals(t, params.JSONMapField["key"], "value")
		assert.Equals(t, params.JSONListField[0], "item1")
		assert.Equals(t, params.JSONListField[1], "item2")

		assert.Equals(t, *params.JSONPtrStringField, "value")
		assert.Equals(t, *params.JSONPtrIntField, 123)
		assert.Equals(t, *params.JSONPtrFloatField, 1.23)
		assert.True(t, *params.JSONPtrBoolField)
		assert.Equals(t, *params.JSONPtrStructField, internalStruct{SubField1: "subValue1", SubField2: 2})
		assert.Equals(t, (*params.JSONPtrMapField)["key"], "value")
		assert.Equals(t, (*params.JSONPtrListField)[0], "item1")
		assert.Equals(t, (*params.JSONPtrListField)[1], "item2")
	})
}
