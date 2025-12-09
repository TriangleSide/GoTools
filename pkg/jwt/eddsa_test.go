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

	ctx := context.Background()
	_, err = jwt.Decode(ctx, token, func(ctx context.Context, keyId string) ([]byte, jwt.SignatureAlgorithm, error) {
		return []byte("short"), jwt.EdDSA, nil
	})
	assert.ErrorPart(t, err, "failed to use public key")
}

func TestEdDSAProvider_VerifyWithGeneratedPrivateKey_Succeeds(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := context.Background()
	decoded, err := jwt.Decode(ctx, token, func(ctx context.Context, keyId string) ([]byte, jwt.SignatureAlgorithm, error) {
		return key, jwt.EdDSA, nil
	})
	assert.NoError(t, err)
	assert.Equals(t, decoded.Issuer, claims.Issuer)
}

func TestEdDSAProvider_VerifyWithDerivedPublicKey_Succeeds(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	privateKey := ed25519.PrivateKey(key)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	ctx := context.Background()
	decoded, err := jwt.Decode(ctx, token, func(ctx context.Context, keyId string) ([]byte, jwt.SignatureAlgorithm, error) {
		return publicKey, jwt.EdDSA, nil
	})
	assert.NoError(t, err)
	assert.Equals(t, decoded.Issuer, claims.Issuer)
}

func TestEdDSAProvider_InvalidSignatureLength_ReturnsError(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	privateKey := ed25519.PrivateKey(key)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	parts := strings.Split(token, ".")
	parts[2] = base64.RawURLEncoding.EncodeToString([]byte("short"))
	token = strings.Join(parts, ".")

	ctx := context.Background()
	_, err = jwt.Decode(ctx, token, func(ctx context.Context, keyId string) ([]byte, jwt.SignatureAlgorithm, error) {
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

	ctx := context.Background()
	_, err = jwt.Decode(ctx, token, func(ctx context.Context, keyId string) ([]byte, jwt.SignatureAlgorithm, error) {
		return secondaryPrivateKey, jwt.EdDSA, nil
	})
	assert.Error(t, err)
}

func TestEdDSAProvider_TamperedPayload_RejectsToken(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	privateKey := ed25519.PrivateKey(key)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	tamperedClaims := claims
	tamperedClaims.TokenID = ptr.Of("tampered-eddsa")

	tamperedToken, _, _, err := jwt.Encode(tamperedClaims, jwt.EdDSA)
	assert.NoError(t, err)

	parts := strings.Split(token, ".")
	tamperedParts := strings.Split(tamperedToken, ".")
	parts[1] = tamperedParts[1]
	tampered := strings.Join(parts, ".")

	ctx := context.Background()
	_, err = jwt.Decode(ctx, tampered, func(ctx context.Context, keyId string) ([]byte, jwt.SignatureAlgorithm, error) {
		return publicKey, jwt.EdDSA, nil
	})
	assert.ErrorPart(t, err, "token signature is invalid")
}

func TestEdDSAProvider_MismatchedAlgorithm_ReturnsError(t *testing.T) {
	t.Parallel()

	claims := eddsaTestClaims()
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	privateKey := ed25519.PrivateKey(key)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	ctx := context.Background()
	_, err = jwt.Decode(ctx, token, func(ctx context.Context, keyId string) ([]byte, jwt.SignatureAlgorithm, error) {
		return publicKey, jwt.SignatureAlgorithm("RS256"), nil
	})
	assert.ErrorPart(t, err, "token algorithm does not match expected algorithm")
}
