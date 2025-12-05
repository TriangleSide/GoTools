package jwt_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestTimestamp(t *testing.T) {
	t.Parallel()

	t.Run("when timestamp is zero it should return empty string from String method", func(t *testing.T) {
		t.Parallel()
		var ts jwt.Timestamp
		assert.Equals(t, ts.String(), "")
	})

	t.Run("when timestamp is non-zero it should return RFC 3339 string from String method", func(t *testing.T) {
		t.Parallel()
		ts := jwt.NewTimestamp(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
		assert.Equals(t, ts.String(), "2024-06-01T12:00:00Z")
	})

	t.Run("when timestamp is zero it should return an error", func(t *testing.T) {
		t.Parallel()
		var ts jwt.Timestamp
		_, err := json.Marshal(ts)
		assert.ErrorPart(t, err, "timestamp is zero while marshaling")
	})

	t.Run("when timestamp is non-zero it should marshal to RFC 3339 string", func(t *testing.T) {
		t.Parallel()
		ts := jwt.NewTimestamp(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
		data, err := json.Marshal(ts)
		assert.NoError(t, err)
		assert.Equals(t, string(data), `"2024-06-01T12:00:00Z"`)
	})

	t.Run("when unmarshaling valid RFC 3339 string it should succeed", func(t *testing.T) {
		t.Parallel()
		var ts jwt.Timestamp
		err := json.Unmarshal([]byte(`"2024-06-01T12:00:00Z"`), &ts)
		assert.NoError(t, err)
		assert.Equals(t, ts.Time(), time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
	})

	t.Run("when unmarshaling empty string it should return error", func(t *testing.T) {
		t.Parallel()
		var ts jwt.Timestamp
		err := json.Unmarshal([]byte(`""`), &ts)
		assert.ErrorPart(t, err, "timestamp cannot be empty")
	})

	t.Run("when unmarshaling invalid RFC 3339 string it should return error", func(t *testing.T) {
		t.Parallel()
		var ts jwt.Timestamp
		err := json.Unmarshal([]byte(`"not-a-timestamp"`), &ts)
		assert.ErrorPart(t, err, "invalid RFC 3339 timestamp")
	})

	t.Run("when unmarshaling non-string value it should return error", func(t *testing.T) {
		t.Parallel()
		var ts jwt.Timestamp
		err := json.Unmarshal([]byte(`12345`), &ts)
		assert.ErrorPart(t, err, "timestamp must be a string")
	})
}
