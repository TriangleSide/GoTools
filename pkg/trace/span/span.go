package span

import (
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TriangleSide/GoTools/pkg/trace/attribute"
	"github.com/TriangleSide/GoTools/pkg/trace/event"
	"github.com/TriangleSide/GoTools/pkg/trace/status"
)

const (
	// errorEventName is the name used for error events recorded on a span.
	errorEventName = "error"

	// errorMessageKey is the attribute key for the error message.
	errorMessageKey = "error.message"

	// errorTypeKey is the attribute key for the error type.
	errorTypeKey = "error.type"
)

// Span represents a unit of work with timing information and hierarchical structure.
type Span struct {
	spanID     string
	idCounter  *atomic.Uint64
	name       string
	traceID    string
	startTime  time.Time
	endTime    time.Time
	parent     *Span
	children   []*Span
	attributes []*attribute.Attribute
	events     []*event.Event
	statusCode status.Code
	opts       *spanOptions
	mu         sync.RWMutex
}

// New creates a new span with the given name, trace ID, and optional parent.
// If a parent is provided, the new span is added as a child of the parent.
func New(name string, traceID string, parent *Span, opts ...Option) *Span {
	var spanID string
	var idCounter *atomic.Uint64

	if parent == nil {
		idCounter = new(atomic.Uint64)
		spanID = "0"
	} else {
		idCounter = parent.idCounter
		spanID = strconv.FormatUint(idCounter.Add(1), 10)
	}

	span := &Span{
		spanID:     spanID,
		idCounter:  idCounter,
		name:       name,
		traceID:    traceID,
		startTime:  time.Now(),
		children:   make([]*Span, 0),
		attributes: make([]*attribute.Attribute, 0),
		events:     make([]*event.Event, 0),
		statusCode: status.Unset,
		opts:       configure(opts...),
	}

	if parent != nil {
		span.parent = parent
		parent.addChild(span)
	}

	return span
}

// End marks the span as complete by recording the end time.
// Only the first call has any effect; subsequent calls are ignored.
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.endTime.IsZero() {
		s.endTime = time.Now()
		if s.opts.endCallback != nil {
			s.opts.endCallback(s)
		}
	}
}

// Name returns the name of the span.
func (s *Span) Name() string {
	return s.name
}

// SpanID returns the unique identifier for this span within its hierarchy.
// This should NOT be used as a globally unique identifier.
// A unique identifier for a span is a combination of an application ID, a TraceID, and a SpanID.
func (s *Span) SpanID() string {
	return s.spanID
}

// TraceID returns the distributed trace ID of the span.
// This should NOT be used as a globally unique identifier.
// A unique identifier for a span is a combination of an application ID, a TraceID, and a SpanID.
func (s *Span) TraceID() string {
	return s.traceID
}

// StartTime returns when the span started.
func (s *Span) StartTime() time.Time {
	return s.startTime
}

// EndTime returns when the span ended.
func (s *Span) EndTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.endTime
}

// Duration returns the time elapsed between start and end.
func (s *Span) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.endTime.IsZero() {
		return time.Since(s.startTime)
	}
	return s.endTime.Sub(s.startTime)
}

// Parent returns the parent span, or nil if this is a root span.
func (s *Span) Parent() *Span {
	return s.parent
}

// Children returns a copy of the child spans.
func (s *Span) Children() []*Span {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Span, len(s.children))
	copy(result, s.children)
	return result
}

// SetAttributes sets attributes on the span.
func (s *Span) SetAttributes(attrs ...*attribute.Attribute) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attributes = append(s.attributes, attrs...)
}

// Attributes returns a copy of all attributes on the span.
func (s *Span) Attributes() []*attribute.Attribute {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*attribute.Attribute, len(s.attributes))
	copy(result, s.attributes)
	return result
}

// AddEvent adds an event to the span.
func (s *Span) AddEvent(e *event.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
}

// Events returns a copy of all events on the span.
func (s *Span) Events() []*event.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*event.Event, len(s.events))
	copy(result, s.events)
	return result
}

// SetStatusCode sets the status code on the span.
func (s *Span) SetStatusCode(code status.Code) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statusCode = code
}

// StatusCode returns the status code of the span.
func (s *Span) StatusCode() status.Code {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.statusCode
}

// RecordError records an error on the span by setting the status to Error and adding an error event.
func (s *Span) RecordError(err error) {
	if err == nil {
		return
	}
	s.SetStatusCode(status.Error)
	s.AddEvent(event.New(errorEventName,
		attribute.String(errorMessageKey, err.Error()),
		attribute.String(errorTypeKey, reflect.TypeOf(err).String()),
	))
}

// addChild adds a child span to this span.
func (s *Span) addChild(child *Span) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.children = append(s.children, child)
}
