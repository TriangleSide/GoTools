/*
Package structs provides utilities for inspecting and manipulating Go struct
types using reflection.

Use this package when you need to extract metadata from struct types, access
field values by name, or assign string-encoded values to struct fields. The
package handles embedded anonymous structs and supports various field types
including basic types, complex types (maps, slices, structs), and types
implementing encoding.TextUnmarshaler.

Field metadata includes the field's type, struct tags parsed into a map, and
the chain of anonymous struct names for embedded fields.
*/
package structs
