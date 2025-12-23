/*
Package parameters provides HTTP request parameter decoding and validation.

Use this package to populate struct fields from HTTP request data including
URL query parameters, HTTP headers, URL path parameters, and JSON request
bodies. The package automatically validates the decoded struct using the
validation package.

Struct fields are tagged to indicate their source:
  - urlQuery: extracts from URL query parameters
  - httpHeader: extracts from HTTP headers
  - urlPath: extracts from URL path parameters
  - json: extracts from JSON request body

Fields with urlQuery, httpHeader, or urlPath tags must also include a json:"-"
tag to prevent conflicts with JSON body parsing. The Decode function handles
all parameter sources in a single call and closes the request body when complete.
*/
package parameters
