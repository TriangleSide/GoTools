package timestamp_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/timestamp"
)

func TestIsZero_WhenTimestampIsZero_ReturnsTrue(t *testing.T) {
	t.Parallel()
	var ts timestamp.Timestamp
	assert.Equals(t, ts.IsZero(), true)
}

func TestIsZero_WhenTimestampIsNonZero_ReturnsFalse(t *testing.T) {
	t.Parallel()
	ts := timestamp.New(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
	assert.Equals(t, ts.IsZero(), false)
}

func TestString_WhenTimestampIsZero_ReturnsEmptyString(t *testing.T) {
	t.Parallel()
	var ts timestamp.Timestamp
	assert.Equals(t, ts.String(), "")
}

func TestString_WhenTimestampIsNonZero_ReturnsRFC3339String(t *testing.T) {
	t.Parallel()
	ts := timestamp.New(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
	assert.Equals(t, ts.String(), "2024-06-01T12:00:00Z")
}

func TestMarshalJSON_WhenTimestampIsZero_ReturnsError(t *testing.T) {
	t.Parallel()
	var ts timestamp.Timestamp
	_, err := json.Marshal(ts)
	assert.ErrorPart(t, err, "timestamp is zero while marshaling")
}

func TestMarshalJSON_WhenTimestampIsNonZero_MarshalToRFC3339String(t *testing.T) {
	t.Parallel()
	ts := timestamp.New(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
	data, err := json.Marshal(ts)
	assert.NoError(t, err)
	assert.Equals(t, string(data), `"2024-06-01T12:00:00Z"`)
}

func TestUnmarshalJSON_ValidRFC3339String_Succeeds(t *testing.T) {
	t.Parallel()
	var ts timestamp.Timestamp
	err := json.Unmarshal([]byte(`"2024-06-01T12:00:00Z"`), &ts)
	assert.NoError(t, err)
	assert.Equals(t, ts.Time(), time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
}

func TestUnmarshalJSON_EmptyString_ReturnsError(t *testing.T) {
	t.Parallel()
	var ts timestamp.Timestamp
	err := json.Unmarshal([]byte(`""`), &ts)
	assert.ErrorPart(t, err, "timestamp cannot be empty")
}

func TestUnmarshalJSON_InvalidRFC3339String_ReturnsError(t *testing.T) {
	t.Parallel()
	var ts timestamp.Timestamp
	err := json.Unmarshal([]byte(`"not-a-timestamp"`), &ts)
	assert.ErrorPart(t, err, "invalid RFC 3339 timestamp")
}

func TestUnmarshalJSON_NonStringValue_ReturnsError(t *testing.T) {
	t.Parallel()
	var ts timestamp.Timestamp
	err := json.Unmarshal([]byte(`12345`), &ts)
	assert.ErrorPart(t, err, "timestamp must be a string")
}

func TestNew_NonUTCTimezone_ConvertsToUTC(t *testing.T) {
	t.Parallel()
	loc := time.FixedZone("UTC+5", 5*60*60)
	input := time.Date(2024, 6, 1, 17, 0, 0, 0, loc)
	ts := timestamp.New(input)
	assert.Equals(t, ts.Time(), time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
}

func TestUnmarshalJSON_TimezoneOffset_ConvertsToUTC(t *testing.T) {
	t.Parallel()
	var ts timestamp.Timestamp
	err := json.Unmarshal([]byte(`"2024-06-01T17:00:00+05:00"`), &ts)
	assert.NoError(t, err)
	assert.Equals(t, ts.Time(), time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
}

func TestTime_ReturnsUnderlyingTime(t *testing.T) {
	t.Parallel()
	expected := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	ts := timestamp.New(expected)
	assert.Equals(t, ts.Time(), expected)
}
