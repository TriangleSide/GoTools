package parameters_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/http/headers"
	"intelligence/pkg/http/parameters"
	"intelligence/pkg/validation"
)

var _ = Describe("decode HTTP request parameters", func() {
	When("a server is started", Ordered, func() {
		var (
			hostAndPort   string
			requestPrefix func(string) string
			handlerPath   func(string, string) string
			listener      net.Listener
			mux           *http.ServeMux
			server        *http.Server
		)

		BeforeAll(func() {
			var err error
			hostAndPort = "[::1]:13531"
			requestPrefix = func(path string) string {
				return "http://" + hostAndPort + path
			}
			handlerPath = func(method, path string) string {
				return method + " " + path
			}
			server = &http.Server{}
			listener, err = net.Listen("tcp", hostAndPort)
			mux = http.NewServeMux()
			Expect(err).To(Not(HaveOccurred()))
			go func() {
				_ = http.Serve(listener, mux)
			}()
		})

		AfterAll(func() {
			Expect(server.Shutdown(context.Background())).To(Succeed())
			_ = listener.Close()
		})

		It("should panic when decoding a struct that fails the tag validation", func() {
			request, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			Expect(func() {
				_, _ = parameters.Decode[struct {
					Field string `urlQuery:"a*" json:"-"`
				}](request)
			}).Should(Panic())
		})

		It("should fail to decode when json is sent with an unknown field", func() {
			request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fieldThatDoesNotExist":"value"}`))
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			request.Header.Set(headers.ContentType, headers.ContentTypeApplicationJson)
			_, err = parameters.Decode[struct {
				Field string `json:"myJsonField"`
			}](request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`unknown field "fieldThatDoesNotExist"`))
		})

		It("should fail to decode when json is not properly formatted", func() {
			request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"myJsonField":"value"`))
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			request.Header.Set(headers.ContentType, headers.ContentTypeApplicationJson)
			_, err = parameters.Decode[struct {
				Field string `json:"myJsonField"`
			}](request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`failed to decode json body`))
		})

		It("should fail to decode when there are multiple values for a query parameter", func() {
			request, err := http.NewRequest(http.MethodGet, "/?TestQuery=value1&TestQuery=value2", nil)
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			_, err = parameters.Decode[struct {
				Field string `urlQuery:"TestQuery" json:"-"`
			}](request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`expecting one value for query parameter TestQuery`))
		})

		It("should fail to decode when there is a query parameter field that can't be set", func() {
			request, err := http.NewRequest(http.MethodGet, "/?TestQuery=NotAnInt", nil)
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			_, err = parameters.Decode[struct {
				Field int `urlQuery:"TestQuery" json:"-"`
			}](request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`failed to set value for query parameter TestQuery`))
		})

		It("should fail to decode when there are multiple values for a header", func() {
			request, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			request.Header["TestHeader"] = []string{"value1", "value2"}
			_, err = parameters.Decode[struct {
				Field string `httpHeader:"TestHeader" json:"-"`
			}](request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`expecting one value for header parameter TestHeader`))
		})

		It("should fail to decode when there is a header field that can't be set", func() {
			request, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			request.Header["TestHeader"] = []string{"NotAndInt"}
			_, err = parameters.Decode[struct {
				Field int `httpHeader:"TestHeader" json:"-"`
			}](request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`failed to set value for header parameter TestHeader`))
		})

		It("should fail to decode when there is a path field that can't be set", func() {
			var decodeErr error
			mux.HandleFunc(handlerPath(http.MethodGet, "/{urlTestPath}"), func(_ http.ResponseWriter, request *http.Request) {
				_, decodeErr = parameters.Decode[struct {
					Field int `urlPath:"urlTestPath" json:"-"`
				}](request)
			})
			_, err := http.Get(requestPrefix("/NotAnInt"))
			Expect(err).ToNot(HaveOccurred())
			Expect(decodeErr).To(HaveOccurred())
			Expect(decodeErr.Error()).To(ContainSubstring(`failed to set value for path parameter urlTestPath`))
		})

		It("should fail when the validation fails", func() {
			request, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			_, err = parameters.Decode[struct {
				Field string `httpHeader:"TestHeader" json:"-" validate:"required"`
			}](request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`validation failed on field 'Field' with validator 'required'`))
		})

		It("should fail when the generic is not a struct", func() {
			request, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			Expect(func() {
				_, _ = parameters.Decode[string](request)
			}).Should(Panic())
		})

		It("should panic when the generic is a struct pointer", func() {
			type parameterParams struct {
				Field string `httpHeader:"TestHeader" json:"-" validate:"required"`
			}
			request, err := http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).NotTo(HaveOccurred())
			request = request.WithContext(context.Background())
			Expect(func() {
				_, _ = parameters.Decode[*parameterParams](request)
			}).Should(Panic())
		})

		It("should successfully decode a struct with many different fields", func() {
			type internalStruct struct {
				SubField1 string `json:"SubField1" validate:"required"`
				SubField2 int    `json:"SubField2" validate:"required"`
			}

			type parameterFields struct {
				QueryStringField string            `urlQuery:"QueryStringField" json:"-" validate:"required"`
				QueryIntField    int               `urlQuery:"QueryIntField" json:"-" validate:"required"`
				QueryFloatField  float64           `urlQuery:"QueryFloatField" json:"-" validate:"required"`
				QueryBoolField   bool              `urlQuery:"QueryBoolField" json:"-" validate:"required"`
				QueryStructField internalStruct    `urlQuery:"QueryStructField" json:"-" validate:"required"`
				QueryMapField    map[string]string `urlQuery:"QueryMapField" json:"-" validate:"required"`
				QueryListField   []string          `urlQuery:"QueryListField" json:"-" validate:"required"`

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

				HeaderPtrStringField *string            `httpHeader:"HeaderPtrStringField" json:"-" validate:"required"`
				HeaderPtrIntField    *int               `httpHeader:"HeaderPtrIntField" json:"-" validate:"required"`
				HeaderPtrFloatField  *float64           `httpHeader:"HeaderPtrFloatField" json:"-" validate:"required"`
				HeaderPtrBoolField   *bool              `httpHeader:"HeaderPtrBoolField" json:"-" validate:"required"`
				HeaderPtrStructField *internalStruct    `httpHeader:"HeaderPtrStructField" json:"-" validate:"required"`
				HeaderPtrMapField    *map[string]string `httpHeader:"HeaderPtrMapField" json:"-" validate:"required"`
				HeaderPtrListField   *[]string          `httpHeader:"HeaderPtrListField" json:"-" validate:"required"`

				PathStringField    string  `urlPath:"PathStringField" json:"-" validate:"required"`
				PathPtrStringField *string `urlPath:"PathPtrStringField" json:"-" validate:"required"`

				JSONStringField string            `json:"JSONStringField,omitempty"  validate:"required"`
				JSONIntField    int               `json:"JSONIntField,omitempty"  validate:"required"`
				JSONFloatField  float64           `json:"JSONFloatField,omitempty"  validate:"required"`
				JSONBoolField   bool              `json:"JSONBoolField,omitempty"  validate:"required"`
				JSONStructField internalStruct    `json:"JSONStructField,omitempty"  validate:"required"`
				JSONMapField    map[string]string `json:"JSONMapField,omitempty"  validate:"required"`
				JSONListField   []string          `json:"JSONListField,omitempty"  validate:"required"`

				JSONPtrStringField *string            `json:"JSONPtrStringField"  validate:"required"`
				JSONPtrIntField    *int               `json:"JSONPtrIntField"  validate:"required"`
				JSONPtrFloatField  *float64           `json:"JSONPtrFloatField"  validate:"required"`
				JSONPtrBoolField   *bool              `json:"JSONPtrBoolField"  validate:"required"`
				JSONPtrStructField *internalStruct    `json:"JSONPtrStructField"  validate:"required"`
				JSONPtrMapField    *map[string]string `json:"JSONPtrMapField"  validate:"required"`
				JSONPtrListField   *[]string          `json:"JSONPtrListField"  validate:"required"`
			}

			params := &parameterFields{}
			Expect(validation.Struct(params)).To(HaveOccurred())

			serverPath := "/{PathStringField}/{PathPtrStringField}"
			clientPath := "/pathStringField/pathPtrStringField"

			var decodeErr error
			mux.HandleFunc(handlerPath(http.MethodPost, serverPath), func(_ http.ResponseWriter, request *http.Request) {
				params, decodeErr = parameters.Decode[parameterFields](request)
			})

			queryParams := "?" +
				"QueryStringField=value" +
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

			request, err := http.NewRequest(http.MethodPost, requestPrefix(clientPath)+queryParams, bytes.NewBufferString(jsonBody))
			Expect(err).NotTo(HaveOccurred())
			request.Header.Set("Content-Type", "application/json")

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
			_, err = client.Do(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(decodeErr).To(Not(HaveOccurred()))
			Expect(validation.Struct(params)).To(Not(HaveOccurred()))

			Expect(params.QueryStringField).To(Equal("value"))
			Expect(params.QueryIntField).To(Equal(123))
			Expect(params.QueryFloatField).To(Equal(1.23))
			Expect(params.QueryBoolField).To(BeTrue())
			Expect(params.QueryStructField).To(Equal(internalStruct{SubField1: "subValue1", SubField2: 2}))
			Expect(params.QueryMapField).To(Equal(map[string]string{"key1": "value1", "key2": "value2"}))
			Expect(params.QueryListField).To(Equal([]string{"item1", "item2"}))

			Expect(*params.QueryPtrStringField).To(Equal("value"))
			Expect(*params.QueryPtrIntField).To(Equal(123))
			Expect(*params.QueryPtrFloatField).To(Equal(1.23))
			Expect(*params.QueryPtrBoolField).To(BeTrue())
			Expect(*params.QueryPtrStructField).To(Equal(internalStruct{SubField1: "subValue1", SubField2: 2}))
			Expect(*params.QueryPtrMapField).To(Equal(map[string]string{"key1": "value1", "key2": "value2"}))
			Expect(*params.QueryPtrListField).To(Equal([]string{"item1", "item2"}))

			Expect(params.HeaderStringField).To(Equal("value"))
			Expect(params.HeaderIntField).To(Equal(123))
			Expect(params.HeaderFloatField).To(Equal(1.23))
			Expect(params.HeaderBoolField).To(BeTrue())
			Expect(params.HeaderStructField).To(Equal(internalStruct{SubField1: "subValue1", SubField2: 2}))
			Expect(params.HeaderMapField).To(Equal(map[string]string{"key": "value"}))
			Expect(params.HeaderListField).To(Equal([]string{"item1", "item2"}))

			Expect(*params.HeaderPtrStringField).To(Equal("value"))
			Expect(*params.HeaderPtrIntField).To(Equal(123))
			Expect(*params.HeaderPtrFloatField).To(Equal(1.23))
			Expect(*params.HeaderPtrBoolField).To(BeTrue())
			Expect(*params.HeaderPtrStructField).To(Equal(internalStruct{SubField1: "subValue1", SubField2: 2}))
			Expect(*params.HeaderPtrMapField).To(Equal(map[string]string{"key": "value"}))
			Expect(*params.HeaderPtrListField).To(Equal([]string{"item1", "item2"}))

			Expect(params.PathStringField).To(Equal("pathStringField"))
			Expect(*params.PathPtrStringField).To(Equal("pathPtrStringField"))

			Expect(params.JSONStringField).To(Equal("value"))
			Expect(params.JSONIntField).To(Equal(123))
			Expect(params.JSONFloatField).To(Equal(1.23))
			Expect(params.JSONBoolField).To(BeTrue())
			Expect(params.JSONStructField).To(Equal(internalStruct{SubField1: "subValue1", SubField2: 2}))
			Expect(params.JSONMapField).To(Equal(map[string]string{"key": "value"}))
			Expect(params.JSONListField).To(Equal([]string{"item1", "item2"}))

			Expect(*params.JSONPtrStringField).To(Equal("value"))
			Expect(*params.JSONPtrIntField).To(Equal(123))
			Expect(*params.JSONPtrFloatField).To(Equal(1.23))
			Expect(*params.JSONPtrBoolField).To(BeTrue())
			Expect(*params.JSONPtrStructField).To(Equal(internalStruct{SubField1: "subValue1", SubField2: 2}))
			Expect(*params.JSONPtrMapField).To(Equal(map[string]string{"key": "value"}))
			Expect(*params.JSONPtrListField).To(Equal([]string{"item1", "item2"}))
		})
	})
})
