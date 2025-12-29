package trace

import (
	"context"

	"github.com/TriangleSide/GoTools/pkg/trace/span"
)

// ctxKey is the type used for storing the span in context.
type ctxKey struct{}

// traceIDCtxKey is the type used for storing the trace ID in context.
type traceIDCtxKey struct{}

var (
	// ctxKeyInstance is the context key for the span.
	ctxKeyInstance ctxKey

	// traceIDCtxKeyInstance is the context key for the trace ID.
	traceIDCtxKeyInstance traceIDCtxKey
)

// Start creates a new span with the given name and adds it to the parent span found in the context.
func Start(ctx context.Context, name string) (context.Context, *span.Span) {
	s := span.New(name, traceIDFromContext(ctx), fromContext(ctx))
	return context.WithValue(ctx, ctxKeyInstance, s), s
}

// fromContext retrieves the current span from the context.
func fromContext(ctx context.Context) *span.Span {
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
