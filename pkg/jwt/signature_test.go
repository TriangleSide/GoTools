package jwt_test

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
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
		algorithm jwt.SignatureAlgorithm
		verifyKey []byte
		wrongKey  []byte
	}{
		{
			algorithm: jwt.EdDSA,
			verifyKey: primaryPublicKey,
			wrongKey:  secondaryPublicKey,
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
				sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
				if err != nil || len(sigBytes) == 0 {
					return token
				}
				sigBytes[0] ^= 0xFF
				parts[2] = base64.RawURLEncoding.EncodeToString(sigBytes)
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

				token, key, keyID, err := jwt.Encode(claims, alg.algorithm)
				assert.NoError(t, err)

				ctx := context.Background()
				_, err = jwt.Decode(ctx, token, func(ctx context.Context, requestedKeyID string) ([]byte, jwt.SignatureAlgorithm, error) {
					assert.Equals(t, requestedKeyID, keyID)
					return alg.wrongKey, alg.algorithm, nil
				})
				assert.Error(t, err)

				if tc.mutateToken != nil {
					token = tc.mutateToken(token)
				}

				decodedBody, err := jwt.Decode(ctx, token, func(ctx context.Context, requestedKeyID string) ([]byte, jwt.SignatureAlgorithm, error) {
					assert.Equals(t, requestedKeyID, keyID)
					return key, alg.algorithm, nil
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

	t.Run("when encoding with unknown algorithm it should return an error", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := jwt.Encode(jwt.Claims{}, jwt.SignatureAlgorithm("Unknown"))
		assert.ErrorPart(t, err, "failed to resolve signature provider")
	})

	t.Run("when encoding with EdDSA it should return a valid key and key ID derived from the key", func(t *testing.T) {
		t.Parallel()

		claims := jwt.Claims{
			Issuer:  ptr.Of("issuer-keygen"),
			Subject: ptr.Of("subject-keygen"),
		}

		token, key, keyID, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)
		assert.Equals(t, len(key), ed25519.PrivateKeySize)
		assert.NotEquals(t, keyID, "")
		assert.NotEquals(t, token, "")

		ctx := context.Background()
		decoded, err := jwt.Decode(ctx, token, func(ctx context.Context, reqKeyId string) ([]byte, jwt.SignatureAlgorithm, error) {
			assert.Equals(t, reqKeyId, keyID)
			return key, jwt.EdDSA, nil
		})
		assert.NoError(t, err)
		assert.Equals(t, *decoded.Issuer, *claims.Issuer)
		assert.Equals(t, *decoded.Subject, *claims.Subject)
	})

	t.Run("when encoding multiple tokens with EdDSA keys should be unique and key IDs should be derived from keys", func(t *testing.T) {
		t.Parallel()

		claims := jwt.Claims{
			Issuer: ptr.Of("issuer"),
		}

		_, key1, keyID1, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		_, key2, keyID2, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		assert.NotEquals(t, string(key1), string(key2))
		assert.NotEquals(t, keyID1, keyID2)

		hash1 := sha256.Sum256(key1)
		expectedKeyID1 := base64.RawURLEncoding.EncodeToString(hash1[:])
		assert.Equals(t, keyID1, expectedKeyID1)

		hash2 := sha256.Sum256(key2)
		expectedKeyID2 := base64.RawURLEncoding.EncodeToString(hash2[:])
		assert.Equals(t, keyID2, expectedKeyID2)
	})
}
