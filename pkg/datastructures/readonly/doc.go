/*
Package readonly provides immutable wrappers for Go collections.

Use this package when you need to expose maps to callers without allowing
modification. The package includes a builder for constructing read-only maps
and ensures immutability by preventing access to the underlying data after
construction.
*/
package readonly
