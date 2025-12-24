/*
Package config populates struct fields from registered configuration sources.

Use this package when you need to load configuration values into a struct from
external sources such as environment variables. Fields are marked for processing
using struct tags that specify the source type and optional default values.

The package includes a built-in environment variable source that maps struct
field names to SNAKE_CASE environment variable names. Custom sources can be
registered to support additional configuration backends.

Processed configurations can optionally be validated to ensure all required
values are present and conform to expected constraints.
*/
package config
