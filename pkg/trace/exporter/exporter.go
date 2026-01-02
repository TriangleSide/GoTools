package exporter

import (
	"context"

	"github.com/TriangleSide/GoTools/pkg/trace/span"
)

// Exporter defines the interface for exporting spans to a tracing backend.
type Exporter interface {
	// Export sends a completed span to the tracing backend.
	// No error is returned because the trace package does not handle them.
	Export(ctx context.Context, s *span.Span)
}
