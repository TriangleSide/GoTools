package tracer

import (
	"context"
	"runtime"
	"time"
)

// ctxKey is the type used for storing the span in context.
type ctxKey struct{}

var (
	// ctxKeyInstance is the context key for the span.
	ctxKeyInstance ctxKey
)

// callerName returns the name of the function at the given call stack depth.
func callerName() string {
	const callerSkip = 2
	pc, _, _, ok := runtime.Caller(callerSkip)
	if !ok {
		return ""
	}
	return runtime.FuncForPC(pc).Name()
}

// StartSpan creates a new span using the caller's function name and adds it to the parent span found in the context.
func StartSpan(ctx context.Context) (context.Context, *Span) {
	name := callerName()

	span := &Span{
		name:      name,
		startTime: time.Now(),
		children:  make([]*Span, 0),
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
