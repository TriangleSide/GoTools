package api

import (
	"fmt"
	"net/http"

	"github.com/TriangleSide/GoTools/pkg/http/middleware"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

// Method is a command used by a client to indicate the desired action to be performed
// on a specified resource within a server as part of the HTTP protocol.
type Method string

// Path specifies the particular resource on the server.
type Path string

// Handler encapsulates middleware and an HTTP handler for request processing.
type Handler struct {
	Middleware []middleware.Middleware
	Handler    http.HandlerFunc
}

// HTTPAPIBuilder is used in the HTTPEndpointHandler's visitor to set routes to handlers.
type HTTPAPIBuilder struct {
	handlers map[Path]map[Method]*Handler
}

// NewHTTPAPIBuilder allocates and sets default values in an HTTPAPIBuilder.
func NewHTTPAPIBuilder() *HTTPAPIBuilder {
	return &HTTPAPIBuilder{
		handlers: make(map[Path]map[Method]*Handler),
	}
}

// MustRegister assigns a Path and Method to a Handler. This function does validation to ensure
// duplicates are not registered. If the path and method is already registered, this function panics.
func (builder *HTTPAPIBuilder) MustRegister(path Path, method Method, handler *Handler) {
	if err := validation.Var(string(path), pathValidationTag); err != nil {
		panic(fmt.Errorf("api path %q is not correctly formatted: %w", path, err))
	}

	if err := validation.Var(string(method), "oneof=GET POST HEAD PUT PATCH DELETE CONNECT OPTIONS TRACE"); err != nil {
		panic(fmt.Errorf("http method %q is invalid: %w", method, err))
	}

	// The handler can be nil in cases like cors requests. The Go HTTP server needs the route
	// to exist to handle the request, but there is no handler needed for it.
	if handler == nil {
		handler = &Handler{}
	}

	if handler.Handler == nil {
		handler.Handler = func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusNotImplemented)
		}
	}

	methodToHandlerMap, pathAlreadyRegistered := builder.handlers[path]
	if !pathAlreadyRegistered {
		methodToHandlerMap = make(map[Method]*Handler)
		builder.handlers[path] = methodToHandlerMap
	}

	if _, methodAlreadyRegistered := methodToHandlerMap[method]; methodAlreadyRegistered {
		panic(fmt.Errorf("method %q already registered for path %q", method, path))
	}

	methodToHandlerMap[method] = handler
}

// Handlers returns a map of Path to Method to Handler.
func (builder *HTTPAPIBuilder) Handlers() map[Path]map[Method]*Handler {
	return builder.handlers
}

// The HTTPEndpointHandler interface is implemented by structs that handle HTTP calls.
type HTTPEndpointHandler interface {
	AcceptHTTPAPIBuilder(builder *HTTPAPIBuilder)
}
