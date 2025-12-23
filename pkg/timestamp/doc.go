/*
Package timestamp provides a UTC-normalized timestamp type with RFC 3339 serialization.

Use this package when you need consistent time representation across systems. All timestamps
are automatically converted to UTC, eliminating timezone ambiguity in serialized data.
The type implements JSON and text marshaling interfaces for seamless integration with
APIs and configuration files.
*/
package timestamp
