package jwt_test

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/ptr"
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
				parts[2] += "?"
				return strings.Join(parts, ".")
			},
			expectErr: "failed to decode signature",
		},
	}

	for _, tc := range testCases {
		for _, alg := range algorithms {
			t.Run("when using "+string(alg.algorithm)+" and "+tc.condition+" it should "+tc.expectation, func(t *testing.T) {
				t.Parallel()

				claims := jwt.Claims{
					Issuer:   ptr.Of("issuer-" + string(alg.algorithm)),
					Subject:  ptr.Of("subject-" + string(alg.algorithm)),
					Audience: ptr.Of("audience-" + string(alg.algorithm)),
					TokenID:  ptr.Of("token-" + string(alg.algorithm)),
				}

				token, err := jwt.Encode(claims, alg.signingKey, alg.keyID, alg.algorithm)
				assert.NoError(t, err)

				ctx := context.Background()
				_, err = jwt.Decode(ctx, token, func(ctx context.Context, requestedKeyID string) ([]byte, jwt.SignatureAlgorithm, error) {
					assert.Equals(t, requestedKeyID, alg.keyID)
					return alg.wrongKey, alg.algorithm, nil
				})
				assert.Error(t, err)

				if tc.mutateToken != nil {
					token = tc.mutateToken(token)
				}

				decodedBody, err := jwt.Decode(ctx, token, func(ctx context.Context, requestedKeyID string) ([]byte, jwt.SignatureAlgorithm, error) {
					assert.Equals(t, requestedKeyID, alg.keyID)
					return alg.verifyKey, alg.algorithm, nil
				})

				if tc.expectErr != "" {
					assert.ErrorPart(t, err, tc.expectErr)
					assert.Nil(t, decodedBody)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, decodedBody)
				assert.Equals(t, *decodedBody, claims)
			})
		}
	}
}
