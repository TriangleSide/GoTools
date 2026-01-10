package exporter

import (
	"github.com/TriangleSide/go-toolkit/pkg/telemetry/trace/span"
)

// Exporter defines the interface for exporting spans to a tracing backend.
type Exporter interface {
	// Export sends a completed span to the tracing backend.
	// Implementations should manage their own context.
	// No error is returned because the trace package does not handle them.
	Export(s *span.Span)
}
