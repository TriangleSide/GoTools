package trace

import (
	"context"

	"github.com/TriangleSide/GoTools/pkg/trace/span"
)

// ctxKey is the type used for storing the span in context.
type ctxKey struct{}

var (
	// ctxKeyInstance is the context key for the span.
	ctxKeyInstance ctxKey
)

// Start creates a new span with the given name and adds it to the parent span found in the context.
func Start(ctx context.Context, name string) (context.Context, *span.Span) {
	s := span.New(name, fromContext(ctx))
	return context.WithValue(ctx, ctxKeyInstance, s), s
}

// fromContext retrieves the current span from the context.
func fromContext(ctx context.Context) *span.Span {
	if s, ok := ctx.Value(ctxKeyInstance).(*span.Span); ok {
		return s
	}
	return nil
}
