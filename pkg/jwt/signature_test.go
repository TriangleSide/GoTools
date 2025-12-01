package jwt_test

import (
	"crypto/ed25519"
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestSignatureAlgorithms(t *testing.T) {
	t.Parallel()

	primarySeed := sha256.Sum256([]byte("eddsa-primary"))
	primaryPrivateKey := ed25519.NewKeyFromSeed(primarySeed[:])
	primaryPublicKey := primaryPrivateKey.Public().(ed25519.PublicKey)

	secondarySeed := sha256.Sum256([]byte("eddsa-secondary"))
	secondaryPrivateKey := ed25519.NewKeyFromSeed(secondarySeed[:])
	secondaryPublicKey := secondaryPrivateKey.Public().(ed25519.PublicKey)

	algorithms := []struct {
		algorithm  jwt.SignatureAlgorithm
		signingKey []byte
		verifyKey  []byte
		keyID      string
		wrongKey   []byte
	}{
		{
			algorithm:  jwt.HS512,
			signingKey: []byte(string(jwt.HS512) + "-secret"),
			verifyKey:  []byte(string(jwt.HS512) + "-secret"),
			keyID:      "kid-" + string(jwt.HS512),
			wrongKey:   []byte("wrong-secret"),
		},
		{
			algorithm:  jwt.EdDSA,
			signingKey: primaryPrivateKey,
			verifyKey:  primaryPublicKey,
			keyID:      "kid-" + string(jwt.EdDSA),
			wrongKey:   secondaryPublicKey,
		},
	}

	testCases := []struct {
		condition   string
		expectation string
		mutateToken func(string) string
		expectErr   string
	}{
		{
			condition:   "token is valid",
			expectation: "decode signed token",
		},
		{
			condition:   "signature is modified",
			expectation: "fail verification",
			mutateToken: func(token string) string {
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					return token
				}
				signature := []byte(parts[2])
				if len(signature) == 0 {
					return token
				}
				if signature[len(signature)-1] == 'A' {
					signature[len(signature)-1] = 'B'
				} else {
					signature[len(signature)-1] = 'A'
				}
				parts[2] = string(signature)
				return strings.Join(parts, ".")
			},
			expectErr: "failed to verify token",
		},
		{
			condition:   "signature is not base64 encoded",
			expectation: "return decode error",
			mutateToken: func(token string) string {
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					return token
				}
				parts[2] = parts[2] + "?"
				return strings.Join(parts, ".")
			},
			expectErr: "failed to decode signature",
		},
	}

	for _, tc := range testCases {
		for _, alg := range algorithms {
			t.Run("when using "+string(alg.algorithm)+" and "+tc.condition+" it should "+tc.expectation, func(t *testing.T) {
				t.Parallel()

				body := jwt.Body{
					Issuer:   "issuer-" + string(alg.algorithm),
					Subject:  "subject-" + string(alg.algorithm),
					Audience: "audience-" + string(alg.algorithm),
					TokenID:  "token-" + string(alg.algorithm),
				}

				token, err := jwt.Encode(body, alg.signingKey, alg.keyID, jwt.WithSignatureAlgorithm(alg.algorithm))
				assert.NoError(t, err)

				_, err = jwt.Decode(token, func(requestedKeyID string) ([]byte, error) {
					assert.Equals(t, requestedKeyID, alg.keyID)
					return alg.wrongKey, nil
				})
				assert.Error(t, err)

				if tc.mutateToken != nil {
					token = tc.mutateToken(token)
				}

				decodedBody, err := jwt.Decode(token, func(requestedKeyID string) ([]byte, error) {
					assert.Equals(t, requestedKeyID, alg.keyID)
					return alg.verifyKey, nil
				})

				if tc.expectErr != "" {
					assert.ErrorPart(t, err, tc.expectErr)
					assert.Nil(t, decodedBody)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, decodedBody)
				assert.Equals(t, *decodedBody, body)
			})
		}
	}
}
