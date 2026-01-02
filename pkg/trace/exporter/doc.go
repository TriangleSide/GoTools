/*
Package exporter defines the interface for exporting trace spans.

An Exporter receives completed spans and sends them to a tracing backend or
storage system. Implementations of the Exporter interface handle the specifics
of how and where span data is transmitted.

Use this package when you need to define a custom span export destination.
*/
package exporter
