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
	"fmt"
	"net/http"

	"intelligence/pkg/http/middleware"
)

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
	methodToHandlerMap, pathAlreadyRegistered := builder.handlers[path]
	if !pathAlreadyRegistered {
		methodToHandlerMap = make(map[Method]*Handler)
		builder.handlers[path] = methodToHandlerMap
	}

	if _, methodAlreadyRegistered := methodToHandlerMap[method]; methodAlreadyRegistered {
		panic(fmt.Errorf("method '%s' already registered for path '%s'", method.String(), path.String()))
	}

	if handler == nil || handler.Handler == nil {
		panic(fmt.Errorf("handler for path %s and method %s is nil", path.String(), method.String()))
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
