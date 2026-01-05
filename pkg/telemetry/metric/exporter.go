package metric

import "context"

// Exporter defines the interface for exporting metric data points to a backend.
type Exporter interface {
	// Export sends a metric data point to the metrics backend.
	// Implementations should manage their own context.
	// No error is returned because the metric package does not handle them.
	Export(p *Point)
}

// exporterCtxKey is the type used for storing the exporter in context.
type exporterCtxKey struct{}

var (
	// exporterCtxKeyInstance is the context key for the exporter.
	exporterCtxKeyInstance exporterCtxKey
)

// SetExporter stores an exporter in the context for metric export.
func SetExporter(ctx context.Context, exp Exporter) context.Context {
	return context.WithValue(ctx, exporterCtxKeyInstance, exp)
}

// exporterFromContext retrieves the exporter from the context.
func exporterFromContext(ctx context.Context) Exporter {
	if exp, ok := ctx.Value(exporterCtxKeyInstance).(Exporter); ok {
		return exp
	}
	return nil
}
