package trace

import (
	"context"

	"github.com/TriangleSide/GoTools/pkg/trace/exporter"
	"github.com/TriangleSide/GoTools/pkg/trace/span"
)

// ctxKey is the type used for storing the span in context.
type ctxKey struct{}

// traceIDCtxKey is the type used for storing the trace ID in context.
type traceIDCtxKey struct{}

// exporterCtxKey is the type used for storing the exporter in context.
type exporterCtxKey struct{}

var (
	// ctxKeyInstance is the context key for the span.
	ctxKeyInstance ctxKey

	// traceIDCtxKeyInstance is the context key for the trace ID.
	traceIDCtxKeyInstance traceIDCtxKey

	// exporterCtxKeyInstance is the context key for the exporter.
	exporterCtxKeyInstance exporterCtxKey
)

// Start creates a new span with the given name and adds it to the parent span found in the context.
// If an exporter is set in the context, it will be invoked when the span ends.
func Start(ctx context.Context, name string) (context.Context, *span.Span) {
	var opts []span.Option
	if exp := exporterFromContext(ctx); exp != nil {
		opts = append(opts, span.WithEndCallback(func(s *span.Span) {
			exp.Export(s)
		}))
	}
	s := span.New(name, traceIDFromContext(ctx), spanFromContext(ctx), opts...)
	return context.WithValue(ctx, ctxKeyInstance, s), s
}

// spanFromContext retrieves the current span from the context.
func spanFromContext(ctx context.Context) *span.Span {
	if s, ok := ctx.Value(ctxKeyInstance).(*span.Span); ok {
		return s
	}
	return nil
}

// SetTraceID stores a trace ID in the context for distributed tracing.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDCtxKeyInstance, traceID)
}

// traceIDFromContext retrieves the trace ID from the context.
func traceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDCtxKeyInstance).(string); ok {
		return id
	}
	return ""
}

// SetExporter stores an exporter in the context for span export.
func SetExporter(ctx context.Context, exp exporter.Exporter) context.Context {
	return context.WithValue(ctx, exporterCtxKeyInstance, exp)
}

// exporterFromContext retrieves the exporter from the context.
func exporterFromContext(ctx context.Context) exporter.Exporter {
	if exp, ok := ctx.Value(exporterCtxKeyInstance).(exporter.Exporter); ok {
		return exp
	}
	return nil
}
