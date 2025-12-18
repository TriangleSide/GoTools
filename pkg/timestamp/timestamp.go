package timestamp

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Timestamp represents a point in time stored as an RFC 3339 string for universal compatibility.
// It is implemented as a struct with a private field rather than a type alias (type Timestamp time.Time)
// to enforce the invariant that all timestamps are normalized to UTC. This prevents callers from
// bypassing the constructor and creating timestamps in non-UTC timezones.
type Timestamp struct {
	time time.Time
}

// New creates a Timestamp from a time.Time value.
func New(t time.Time) Timestamp {
	return Timestamp{time: t.UTC()}
}

// Time returns the underlying time.Time value.
func (ts Timestamp) Time() time.Time {
	return ts.time
}

// IsZero reports whether the timestamp represents the zero time instant.
func (ts Timestamp) IsZero() bool {
	return ts.time.IsZero()
}

// String returns the RFC 3339 representation of the timestamp.
func (ts Timestamp) String() string {
	if ts.time.IsZero() {
		return ""
	}
	return ts.time.Format(time.RFC3339)
}

// MarshalJSON implements json.Marshaler using RFC 3339 format.
func (ts Timestamp) MarshalJSON() ([]byte, error) {
	if ts.time.IsZero() {
		return nil, errors.New("timestamp is zero while marshaling")
	}
	jsonBytes, err := json.Marshal(ts.time.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal timestamp: %w", err)
	}
	return jsonBytes, nil
}

// UnmarshalJSON implements json.Unmarshaler expecting RFC 3339 format.
func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("timestamp must be a string: %w", err)
	}
	if str == "" {
		return errors.New("timestamp cannot be empty")
	}
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return fmt.Errorf("invalid RFC 3339 timestamp: %w", err)
	}
	ts.time = t.UTC()
	return nil
}

// MarshalText implements encoding.TextMarshaler using RFC 3339 format.
func (ts Timestamp) MarshalText() ([]byte, error) {
	if ts.time.IsZero() {
		return nil, errors.New("timestamp is zero while marshaling")
	}
	return []byte(ts.time.Format(time.RFC3339)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler expecting RFC 3339 format.
func (ts *Timestamp) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return errors.New("timestamp cannot be empty")
	}
	t, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return fmt.Errorf("invalid RFC 3339 timestamp: %w", err)
	}
	ts.time = t.UTC()
	return nil
}
