package jwt_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestStableJSONMarshalling(t *testing.T) {
	t.Parallel()

	t.Run("when the stable json marshaller is used it should encode fields in sorted order", func(t *testing.T) {
		t.Parallel()

		exp := jwt.NewTimestamp(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))
		nbf := jwt.NewTimestamp(time.Date(2021, 3, 20, 8, 30, 0, 0, time.UTC))
		iat := jwt.NewTimestamp(time.Date(2018, 1, 10, 4, 15, 0, 0, time.UTC))

		claims := jwt.Claims{
			Issuer:    ptr.Of("issuer"),
			Subject:   ptr.Of("subject"),
			Audience:  ptr.Of("audience"),
			ExpiresAt: ptr.Of(exp),
			NotBefore: ptr.Of(nbf),
			IssuedAt:  ptr.Of(iat),
			TokenID:   ptr.Of("token"),
		}

		token, _, keyID, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		assert.Equals(t, len(parts), 3)

		headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
		assert.NoError(t, err)
		var header jwt.Header
		err = json.Unmarshal(headerJSON, &header)
		assert.NoError(t, err)
		assert.Equals(t, header.Algorithm, string(jwt.EdDSA))
		assert.Equals(t, header.KeyID, keyID)
		assert.Equals(t, header.Type, "JWT")

		bodyJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
		assert.NoError(t, err)
		expectedBody := `{"aud":"audience","exp":"2024-06-01T12:00:00Z","iat":"2018-01-10T04:15:00Z","iss":"issuer","jti":"token","nbf":"2021-03-20T08:30:00Z","sub":"subject"}`
		assert.Equals(t, string(bodyJSON), expectedBody)
		var decodedClaims jwt.Claims
		err = json.Unmarshal(bodyJSON, &decodedClaims)
		assert.NoError(t, err)
		assert.Equals(t, decodedClaims, claims)
	})

	t.Run("when zero values are supplied it should omit empty fields", func(t *testing.T) {
		t.Parallel()

		exp := jwt.NewTimestamp(time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC))

		claims := jwt.Claims{
			Subject:   ptr.Of("subject"),
			ExpiresAt: ptr.Of(exp),
		}

		token, _, _, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		assert.Equals(t, len(parts), 3)

		bodyJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
		assert.NoError(t, err)
		assert.Equals(t, string(bodyJSON), `{"exp":"2024-06-01T12:00:00Z","sub":"subject"}`)
	})

	t.Run("when no claims are provided it should encode an empty body", func(t *testing.T) {
		t.Parallel()

		token, _, _, err := jwt.Encode(jwt.Claims{}, jwt.EdDSA)
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		assert.Equals(t, len(parts), 3)

		bodyJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
		assert.NoError(t, err)
		assert.Equals(t, string(bodyJSON), `{}`)
	})

	t.Run("when fields require escaping it should encode escaped strings", func(t *testing.T) {
		t.Parallel()

		claims := jwt.Claims{
			Audience: ptr.Of("audience \"quoted\" path\\folder"),
			Issuer:   ptr.Of("issuer-with-escape\\"),
			TokenID:  ptr.Of("token \"complex\" value\\id"),
		}

		token, _, _, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		assert.Equals(t, len(parts), 3)

		bodyJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
		assert.NoError(t, err)
		assert.Equals(t, string(bodyJSON), `{"aud":"audience \"quoted\" path\\folder","iss":"issuer-with-escape\\","jti":"token \"complex\" value\\id"}`)
	})

	t.Run("when timestamps with different timezones are used it should normalize to UTC", func(t *testing.T) {
		t.Parallel()

		claims := jwt.Claims{
			Audience:  ptr.Of("timezone-audience"),
			ExpiresAt: ptr.Of(jwt.NewTimestamp(time.Date(2030, 12, 25, 10, 30, 0, 0, time.FixedZone("EST", -5*3600)))),
			IssuedAt:  ptr.Of(jwt.NewTimestamp(time.Date(2020, 6, 15, 14, 45, 0, 0, time.FixedZone("PST", -8*3600)))),
			NotBefore: ptr.Of(jwt.NewTimestamp(time.Date(2025, 9, 1, 0, 0, 0, 0, time.FixedZone("CET", 1*3600)))),
			Subject:   ptr.Of("subject"),
		}

		token, _, _, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		assert.Equals(t, len(parts), 3)

		bodyJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
		assert.NoError(t, err)
		assert.Equals(t, string(bodyJSON), `{"aud":"timezone-audience","exp":"2030-12-25T15:30:00Z","iat":"2020-06-15T22:45:00Z","nbf":"2025-08-31T23:00:00Z","sub":"subject"}`)
	})

	t.Run("when timestamp fields are zero it should omit them from encoded body", func(t *testing.T) {
		t.Parallel()

		claims := jwt.Claims{
			Subject:   ptr.Of("subject"),
			ExpiresAt: ptr.Of(jwt.Timestamp{}),
			NotBefore: ptr.Of(jwt.Timestamp{}),
			IssuedAt:  ptr.Of(jwt.Timestamp{}),
		}

		token, _, _, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		assert.Equals(t, len(parts), 3)

		bodyJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
		assert.NoError(t, err)
		assert.Equals(t, string(bodyJSON), `{"sub":"subject"}`)
	})
}
