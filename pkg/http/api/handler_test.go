package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/http/api"
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
			Expect(builder.Handlers()).To(HaveLen(0))
		})

		When("a path of / with a method of GET is registered", func() {
			var (
				path   api.Path
				method api.Method
			)

			BeforeEach(func() {
				path = api.NewPath("/")
				method = api.NewMethod(http.MethodGet)

				builder.MustRegister(path, method, &api.Handler{
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

				handler, methodFound := methodToHandler[method]
				Expect(methodFound).To(BeTrue())
				Expect(handler).To(Not(BeNil()))
				Expect(handler.Handler).To(Not(BeNil()))
				Expect(handler.Middleware).To(BeNil())
			})
		})

		When("a path of / with methods of GET and POST is registered", func() {
			var (
				path       api.Path
				getMethod  api.Method
				postMethod api.Method
			)

			BeforeEach(func() {
				path = api.NewPath("/")
				getMethod = api.NewMethod(http.MethodGet)
				postMethod = api.NewMethod(http.MethodPost)

				builder.MustRegister(path, getMethod, &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
				builder.MustRegister(path, postMethod, &api.Handler{
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

				getHandler, getMethodFound := methodToHandler[getMethod]
				Expect(getMethodFound).To(BeTrue())
				Expect(getHandler).To(Not(BeNil()))
				Expect(getHandler.Handler).To(Not(BeNil()))
				Expect(getHandler.Middleware).To(BeNil())

				postHandler, postMethodFound := methodToHandler[postMethod]
				Expect(postMethodFound).To(BeTrue())
				Expect(postHandler).To(Not(BeNil()))
				Expect(postHandler.Handler).To(Not(BeNil()))
				Expect(postHandler.Middleware).To(BeNil())
			})
		})

		When("paths of /test1 and /test2 with a method of GET respectively are registered", func() {
			var (
				path1     api.Path
				path2     api.Path
				getMethod api.Method
			)

			BeforeEach(func() {
				path1 = api.NewPath("/test1")
				path2 = api.NewPath("/test2")
				getMethod = api.NewMethod(http.MethodGet)

				builder.MustRegister(path1, getMethod, &api.Handler{
					Middleware: nil,
					Handler:    func(writer http.ResponseWriter, request *http.Request) {},
				})
				builder.MustRegister(path2, getMethod, &api.Handler{
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

				get1Handler, get1MethodFound := path1MethodToHandler[getMethod]
				Expect(get1MethodFound).To(BeTrue())
				Expect(get1Handler).To(Not(BeNil()))
				Expect(get1Handler.Handler).To(Not(BeNil()))
				Expect(get1Handler.Middleware).To(BeNil())

				path2MethodToHandler, path2Found := pathToMethodToHandler[path2]
				Expect(path2Found).To(BeTrue())
				Expect(path2MethodToHandler).To(Not(BeNil()))

				get2Handler, get2MethodFound := path2MethodToHandler[getMethod]
				Expect(get2MethodFound).To(BeTrue())
				Expect(get2Handler).To(Not(BeNil()))
				Expect(get2Handler.Handler).To(Not(BeNil()))
				Expect(get2Handler.Middleware).To(BeNil())
			})
		})
	})
})
