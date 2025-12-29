/*
Package trace provides context-based span propagation for tracing operations.

This package offers the Start function to create spans that are automatically
linked to parent spans found in the context. Spans represent units of work and
can be organized in parent-child relationships, enabling measurement of how long
operations take across function boundaries. The SetTraceID function enables
distributed tracing by propagating a trace identifier across service boundaries.

Use this package when you need to instrument code with context-propagated tracing.
For direct span manipulation, use the span subpackage.
*/
package trace
