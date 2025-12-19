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

func TestEdDSA_ValidToken_DecodesSuccessfully(t *testing.T) {
	t.Parallel()

	primarySeed := sha256.Sum256([]byte("eddsa-primary"))
	primaryPrivateKey := ed25519.NewKeyFromSeed(primarySeed[:])

	secondarySeed := sha256.Sum256([]byte("eddsa-secondary"))
	secondaryPrivateKey := ed25519.NewKeyFromSeed(secondarySeed[:])
	secondaryPublicKey := secondaryPrivateKey.Public().(ed25519.PublicKey)

	claims := jwt.Claims{
		Issuer:   ptr.Of("issuer-" + string(jwt.EdDSA)),
		Subject:  ptr.Of("subject-" + string(jwt.EdDSA)),
		Audience: ptr.Of("audience-" + string(jwt.EdDSA)),
		TokenID:  ptr.Of("token-" + string(jwt.EdDSA)),
	}

	token, publicKey, keyID, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	keyProviderSecondary := func(_ context.Context, requestedKeyID string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		assert.Equals(t, requestedKeyID, keyID)
		return jwt.PublicKey(secondaryPublicKey), jwt.EdDSA, nil
	}
	_, err = jwt.Decode(t.Context(), token, keyProviderSecondary)
	assert.Error(t, err)

	_ = primaryPrivateKey

	keyProvider := func(_ context.Context, requestedKeyID string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		assert.Equals(t, requestedKeyID, keyID)
		return publicKey, jwt.EdDSA, nil
	}
	decodedBody, err := jwt.Decode(t.Context(), token, keyProvider)

	assert.NoError(t, err)
	assert.NotNil(t, decodedBody)
	assert.Equals(t, *decodedBody, claims)
}

func TestEdDSA_ModifiedSignature_FailsVerification(t *testing.T) {
	t.Parallel()

	secondarySeed := sha256.Sum256([]byte("eddsa-secondary"))
	secondaryPrivateKey := ed25519.NewKeyFromSeed(secondarySeed[:])
	secondaryPublicKey := secondaryPrivateKey.Public().(ed25519.PublicKey)

	claims := jwt.Claims{
		Issuer:   ptr.Of("issuer-" + string(jwt.EdDSA)),
		Subject:  ptr.Of("subject-" + string(jwt.EdDSA)),
		Audience: ptr.Of("audience-" + string(jwt.EdDSA)),
		TokenID:  ptr.Of("token-" + string(jwt.EdDSA)),
	}

	token, publicKey, keyID, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	wrongKeyCallback := func(_ context.Context, reqKeyID string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		assert.Equals(t, reqKeyID, keyID)
		return jwt.PublicKey(secondaryPublicKey), jwt.EdDSA, nil
	}
	_, err = jwt.Decode(ctx, token, wrongKeyCallback)
	assert.Error(t, err)

	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		sigBytes, decodeErr := base64.RawURLEncoding.DecodeString(parts[2])
		if decodeErr == nil && len(sigBytes) > 0 {
			sigBytes[0] ^= 0xFF
			parts[2] = base64.RawURLEncoding.EncodeToString(sigBytes)
			token = strings.Join(parts, ".")
		}
	}

	keyProvider := func(_ context.Context, requestedKeyID string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		assert.Equals(t, requestedKeyID, keyID)
		return publicKey, jwt.EdDSA, nil
	}
	decodedBody, err := jwt.Decode(ctx, token, keyProvider)

	assert.ErrorPart(t, err, "failed to verify token")
	assert.Nil(t, decodedBody)
}

func TestEdDSA_InvalidBase64Signature_ReturnsDecodeError(t *testing.T) {
	t.Parallel()

	secondarySeed := sha256.Sum256([]byte("eddsa-secondary"))
	secondaryPrivateKey := ed25519.NewKeyFromSeed(secondarySeed[:])
	secondaryPublicKey := secondaryPrivateKey.Public().(ed25519.PublicKey)

	claims := jwt.Claims{
		Issuer:   ptr.Of("issuer-" + string(jwt.EdDSA)),
		Subject:  ptr.Of("subject-" + string(jwt.EdDSA)),
		Audience: ptr.Of("audience-" + string(jwt.EdDSA)),
		TokenID:  ptr.Of("token-" + string(jwt.EdDSA)),
	}

	token, publicKey, keyID, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	keyCallback := func(_ context.Context, reqKeyID string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		assert.Equals(t, reqKeyID, keyID)
		return jwt.PublicKey(secondaryPublicKey), jwt.EdDSA, nil
	}
	_, err = jwt.Decode(ctx, token, keyCallback)
	assert.Error(t, err)

	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		parts[2] += "?"
		token = strings.Join(parts, ".")
	}

	keyProvider := func(_ context.Context, requestedKeyID string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		assert.Equals(t, requestedKeyID, keyID)
		return publicKey, jwt.EdDSA, nil
	}
	decodedBody, err := jwt.Decode(ctx, token, keyProvider)

	assert.ErrorPart(t, err, "failed to decode signature")
	assert.Nil(t, decodedBody)
}

func TestEncode_UnknownAlgorithm_ReturnsError(t *testing.T) {
	t.Parallel()

	token, publicKey, keyID, err := jwt.Encode(jwt.Claims{}, jwt.SignatureAlgorithm("Unknown"))
	assert.ErrorPart(t, err, "failed to resolve signature provider")
	assert.Equals(t, token, "")
	assert.Nil(t, publicKey)
	assert.Equals(t, keyID, "")
}

func TestEncode_EdDSA_ReturnsValidKeyAndKeyID(t *testing.T) {
	t.Parallel()

	claims := jwt.Claims{
		Issuer:  ptr.Of("issuer-keygen"),
		Subject: ptr.Of("subject-keygen"),
	}

	token, publicKey, keyID, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)
	assert.Equals(t, len(publicKey), ed25519.PublicKeySize)
	assert.NotEquals(t, keyID, "")
	assert.NotEquals(t, token, "")

	ctx := t.Context()
	keyProvider := func(_ context.Context, reqKeyId string) (jwt.PublicKey, jwt.SignatureAlgorithm, error) {
		assert.Equals(t, reqKeyId, keyID)
		return publicKey, jwt.EdDSA, nil
	}
	decoded, err := jwt.Decode(ctx, token, keyProvider)
	assert.NoError(t, err)
	assert.Equals(t, *decoded.Issuer, *claims.Issuer)
	assert.Equals(t, *decoded.Subject, *claims.Subject)
}

func TestEncode_MultipleEdDSATokens_GeneratesUniqueKeysAndKeyIDs(t *testing.T) {
	t.Parallel()

	claims := jwt.Claims{
		Issuer: ptr.Of("issuer"),
	}

	_, publicKey1, keyID1, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	_, publicKey2, keyID2, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	assert.NotEquals(t, string(publicKey1), string(publicKey2))
	assert.NotEquals(t, keyID1, keyID2)

	hash1 := sha256.Sum256(publicKey1)
	expectedKeyID1 := base64.RawURLEncoding.EncodeToString(hash1[:])
	assert.Equals(t, keyID1, expectedKeyID1)

	hash2 := sha256.Sum256(publicKey2)
	expectedKeyID2 := base64.RawURLEncoding.EncodeToString(hash2[:])
	assert.Equals(t, keyID2, expectedKeyID2)
}
