package trace

import (
	"maps"
	"sync"
	"time"
)

// Span represents a unit of work with timing information and hierarchical structure.
type Span struct {
	name       string
	startTime  time.Time
	endTime    time.Time
	parent     *Span
	children   []*Span
	attributes map[string]any
	mu         sync.RWMutex
}

// addChild adds a child span to this span.
func (s *Span) addChild(child *Span) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.children = append(s.children, child)
}

// End marks the span as complete by recording the end time.
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endTime = time.Now()
}

// Name returns the name of the span.
func (s *Span) Name() string {
	return s.name
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

// SetAttribute sets a key-value pair on the span.
func (s *Span) SetAttribute(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attributes[key] = value
}

// Attribute returns the value associated with the given key and whether it exists.
func (s *Span) Attribute(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.attributes[key]
	return value, ok
}

// Attributes returns a copy of all attributes on the span.
func (s *Span) Attributes() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]any, len(s.attributes))
	maps.Copy(result, s.attributes)
	return result
}
