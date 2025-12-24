/*
Package readonly provides immutable wrappers for Go collections.

Use this package when you need to expose maps or slices to callers without
allowing modification. The package includes builders for constructing read-only
collections and ensures immutability by preventing access to the underlying data
after construction.
*/
package readonly
