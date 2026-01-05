package event

import (
	"time"

	"github.com/TriangleSide/GoTools/pkg/telemetry/trace/attribute"
)

// Event represents a timestamped occurrence within a span.
type Event struct {
	name       string
	timestamp  time.Time
	attributes []*attribute.Attribute
}

// New creates a new event with the given name and optional attributes.
func New(name string, attrs ...*attribute.Attribute) *Event {
	attrsCopy := make([]*attribute.Attribute, len(attrs))
	copy(attrsCopy, attrs)
	return &Event{
		name:       name,
		timestamp:  time.Now(),
		attributes: attrsCopy,
	}
}

// Name returns the event's name.
func (e *Event) Name() string {
	return e.name
}

// Timestamp returns when the event occurred.
func (e *Event) Timestamp() time.Time {
	return e.timestamp
}

// Attributes returns a copy of the event's attributes.
func (e *Event) Attributes() []*attribute.Attribute {
	result := make([]*attribute.Attribute, len(e.attributes))
	copy(result, e.attributes)
	return result
}
