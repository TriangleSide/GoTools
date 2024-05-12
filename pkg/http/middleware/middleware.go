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

package middleware

import (
	"net/http"
)

// Middleware functions are invoked each time the server handles a route. The chain of middleware is followed until the
// final handler is reached. The middleware must call next(w, r) if the middleware does not handle the request.
//
// For example:
//
//	middleware := func(next http.HandlerFunc) http.HandlerFunc {
//	    return func(writer http.ResponseWriter, request *http.Request) {
//	        // Do middleware actions here.
//	        // Calling next to invoke the next middleware or request handler.
//	        next(writer, request)
//	    }
//	}
type Middleware func(next http.HandlerFunc) http.HandlerFunc

// CreateChain returns a http.HandlerFunc that invokes each middleware in order then the final http.HandlerFunc.
func CreateChain(mw []Middleware, finalHandlerFunc http.HandlerFunc) http.HandlerFunc {
	if len(mw) == 0 {
		return finalHandlerFunc
	}
	lastHandler := finalHandlerFunc
	for i := int(len(mw)) - 1; i >= 0; i-- {
		lastHandler = mw[i](lastHandler)
	}
	return lastHandler
}
