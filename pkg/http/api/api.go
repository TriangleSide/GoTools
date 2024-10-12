package api

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/TriangleSide/GoBase/pkg/http/middleware"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

const (
	pathValidationTag = "api_path"
)

// init adds a validator for the Path.
func init() {
	isValidCharacters := regexp.MustCompile(`^[a-zA-Z0-9/{}]+$`).MatchString
	validation.MustRegisterValidator(pathValidationTag, func(params *validation.CallbackParameters) error {
		value := params.Value
		if validation.ValueIsNil(value) {
			return validation.NewViolation(pathValidationTag, params, "the path is a nil value")
		}
		validation.DereferenceValue(&value)
		if value.Kind() != reflect.String {
			return fmt.Errorf("path is of type %s but it must be a string or a ptr to a string", value.Kind().String())
		}
		path := value.String()
		if len(path) == 0 {
			return validation.NewViolation(pathValidationTag, params, "the path cannot be empty")
		}
		if path == "/" {
			return nil
		}
		if !isValidCharacters(path) {
			return validation.NewViolation(pathValidationTag, params, "the path contains invalid characters")
		}
		if !strings.HasPrefix(path, "/") {
			return validation.NewViolation(pathValidationTag, params, "the path must start with '/'")
		}
		if strings.HasSuffix(path, "/") {
			return validation.NewViolation(pathValidationTag, params, "the path cannot end with '/'")
		}
		parts := strings.Split(path, "/")
		parameters := map[string]bool{}
		for i := 1; i < len(parts); i++ {
			part := parts[i]
			if part == "" {
				return validation.NewViolation(pathValidationTag, params, "the path parts cannot be empty")
			}
			if _, foundPart := parameters[part]; foundPart {
				return validation.NewViolation(pathValidationTag, params, "the path parts must be unique")
			}
			parameters[part] = true
			if strings.Contains(part, "{") || strings.Contains(part, "}") {
				if !strings.HasPrefix(part, "{") || !strings.HasSuffix(part, "}") {
					return validation.NewViolation(pathValidationTag, params, "the path parameters must start with '{' and end with '}'")
				}
				if strings.Count(part, "{") != 1 || strings.Count(part, "}") != 1 {
					return validation.NewViolation(pathValidationTag, params, "the path parameters must have only one '{' and '}'")
				}
				if part == "{}" {
					return validation.NewViolation(pathValidationTag, params, "the path parameters cannot be empty")
				}
			}
		}
		return nil
	})
}

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
		panic(fmt.Sprintf("The API path '%s' is not correctly formatted (%s).", path, err.Error()))
	}

	if err := validation.Var(string(method), "oneof=GET POST HEAD PUT PATCH DELETE CONNECT OPTIONS TRACE"); err != nil {
		panic(fmt.Sprintf("HTTP method '%s' is invalid (%s).", method, err.Error()))
	}

	// The handler can be nil in cases like cors requests. The Go HTTP server needs the route
	// to exist to handle the request, but there is no handler needed for it.
	if handler == nil {
		handler = &Handler{}
	}

	if handler.Handler == nil {
		handler.Handler = func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotImplemented)
		}
	}

	methodToHandlerMap, pathAlreadyRegistered := builder.handlers[path]
	if !pathAlreadyRegistered {
		methodToHandlerMap = make(map[Method]*Handler)
		builder.handlers[path] = methodToHandlerMap
	}

	if _, methodAlreadyRegistered := methodToHandlerMap[method]; methodAlreadyRegistered {
		panic(fmt.Sprintf("method '%s' already registered for path '%s'", method, path))
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
