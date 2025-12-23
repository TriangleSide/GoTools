/*
Package responders provides HTTP response helpers for JSON APIs.

Use this package to handle the common patterns of decoding request parameters,
invoking a callback, and writing the response. The responders automatically
handle parameter validation, JSON encoding, and error responses.

The JSON responder encodes a response body as JSON. The JSONStream responder
streams multiple JSON objects using chunked transfer encoding. The Status
responder returns only an HTTP status code without a body. The Error responder
converts errors to appropriate HTTP status codes and JSON error messages.

Custom error types can be registered with MustRegisterErrorResponse to map
specific error types to HTTP status codes and formatted error messages.
Unregistered errors default to HTTP 500 Internal Server Error.
*/
package responders
