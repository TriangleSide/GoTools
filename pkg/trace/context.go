package trace

import (
	"context"
	"time"

	"github.com/TriangleSide/GoTools/pkg/trace/attribute"
)

// ctxKey is the type used for storing the span in context.
type ctxKey struct{}

var (
	// ctxKeyInstance is the context key for the span.
	ctxKeyInstance ctxKey
)

// Start creates a new span with the given name and adds it to the parent span found in the context.
func Start(ctx context.Context, name string) (context.Context, *Span) {
	span := &Span{
		name:       name,
		startTime:  time.Now(),
		children:   make([]*Span, 0),
		attributes: make([]*attribute.Attribute, 0),
	}

	if parent, ok := ctx.Value(ctxKeyInstance).(*Span); ok {
		span.parent = parent
		parent.addChild(span)
	}

	return context.WithValue(ctx, ctxKeyInstance, span), span
}

// FromContext retrieves the current span from the context.
func FromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value(ctxKeyInstance).(*Span); ok {
		return span
	}
	return nil
}
