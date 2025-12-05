package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Timestamp represents a point in time stored as an RFC 3339 string for universal compatibility.
type Timestamp struct {
	time time.Time
}

// NewTimestamp creates a Timestamp from a time.Time value.
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{time: t.UTC()}
}

// Time returns the underlying time.Time value.
func (ts Timestamp) Time() time.Time {
	return ts.time
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
		return []byte("null"), nil
	}
	return json.Marshal(ts.time.Format(time.RFC3339))
}

// UnmarshalJSON implements json.Unmarshaler expecting RFC 3339 format.
func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("timestamp must be a string (%w)", err)
	}
	if s == "" {
		return errors.New("timestamp cannot be empty")
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("invalid RFC 3339 timestamp (%w)", err)
	}
	ts.time = t.UTC()
	return nil
}
