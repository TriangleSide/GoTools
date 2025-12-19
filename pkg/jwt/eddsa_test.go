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

func eddsaTestClaims() jwt.Claims {
	return jwt.Claims{
		Issuer:   ptr.Of("issuer-eddsa"),
		Subject:  ptr.Of("subject-eddsa"),
		Audience: ptr.Of("audience-eddsa"),
		TokenID:  ptr.Of("token-eddsa"),
	}
}

func TestEdDSAProvider_InvalidVerifyingKey_ReturnsError(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, _, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	_, err = jwt.Decode(ctx, token, func(context.Context, string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		return []byte("short"), jwt.EdDSA, nil
	})
	assert.ErrorPart(t, err, "failed to use public key")
}

func TestEdDSAProvider_VerifyWithGeneratedPublicKey_Succeeds(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, publicKey, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	decoded, err := jwt.Decode(ctx, token, func(context.Context, string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		return publicKey, jwt.EdDSA, nil
	})
	assert.NoError(t, err)
	assert.Equals(t, decoded.Issuer, claims.Issuer)
}

func TestEdDSAProvider_InvalidSignatureLength_ReturnsError(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, publicKey, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	parts := strings.Split(token, ".")
	parts[2] = base64.RawURLEncoding.EncodeToString([]byte("short"))
	token = strings.Join(parts, ".")

	ctx := t.Context()
	_, err = jwt.Decode(ctx, token, func(context.Context, string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		return publicKey, jwt.EdDSA, nil
	})
	assert.ErrorPart(t, err, "eddsa signature must be 64 bytes")
}

func TestEdDSAProvider_VerifyWithWrongKey_Fails(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, _, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	secondarySeed := sha256.Sum256([]byte("eddsa-provider-secondary"))
	secondaryPrivateKey := ed25519.NewKeyFromSeed(secondarySeed[:])
	secondaryPublicKey := secondaryPrivateKey.Public().(ed25519.PublicKey)

	ctx := t.Context()
	_, err = jwt.Decode(ctx, token, func(context.Context, string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		return jwt.PublicKey(secondaryPublicKey), jwt.EdDSA, nil
	})
	assert.Error(t, err)
}

func TestEdDSAProvider_TamperedPayload_RejectsToken(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, publicKey, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	tamperedClaims := claims
	tamperedClaims.TokenID = ptr.Of("tampered-eddsa")

	tamperedToken, _, _, err := jwt.Encode(tamperedClaims, jwt.EdDSA)
	assert.NoError(t, err)

	parts := strings.Split(token, ".")
	tamperedParts := strings.Split(tamperedToken, ".")
	parts[1] = tamperedParts[1]
	tampered := strings.Join(parts, ".")

	ctx := t.Context()
	_, err = jwt.Decode(ctx, tampered, func(context.Context, string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		return publicKey, jwt.EdDSA, nil
	})
	assert.ErrorPart(t, err, "token signature is invalid")
}

func TestEdDSAProvider_MismatchedAlgorithm_ReturnsError(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, publicKey, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	_, err = jwt.Decode(ctx, token, func(context.Context, string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		return publicKey, jwt.SignatureAlgorithm("RS256"), nil
	})
	assert.ErrorPart(t, err, "token algorithm does not match expected algorithm")
}
