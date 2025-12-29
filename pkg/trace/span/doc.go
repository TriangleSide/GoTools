/*
Package span provides the Span type for representing units of work with timing
information and hierarchical structure.

A Span records when an operation started and ended, maintains parent-child
relationships with other spans, and can hold attributes and events. Spans are
thread-safe for concurrent access.

Use this package when you need to create and manipulate spans directly. For
context-based span propagation, use the parent trace package.
*/
package span
