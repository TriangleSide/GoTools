/*
Package middleware provides HTTP middleware chaining for request processing.

Use this package to compose multiple middleware functions into a single handler
chain. Each middleware in the chain can perform actions before and after calling
the next handler, enabling cross-cutting concerns like logging, authentication,
or request modification.

Middleware functions wrap the next handler and control whether to proceed with
the chain by calling the next function. The CreateChain function assembles
middleware in order, with the first middleware in the slice being the outermost
wrapper.
*/
package middleware
