package jwt_test

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestEdDSAProvider(t *testing.T) {
	t.Parallel()

	primarySeed := sha256.Sum256([]byte("eddsa-provider-primary"))
	primaryPrivateKey := ed25519.NewKeyFromSeed(primarySeed[:])
	primaryPublicKey := primaryPrivateKey.Public().(ed25519.PublicKey)

	secondarySeed := sha256.Sum256([]byte("eddsa-provider-secondary"))
	secondaryPrivateKey := ed25519.NewKeyFromSeed(secondarySeed[:])

	body := jwt.Body{
		Issuer:   "issuer-eddsa",
		Subject:  "subject-eddsa",
		Audience: "audience-eddsa",
		TokenID:  "token-eddsa",
	}

	t.Run("when signing key is invalid it should return an error", func(t *testing.T) {
		t.Parallel()

		_, err := jwt.Encode(body, []byte("short"), "eddsa", jwt.WithSignatureAlgorithm(jwt.EdDSA))
		assert.ErrorPart(t, err, "failed to sign token")
		assert.ErrorPart(t, err, "failed to use private key")
	})

	t.Run("when verifying key is invalid it should return an error", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(body, primaryPrivateKey, "eddsa", jwt.WithSignatureAlgorithm(jwt.EdDSA))
		assert.NoError(t, err)

		_, err = jwt.Decode(token, func(string) ([]byte, error) {
			return []byte("short"), nil
		})
		assert.ErrorPart(t, err, "failed to use public key")
	})

	t.Run("when verifying with a private key it should succeed", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(body, primaryPrivateKey, "eddsa", jwt.WithSignatureAlgorithm(jwt.EdDSA))
		assert.NoError(t, err)

		decoded, err := jwt.Decode(token, func(string) ([]byte, error) {
			return primaryPrivateKey, nil
		})
		assert.NoError(t, err)
		assert.Equals(t, decoded.Issuer, body.Issuer)
	})

	t.Run("when verifying with a public key it should succeed", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(body, primaryPrivateKey, "eddsa", jwt.WithSignatureAlgorithm(jwt.EdDSA))
		assert.NoError(t, err)

		decoded, err := jwt.Decode(token, func(string) ([]byte, error) {
			return primaryPublicKey, nil
		})
		assert.NoError(t, err)
		assert.Equals(t, decoded.Issuer, body.Issuer)
	})

	t.Run("when signature length is invalid it should return an error", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(body, primaryPrivateKey, "eddsa", jwt.WithSignatureAlgorithm(jwt.EdDSA))
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		parts[2] = base64.RawURLEncoding.EncodeToString([]byte("short"))
		token = strings.Join(parts, ".")

		_, err = jwt.Decode(token, func(string) ([]byte, error) {
			return primaryPublicKey, nil
		})
		assert.ErrorPart(t, err, "eddsa signature must be 64 bytes")
	})

	t.Run("when verifying with the wrong private key it should fail", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(body, primaryPrivateKey, "eddsa", jwt.WithSignatureAlgorithm(jwt.EdDSA))
		assert.NoError(t, err)

		_, err = jwt.Decode(token, func(string) ([]byte, error) {
			return secondaryPrivateKey, nil
		})
		assert.Error(t, err)
	})

	t.Run("when payload is tampered it should reject the token", func(t *testing.T) {
		t.Parallel()

		token, err := jwt.Encode(body, primaryPrivateKey, "eddsa", jwt.WithSignatureAlgorithm(jwt.EdDSA))
		assert.NoError(t, err)

		tamperedBody := body
		tamperedBody.TokenID = "tampered-eddsa"

		tamperedToken, err := jwt.Encode(tamperedBody, primaryPrivateKey, "eddsa", jwt.WithSignatureAlgorithm(jwt.EdDSA))
		assert.NoError(t, err)

		parts := strings.Split(token, ".")
		tamperedParts := strings.Split(tamperedToken, ".")
		parts[1] = tamperedParts[1]
		tampered := strings.Join(parts, ".")

		_, err = jwt.Decode(tampered, func(string) ([]byte, error) {
			return primaryPublicKey, nil
		})
		assert.ErrorPart(t, err, "token signature is invalid")
	})
}
