package middleware

import (
	"errors"
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
func CreateChain(middlewares []Middleware, finalHandlerFunc http.HandlerFunc) http.HandlerFunc {
	if finalHandlerFunc == nil {
		panic(errors.New("final handler cannot be nil"))
	}
	if len(middlewares) == 0 {
		return finalHandlerFunc
	}
	lastHandler := finalHandlerFunc
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] != nil {
			lastHandler = middlewares[i](lastHandler)
		}
	}
	return lastHandler
}
