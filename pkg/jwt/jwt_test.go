package jwt_test

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/timestamp"
)

func TestEncode_WithEdDSA_ReturnsValidTokenKeyAndKeyID(t *testing.T) {
	t.Parallel()
	claims := jwt.Claims{
		Issuer:  ptr.Of("test-issuer"),
		Subject: ptr.Of("test-subject"),
	}
	token, key, keyID, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)
	assert.NotEquals(t, token, "")
	assert.NotEquals(t, len(key), 0)
	assert.NotEquals(t, keyID, "")
	parts := strings.Split(token, ".")
	assert.Equals(t, len(parts), 3)
}

func TestDecode_EmptyToken_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	claims, err := jwt.Decode(ctx, "", func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return nil, jwt.EdDSA, nil
	})
	assert.Nil(t, claims)
	assert.ErrorExact(t, err, "token cannot be empty")
}

func TestDecode_OneSegment_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	claims, err := jwt.Decode(ctx, "header-only", func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return nil, jwt.EdDSA, nil
	})
	assert.Nil(t, claims)
	assert.ErrorExact(t, err, "token must contain header, body, and signature")
}

func TestDecode_TwoSegments_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	claims, err := jwt.Decode(ctx, "header.body", func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return nil, jwt.EdDSA, nil
	})
	assert.Nil(t, claims)
	assert.ErrorExact(t, err, "token must contain header, body, and signature")
}

func TestDecode_FourSegments_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	keyProvider := func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return nil, jwt.EdDSA, nil
	}
	claims, err := jwt.Decode(ctx, "header.body.signature.extra", keyProvider)
	assert.Nil(t, claims)
	assert.ErrorExact(t, err, "token must contain header, body, and signature")
}

func TestDecode_InvalidBase64Header_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	keyProvider := func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return nil, jwt.EdDSA, nil
	}
	claims, err := jwt.Decode(ctx, "not-valid-base64!@#.body.signature", keyProvider)
	assert.Nil(t, claims)
	assert.ErrorPart(t, err, "failed to decode header")
}

func TestDecode_InvalidJSONHeader_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	invalidHeader := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
	keyProvider := func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return nil, jwt.EdDSA, nil
	}
	claims, err := jwt.Decode(ctx, invalidHeader+".body.signature", keyProvider)
	assert.Nil(t, claims)
	assert.ErrorPart(t, err, "json unmarshal error")
}

func TestDecode_KeyProviderReturnsError_ReturnsError(t *testing.T) {
	t.Parallel()
	claims := jwt.Claims{
		Issuer: ptr.Of("test-issuer"),
	}
	token, _, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	decodedClaims, err := jwt.Decode(ctx, token, func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return nil, "", errors.New("key not found")
	})
	assert.Nil(t, decodedClaims)
	assert.ErrorPart(t, err, "failed to retrieve key")
	assert.ErrorPart(t, err, "key not found")
}

func TestDecode_AlgorithmMismatch_ReturnsError(t *testing.T) {
	t.Parallel()
	claims := jwt.Claims{
		Issuer: ptr.Of("test-issuer"),
	}
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	decodedClaims, err := jwt.Decode(ctx, token, func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return key, jwt.SignatureAlgorithm("HS256"), nil
	})
	assert.Nil(t, decodedClaims)
	assert.ErrorExact(t, err, "token algorithm does not match expected algorithm")
}

func TestDecode_UnknownAlgorithm_ReturnsError(t *testing.T) {
	t.Parallel()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"Unknown","typ":"JWT","kid":"test-key"}`))
	body := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
	token := header + "." + body + "." + signature

	ctx := t.Context()
	decodedClaims, err := jwt.Decode(ctx, token, func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return []byte("key"), jwt.SignatureAlgorithm("Unknown"), nil
	})
	assert.Nil(t, decodedClaims)
	assert.ErrorPart(t, err, "failed to resolve signature provider")
}

func TestDecode_InvalidBase64Body_ReturnsError(t *testing.T) {
	t.Parallel()

	seed := sha256.Sum256([]byte("test-seed-invalid-base64-body"))
	privateKey := ed25519.NewKeyFromSeed(seed[:])

	keyHash := sha256.Sum256(privateKey)
	keyID := base64.RawURLEncoding.EncodeToString(keyHash[:])

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA","typ":"JWT","kid":"` + keyID + `"}`))
	invalidBody := "not-valid-base64!@#"
	signedData := header + "." + invalidBody
	signature := ed25519.Sign(privateKey, []byte(signedData))
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)
	token := signedData + "." + encodedSignature

	ctx := t.Context()
	decodedClaims, err := jwt.Decode(ctx, token, func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return privateKey, jwt.EdDSA, nil
	})
	assert.Nil(t, decodedClaims)
	assert.ErrorPart(t, err, "failed to decode body")
}

