// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

package api

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"intelligence/pkg/http/middleware"
	"intelligence/pkg/validation"
)

const (
	pathValidationTag = "api_path"
)

// init adds a validator for the Path.
func init() {
	isValidCharacters := regexp.MustCompile(`^[a-zA-Z0-9/{}]+$`).MatchString

	errMsgForValidation := func(value any) error {
		path, ok := value.(string)
		if !ok {
			return errors.New("path must be a string")
		}
		if len(path) == 0 {
			return errors.New("path cannot be empty")
		}
		if path == "/" {
			return nil
		}
		if !isValidCharacters(path) {
			return errors.New("path contains invalid characters")
		}
		if !strings.HasPrefix(path, "/") {
			return errors.New("path must start with '/'")
		}
		if strings.HasSuffix(path, "/") {
			return errors.New("path cannot end with '/'")
		}
		parts := strings.Split(path, "/")
		parameters := map[string]bool{}
		for i := 1; i < len(parts); i++ {
			part := parts[i]
			if part == "" {
				return errors.New("path parts cannot be empty")
			}
			if _, foundPart := parameters[part]; foundPart {
				return errors.New("path part must be unique")
			}
			parameters[part] = true
			if strings.Contains(part, "{") || strings.Contains(part, "}") {
				if !strings.HasPrefix(part, "{") || !strings.HasSuffix(part, "}") {
					return errors.New("path parameters must start with '{' and end with '}'")
				}
				if strings.Count(part, "{") != 1 || strings.Count(part, "}") != 1 {
					return errors.New("path parameters have only one '{' and '}'")
				}
				if part == "{}" {
					return errors.New("path parameters cannot be empty")
				}
			}
		}
		return nil
	}

	validation.RegisterValidation(pathValidationTag, func(field validator.FieldLevel) bool {
		return errMsgForValidation(field.Field().String()) == nil
	}, func(fieldErr validator.FieldError) string {
		return errMsgForValidation(fieldErr.Value()).Error()
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

	if handler == nil {
		panic(fmt.Sprintf("The handler for path %s and method %s is nil.", path, method))
	}

	if handler.Handler == nil {
		panic(fmt.Sprintf("The handler func for path %s and method %s is nil.", path, method))
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
