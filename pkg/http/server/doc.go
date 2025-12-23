/*
Package server provides a configurable HTTP server with TLS support.

Use this package to create and run an HTTP server that handles requests using
registered endpoint handlers. The server supports configuration through
environment variables and options, including timeouts, TLS modes, and keep-alive
settings.

The server supports three TLS modes: OFF for plain HTTP, TLS for server-side
TLS, and MUTUAL_TLS for client certificate verification. Endpoint handlers
are registered using the api.HTTPEndpointHandler interface, and common
middleware can be applied to all routes.

The server provides graceful shutdown capabilities and notifies when the
network listener is bound through an optional callback.
*/
package server
