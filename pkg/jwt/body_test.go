package jwt

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestJWTBody(t *testing.T) {
	t.Run("it should encode and decode a body", func(t *testing.T) {
		original := Body{Issuer: "iss", Subject: "sub", Audience: "aud", ExpiresAt: 1, NotBefore: 2, IssuedAt: 3, TokenID: "id"}
		encoded, err := encodeBody(original)
		assert.NoError(t, err)
		assert.NotEquals(t, encoded, "")

		decoded, err := decodeBody(encoded)
		assert.NoError(t, err)
		assert.Equals(t, *decoded, original)
	})

	t.Run("when encoded string is invalid base64 it should return an error", func(t *testing.T) {
		decoded, err := decodeBody("!invalid-base64!")
		assert.ErrorPart(t, err, "base64 decode error")
		assert.Nil(t, decoded)
	})

	t.Run("when encoded string is not valid JSON it should return an error", func(t *testing.T) {
		invalid := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
		decoded, err := decodeBody(invalid)
		assert.ErrorPart(t, err, "json unmarshal error")
		assert.Nil(t, decoded)
	})

	t.Run("when encoded string contains json with wrong types it should return an error", func(t *testing.T) {
		invalid := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":123}`))
		decoded, err := decodeBody(invalid)
		assert.ErrorPart(t, err, "json unmarshal error")
		assert.Nil(t, decoded)
	})

	t.Run("when json marshal fails it should return an error", func(t *testing.T) {
		originalMarshal := MarshalFunc
		defer func() { MarshalFunc = originalMarshal }()

		MarshalFunc = func(v any) ([]byte, error) { return nil, errors.New("marshal fail") }
		encoded, err := encodeBody(Body{})
		assert.ErrorPart(t, err, "json marshal error")
		assert.Equals(t, encoded, "")
	})
}
