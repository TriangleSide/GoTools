package endpoints

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

// Endpoint encapsulates middleware and an HTTP handler for request processing.
type Endpoint struct {
	Middleware []middleware.Middleware
	Handler    http.HandlerFunc
}

// Builder is used in the Registrar's visitor to set routes to handlers.
type Builder struct {
	handlers map[Path]map[Method]*Endpoint
}

// NewBuilder allocates and sets default values in a Builder.
func NewBuilder() *Builder {
	return &Builder{
		handlers: make(map[Path]map[Method]*Endpoint),
	}
}

// MustRegister assigns a Path and Method to a Endpoint. This function does validation to ensure
// duplicates are not registered. If the path and method is already registered, this function panics.
func (builder *Builder) MustRegister(path Path, method Method, route *Endpoint) {
	if err := validation.Var(string(path), pathValidationTag); err != nil {
		panic(fmt.Errorf("route path %q is not correctly formatted: %w", path, err))
	}

	if err := validation.Var(string(method), "oneof=GET POST HEAD PUT PATCH DELETE CONNECT OPTIONS TRACE"); err != nil {
		panic(fmt.Errorf("http method %q is invalid: %w", method, err))
	}

	// The route can be nil in cases like cors requests. The Go HTTP server needs the route
	// to exist to handle the request, but there is no handler needed for it.
	if route == nil {
		route = &Endpoint{}
	}

	if route.Handler == nil {
		route.Handler = func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusNotImplemented)
		}
	}

	methodToRouteMap, pathAlreadyRegistered := builder.handlers[path]
	if !pathAlreadyRegistered {
		methodToRouteMap = make(map[Method]*Endpoint)
		builder.handlers[path] = methodToRouteMap
	}

	if _, methodAlreadyRegistered := methodToRouteMap[method]; methodAlreadyRegistered {
		panic(fmt.Errorf("method %q already registered for path %q", method, path))
	}

	methodToRouteMap[method] = route
}

// API returns a map of Path to Method to Endpoint.
func (builder *Builder) API() map[Path]map[Method]*Endpoint {
	return builder.handlers
}

// Registrar is implemented by types that register HTTP routes.
type Registrar interface {
	RegisterEndpoints(builder *Builder)
}
