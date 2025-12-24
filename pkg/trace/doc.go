/*
Package trace provides lightweight tracing with hierarchical spans for measuring
the duration of operations.

Spans represent units of work and can be organized in parent-child relationships.
Each span records its start and end times, allowing measurement of how long
operations take. Spans are propagated through context, enabling automatic
parent-child linking across function boundaries.

Use this package when you need to instrument code to understand timing and
call hierarchies without external dependencies.
*/
package trace
