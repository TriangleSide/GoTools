package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/http/api"
	"intelligence/pkg/validation"
)

var _ = Describe("handler", func() {
	When("an HTTP API builder is created", func() {
		var (
			builder *api.HTTPAPIBuilder
		)

		BeforeEach(func() {
			builder = api.NewHTTPAPIBuilder()
		})

		It("should have nothing when Handlers() is called", func() {
			Expect(builder.Handlers()).To(BeEmpty())
		})

		It("should panic when an invalid method is registered", func() {
			Expect(func() {
				builder.MustRegister("/", "BAD_METHOD", &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
			}).To(PanicWith(ContainSubstring("HTTP method 'BAD_METHOD' is invalid ")))
		})

		It("should panic when an invalid path is registered", func() {
			Expect(func() {
				builder.MustRegister("/!@#$%/{}", http.MethodGet, &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
			}).To(PanicWith(ContainSubstring("path contains invalid characters")))
		})

		It("should panic when a nil handler is registered", func() {
			Expect(func() {
				builder.MustRegister("/", http.MethodGet, nil)
			}).To(PanicWith(ContainSubstring("handler for path / and method GET is nil")))
		})

		It("should panic when a nil handler func is registered", func() {
			Expect(func() {
				builder.MustRegister("/", http.MethodGet, &api.Handler{
					Middleware: nil,
					Handler:    nil,
				})
			}).To(PanicWith(ContainSubstring("handler func for path / and method GET is nil")))
		})

		It("should panic when path and method is registered twice", func() {
			Expect(func() {
				for i := 0; i < 2; i++ {
					builder.MustRegister("/", http.MethodGet, &api.Handler{
						Middleware: nil,
						Handler:    func(writer http.ResponseWriter, request *http.Request) {},
					})
				}
			}).To(PanicWith(ContainSubstring("method 'GET' already registered for path '/'")))
		})

		When("a path of / with a method of GET is registered", func() {
			var (
				path api.Path
			)

			BeforeEach(func() {
				path = "/"
				builder.MustRegister(path, http.MethodGet, &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
			})

			It("should be present when calling Handlers()", func() {
				pathToMethodToHandler := builder.Handlers()
				Expect(pathToMethodToHandler).To(Not(BeNil()))

				methodToHandler, pathFound := pathToMethodToHandler[path]
				Expect(pathFound).To(BeTrue())
				Expect(methodToHandler).To(Not(BeNil()))

				handler, methodFound := methodToHandler[http.MethodGet]
				Expect(methodFound).To(BeTrue())
				Expect(handler).To(Not(BeNil()))
				Expect(handler.Handler).To(Not(BeNil()))
				Expect(handler.Middleware).To(BeNil())
			})
		})

		When("a path of / with methods of GET and POST is registered", func() {
			var (
				path api.Path
			)

			BeforeEach(func() {
				path = "/"
				builder.MustRegister(path, http.MethodGet, &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
				builder.MustRegister(path, http.MethodPost, &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
			})

			It("should be present when calling Handlers()", func() {
				pathToMethodToHandler := builder.Handlers()
				Expect(pathToMethodToHandler).To(Not(BeNil()))

				methodToHandler, pathFound := pathToMethodToHandler[path]
				Expect(pathFound).To(BeTrue())
				Expect(methodToHandler).To(Not(BeNil()))

				getHandler, getMethodFound := methodToHandler[http.MethodGet]
				Expect(getMethodFound).To(BeTrue())
				Expect(getHandler).To(Not(BeNil()))
				Expect(getHandler.Handler).To(Not(BeNil()))
				Expect(getHandler.Middleware).To(BeNil())

				postHandler, postMethodFound := methodToHandler[http.MethodPost]
				Expect(postMethodFound).To(BeTrue())
				Expect(postHandler).To(Not(BeNil()))
				Expect(postHandler.Handler).To(Not(BeNil()))
				Expect(postHandler.Middleware).To(BeNil())
			})
		})

		When("paths of /test1 and /test2 with a method of GET respectively are registered", func() {
			var (
				path1 api.Path
				path2 api.Path
			)

			BeforeEach(func() {
				path1 = "/test1"
				path2 = "/test2"

				builder.MustRegister(path1, http.MethodGet, &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
				builder.MustRegister(path2, http.MethodGet, &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
			})

			It("should be present when calling Handlers()", func() {
				pathToMethodToHandler := builder.Handlers()
				Expect(pathToMethodToHandler).To(Not(BeNil()))

				path1MethodToHandler, path1Found := pathToMethodToHandler[path1]
				Expect(path1Found).To(BeTrue())
				Expect(path1MethodToHandler).To(Not(BeNil()))

				get1Handler, get1MethodFound := path1MethodToHandler[http.MethodGet]
				Expect(get1MethodFound).To(BeTrue())
				Expect(get1Handler).To(Not(BeNil()))
				Expect(get1Handler.Handler).To(Not(BeNil()))
				Expect(get1Handler.Middleware).To(BeNil())

				path2MethodToHandler, path2Found := pathToMethodToHandler[path2]
				Expect(path2Found).To(BeTrue())
				Expect(path2MethodToHandler).To(Not(BeNil()))

				get2Handler, get2MethodFound := path2MethodToHandler[http.MethodGet]
				Expect(get2MethodFound).To(BeTrue())
				Expect(get2Handler).To(Not(BeNil()))
				Expect(get2Handler.Handler).To(Not(BeNil()))
				Expect(get2Handler.Middleware).To(BeNil())
			})
		})
	})

	DescribeTable("path validation",
		func(path string, expectedErrorMsg string) {
			errCheck := func(err error) {
				if expectedErrorMsg != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(expectedErrorMsg))
				} else {
					Expect(err).To(Succeed())
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
		},
		Entry("root path", "/", ""),
		Entry("sub paths", "/a/b/c/1/2/3", ""),
		Entry("sub paths with a param", "/a/{b}/c", ""),
		Entry("empty string", "", "path cannot be empty"),
		Entry("invalid characters", "/+", "path contains invalid characters"),
		Entry("invalid characters", " /a", "path contains invalid characters"),
		Entry("invalid characters", "/a ", "path contains invalid characters"),
		Entry("end with /", "/a/", "path cannot end with '/'"),
		Entry("start with /", "a/b", "path must start with '/'"),
		Entry("start with /", "a", "path must start with '/'"),
		Entry("empty path part", "/a//b", "path parts cannot be empty"),
		Entry("empty path part", "//a", "path parts cannot be empty"),
		Entry("no matching }", "/a/{b", "path parameters must start with '{' and end with '}'"),
		Entry("no matching {", "/a/b}", "path parameters must start with '{' and end with '}'"),
		Entry("many {", "/a/{{b}", "path parameters have only one '{' and '}'"),
		Entry("many }", "/a/{b}}", "path parameters have only one '{' and '}'"),
		Entry("empty param", "/a/{}", "path parameters cannot be empty"),
		Entry("reused param", "/a/{b}/{b}", "path part must be unique"),
		Entry("reused part", "/a/a", "path part must be unique"),
		Entry("reused part", "/a/b/a", "path part must be unique"),
	)

	When("path validation is done on a reference field that it not a string", func() {
		type testStruct struct {
			Path int `validate:"api_path"`
		}

		var (
			test testStruct
		)

		BeforeEach(func() {
			test = testStruct{
				Path: 1,
			}
		})

		It("should fail the validation", func() {
			err := validation.Struct(&test)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path must be a string"))
		})
	})

	When("path validation is done on a pointer field that it not a string", func() {
		type testStruct struct {
			Path *int `validate:"api_path"`
		}

		var (
			test testStruct
		)

		BeforeEach(func() {
			i := 0
			test = testStruct{
				Path: &i,
			}
		})

		It("should fail the validation", func() {
			err := validation.Struct(&test)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path must be a string"))
		})
	})
})
