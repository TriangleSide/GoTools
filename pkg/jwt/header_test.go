package jwt_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestJWTHeader(t *testing.T) {
	t.Run("it should encode and decode a header", func(t *testing.T) {
		original := jwt.Header{Algorithm: "HS256", Type: "JWT", KeyID: "1"}
		encoded, err := jwt.EncodeHeader(original)
		assert.NoError(t, err)
		assert.NotEquals(t, encoded, "")

		decoded, err := jwt.DecodeHeader(encoded)
		assert.NoError(t, err)
		assert.Equals(t, *decoded, original)
	})

	t.Run("when encoded string is invalid base64 it should return an error", func(t *testing.T) {
		decoded, err := jwt.DecodeHeader("!invalid-base64!")
		assert.ErrorPart(t, err, "base64 decode error")
		assert.Nil(t, decoded)
	})

	t.Run("when encoded string is not valid JSON it should return an error", func(t *testing.T) {
		invalid := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
		decoded, err := jwt.DecodeHeader(invalid)
		assert.ErrorPart(t, err, "json unmarshal error")
		assert.Nil(t, decoded)
	})

	t.Run("when encoded string contains json with wrong types it should return an error", func(t *testing.T) {
		invalid := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":123}`))
		decoded, err := jwt.DecodeHeader(invalid)
		assert.ErrorPart(t, err, "json unmarshal error")
		assert.Nil(t, decoded)
	})

	t.Run("when json marshal fails it should return an error", func(t *testing.T) {
		originalMarshal := jwt.MarshalFunc
		defer func() { jwt.MarshalFunc = originalMarshal }()

		jwt.MarshalFunc = func(v any) ([]byte, error) { return nil, errors.New("marshal fail") }
		encoded, err := jwt.EncodeHeader(jwt.Header{})
		assert.ErrorPart(t, err, "json marshal error")
		assert.Equals(t, encoded, "")
	})
}
