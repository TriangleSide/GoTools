package jwt_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestJSONEncoding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		body           jwt.Body
		key            string
		keyID          string
		expectedHeader string
		expectedBody   string
	}{
		{
			name: "when all fields are set it should encode fields in sorted order",
			body: jwt.Body{
				Issuer:    "issuer",
				Subject:   "subject",
				Audience:  "audience",
				ExpiresAt: 1717171717,
				NotBefore: 1616161616,
				IssuedAt:  1515151515,
				TokenID:   "token",
			},
			key:            "secret",
			keyID:          "kid",
			expectedHeader: `{"alg":"HS512","kid":"kid","typ":"JWT"}`,
			expectedBody:   `{"aud":"audience","exp":1717171717,"iat":1515151515,"iss":"issuer","jti":"token","nbf":1616161616,"sub":"subject"}`,
		},
		{
			name: "when zero values are supplied it should omit empty fields",
			body: jwt.Body{
				Subject:   "subject",
				ExpiresAt: 42,
			},
			key:            "spare-key",
			keyID:          "",
			expectedHeader: `{"alg":"HS512","typ":"JWT"}`,
			expectedBody:   `{"exp":42,"sub":"subject"}`,
		},
		{
			name:           "when no claims are provided it should encode an empty body",
			body:           jwt.Body{},
			key:            "empty-claims",
			keyID:          "kid-only",
			expectedHeader: `{"alg":"HS512","kid":"kid-only","typ":"JWT"}`,
			expectedBody:   `{}`,
		},
		{
			name: "when fields require escaping it should encode escaped strings",
			body: jwt.Body{
				Audience: "audience \"quoted\" path\\folder",
				Issuer:   "issuer-with-escape\\",
				TokenID:  "token \"complex\" value\\id",
			},
			key:            "escape-key",
			keyID:          "",
			expectedHeader: `{"alg":"HS512","typ":"JWT"}`,
			expectedBody:   `{"aud":"audience \"quoted\" path\\folder","iss":"issuer-with-escape\\","jti":"token \"complex\" value\\id"}`,
		},
		{
			name: "when negative timestamps are used it should preserve ordering and sign",
			body: jwt.Body{
				Audience:  "negative-audience",
				ExpiresAt: -10,
				IssuedAt:  -20,
				NotBefore: -30,
				Subject:   "subject",
			},
			key:            "negative-key",
			keyID:          "negative-kid",
			expectedHeader: `{"alg":"HS512","kid":"negative-kid","typ":"JWT"}`,
			expectedBody:   `{"aud":"negative-audience","exp":-10,"iat":-20,"nbf":-30,"sub":"subject"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token, err := jwt.Encode(tc.body, []byte(tc.key), tc.keyID)
			assert.NoError(t, err)

			parts := strings.Split(token, ".")
			assert.Equals(t, len(parts), 3)

			headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
			assert.NoError(t, err)
			assert.Equals(t, string(headerJSON), tc.expectedHeader)
			var header jwt.Header
			err = json.Unmarshal(headerJSON, &header)
			assert.NoError(t, err)
			assert.Equals(t, header.Algorithm, "HS512")
			assert.Equals(t, header.KeyID, tc.keyID)
			assert.Equals(t, header.Type, "JWT")

			bodyJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			assert.NoError(t, err)
			assert.Equals(t, string(bodyJSON), tc.expectedBody)
			var body jwt.Body
			err = json.Unmarshal(bodyJSON, &body)
			assert.NoError(t, err)
			assert.Equals(t, body, tc.body)
		})
	}
}