func TestDecode_InvalidJSONBody_ReturnsError(t *testing.T) {
	t.Parallel()

	seed := sha256.Sum256([]byte("test-seed-invalid-json-body"))
	privateKey := ed25519.NewKeyFromSeed(seed[:])

	keyHash := sha256.Sum256(privateKey)
	keyID := base64.RawURLEncoding.EncodeToString(keyHash[:])

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA","typ":"JWT","kid":"` + keyID + `"}`))
	invalidBody := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
	signedData := header + "." + invalidBody
	signature := ed25519.Sign(privateKey, []byte(signedData))
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)
	token := signedData + "." + encodedSignature

	ctx := t.Context()
	decodedClaims, err := jwt.Decode(ctx, token, func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return privateKey, jwt.EdDSA, nil
	})
	assert.Nil(t, decodedClaims)
	assert.ErrorPart(t, err, "json unmarshal error")
}

func TestDecode_ValidToken_ReturnsClaims(t *testing.T) {
	t.Parallel()
	claims := jwt.Claims{
		Issuer:   ptr.Of("test-issuer"),
		Subject:  ptr.Of("test-subject"),
		Audience: ptr.Of("test-audience"),
		TokenID:  ptr.Of("test-token-id"),
	}
	token, key, keyID, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	keyProvider := func(_ context.Context, reqKeyID string) ([]byte, jwt.SignatureAlgorithm, error) {
		assert.Equals(t, reqKeyID, keyID)
		return key, jwt.EdDSA, nil
	}
	decodedClaims, err := jwt.Decode(ctx, token, keyProvider)
	assert.NoError(t, err)
	assert.NotNil(t, decodedClaims)
	assert.Equals(t, *decodedClaims.Issuer, "test-issuer")
	assert.Equals(t, *decodedClaims.Subject, "test-subject")
	assert.Equals(t, *decodedClaims.Audience, "test-audience")
	assert.Equals(t, *decodedClaims.TokenID, "test-token-id")
}

func TestDecode_WithContext_PassesContextToKeyProvider(t *testing.T) {
	t.Parallel()
	type contextKey string
	claims := jwt.Claims{
		Issuer: ptr.Of("test-issuer"),
	}
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := context.WithValue(t.Context(), contextKey("test-key"), "test-value")
	decodedClaims, err := jwt.Decode(ctx, token, func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		val := ctx.Value(contextKey("test-key"))
		assert.NotNil(t, val)
		assert.Equals(t, val.(string), "test-value")
		return key, jwt.EdDSA, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, decodedClaims)
}

func TestDecode_CorruptedSignature_ReturnsError(t *testing.T) {
	t.Parallel()
	claims := jwt.Claims{
		Issuer: ptr.Of("test-issuer"),
	}
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	parts := strings.Split(token, ".")
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	assert.NoError(t, err)
	sigBytes[0] ^= 0xFF
	parts[2] = base64.RawURLEncoding.EncodeToString(sigBytes)
	corruptedToken := strings.Join(parts, ".")

	ctx := t.Context()
	keyProvider := func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return key, jwt.EdDSA, nil
	}
	decodedClaims, err := jwt.Decode(ctx, corruptedToken, keyProvider)
	assert.Nil(t, decodedClaims)
	assert.ErrorPart(t, err, "failed to verify token")
}

func TestDecode_ValidTokenWithTimestampClaims_ReturnsClaimsWithTimestamps(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC().Truncate(time.Second)
	claims := jwt.Claims{
		Issuer:    ptr.Of("test-issuer"),
		ExpiresAt: ptr.Of(timestamp.New(now.Add(time.Hour))),
		NotBefore: ptr.Of(timestamp.New(now.Add(-time.Minute))),
		IssuedAt:  ptr.Of(timestamp.New(now)),
	}
	token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
	assert.NoError(t, err)

	ctx := t.Context()
	decodedClaims, err := jwt.Decode(ctx, token, func(context.Context, string) ([]byte, jwt.SignatureAlgorithm, error) {
		return key, jwt.EdDSA, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, decodedClaims)
	assert.Equals(t, decodedClaims.ExpiresAt.Time(), claims.ExpiresAt.Time())
	assert.Equals(t, decodedClaims.NotBefore.Time(), claims.NotBefore.Time())
	assert.Equals(t, decodedClaims.IssuedAt.Time(), claims.IssuedAt.Time())
}
