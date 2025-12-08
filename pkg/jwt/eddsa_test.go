package jwt_test

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestEdDSAProvider(t *testing.T) {
	t.Parallel()

	primarySeed := sha256.Sum256([]byte("eddsa-provider-primary"))
	primaryPrivateKey := ed25519.NewKeyFromSeed(primarySeed[:])
	primaryPublicKey := primaryPrivateKey.Public().(ed25519.PublicKey)

	secondarySeed := sha256.Sum256([]byte("eddsa-provider-secondary"))
	secondaryPrivateKey := ed25519.NewKeyFromSeed(secondarySeed[:])

	claims := jwt.Claims{
		Issuer:   ptr.Of("issuer-eddsa"),
		Subject:  ptr.Of("subject-eddsa"),
		Audience: ptr.Of("audience-eddsa"),
		TokenID:  ptr.Of("token-eddsa"),
	}

	t.Run("when signing key is invalid it should return an error", func(t *testing.T) {
		t.Parallel()

		_, err := jwt.Encode(claims, []byte("short"), "eddsa", jwt.EdDSA)
		assert.ErrorPart(t, err, "failed to sign token")
		assert.ErrorPart(t, err, "failed to use private key")
	})

	t.Run("when verifying key is invalid it should return an error", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(claims, primaryPrivateKey, "eddsa", jwt.EdDSA)
		assert.NoError(t, err)

		_, err = jwt.Decode(token, func(string) ([]byte, jwt.SignatureAlgorithm, error) {
			return []byte("short"), jwt.EdDSA, nil
		})
		assert.ErrorPart(t, err, "failed to use public key")
	})

	t.Run("when verifying with a private key it should succeed", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(claims, primaryPrivateKey, "eddsa", jwt.EdDSA)
		assert.NoError(t, err)

		decoded, err := jwt.Decode(token, func(string) ([]byte, jwt.SignatureAlgorithm, error) {
			return primaryPrivateKey, jwt.EdDSA, nil
		})
		assert.NoError(t, err)
		assert.Equals(t, decoded.Issuer, claims.Issuer)
	})

	t.Run("when verifying with a public key it should succeed", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(claims, primaryPrivateKey, "eddsa", jwt.EdDSA)
		assert.NoError(t, err)

		decoded, err := jwt.Decode(token, func(string) ([]byte, jwt.SignatureAlgorithm, error) {
			return primaryPublicKey, jwt.EdDSA, nil
		})
		assert.NoError(t, err)
		assert.Equals(t, decoded.Issuer, claims.Issuer)
	})

	t.Run("when signature length is invalid it should return an error", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(claims, primaryPrivateKey, "eddsa", jwt.EdDSA)
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		parts[2] = base64.RawURLEncoding.EncodeToString([]byte("short"))
		token = strings.Join(parts, ".")

		_, err = jwt.Decode(token, func(string) ([]byte, jwt.SignatureAlgorithm, error) {
			return primaryPublicKey, jwt.EdDSA, nil
		})
		assert.ErrorPart(t, err, "eddsa signature must be 64 bytes")
	})

	t.Run("when verifying with the wrong private key it should fail", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(claims, primaryPrivateKey, "eddsa", jwt.EdDSA)
		assert.NoError(t, err)

		_, err = jwt.Decode(token, func(string) ([]byte, jwt.SignatureAlgorithm, error) {
			return secondaryPrivateKey, jwt.EdDSA, nil
		})
		assert.Error(t, err)
	})

	t.Run("when payload is tampered it should reject the token", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(claims, primaryPrivateKey, "eddsa", jwt.EdDSA)
		assert.NoError(t, err)

		tamperedClaims := claims
		tamperedClaims.TokenID = ptr.Of("tampered-eddsa")

		tamperedToken, err := jwt.Encode(tamperedClaims, primaryPrivateKey, "eddsa", jwt.EdDSA)
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		tamperedParts := strings.Split(tamperedToken, ".")
		parts[1] = tamperedParts[1]
		tampered := strings.Join(parts, ".")

		_, err = jwt.Decode(tampered, func(string) ([]byte, jwt.SignatureAlgorithm, error) {
			return primaryPublicKey, jwt.EdDSA, nil
		})
		assert.ErrorPart(t, err, "token signature is invalid")
	})

	t.Run("when key provider returns mismatched algorithm it should return an error", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(claims, primaryPrivateKey, "eddsa", jwt.EdDSA)
		assert.NoError(t, err)

		_, err = jwt.Decode(token, func(string) ([]byte, jwt.SignatureAlgorithm, error) {
			return primaryPublicKey, jwt.SignatureAlgorithm("RS256"), nil
		})
		assert.ErrorPart(t, err, "token algorithm does not match expected algorithm")
	})
}
