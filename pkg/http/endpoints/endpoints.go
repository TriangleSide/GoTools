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

// Builder is used in the EndpointHandler's visitor to set paths to handlers.
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
func (builder *Builder) MustRegister(path Path, method Method, endpoint *Endpoint) {
	if err := validation.Var(string(path), pathValidationTag); err != nil {
		panic(fmt.Errorf("endpoint path %q is not correctly formatted: %w", path, err))
	}

	if err := validation.Var(string(method), "oneof=GET POST HEAD PUT PATCH DELETE CONNECT OPTIONS TRACE"); err != nil {
		panic(fmt.Errorf("http method %q is invalid: %w", method, err))
	}

	if endpoint == nil {
		endpoint = &Endpoint{}
	}

	if endpoint.Handler == nil {
		endpoint.Handler = func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusNotImplemented)
		}
	}

	methodToEndpointMap, pathAlreadyRegistered := builder.handlers[path]
	if !pathAlreadyRegistered {
		methodToEndpointMap = make(map[Method]*Endpoint)
		builder.handlers[path] = methodToEndpointMap
	}

	if _, methodAlreadyRegistered := methodToEndpointMap[method]; methodAlreadyRegistered {
		panic(fmt.Errorf("method %q already registered for path %q", method, path))
	}

	methodToEndpointMap[method] = endpoint
}

// API returns a map of Path to Method to Endpoint.
func (builder *Builder) API() map[Path]map[Method]*Endpoint {
	return builder.handlers
}

// EndpointHandler is implemented by types that register HTTP endpoints.
type EndpointHandler interface {
	RegisterEndpoints(builder *Builder)
}
