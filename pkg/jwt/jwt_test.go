package jwt_test

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestJWT(t *testing.T) {
	t.Parallel()

	t.Run("when encoding with EdDSA it should return a valid token key and key ID", func(t *testing.T) {
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
	})

	t.Run("when decoding an empty token it should return an error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		claims, err := jwt.Decode(ctx, "", func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return nil, jwt.EdDSA, nil
		})
		assert.Nil(t, claims)
		assert.ErrorExact(t, err, "token cannot be empty")
	})

	t.Run("when decoding a token with one segment it should return an error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		claims, err := jwt.Decode(ctx, "header-only", func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return nil, jwt.EdDSA, nil
		})
		assert.Nil(t, claims)
		assert.ErrorExact(t, err, "token must contain header, body, and signature")
	})

	t.Run("when decoding a token with two segments it should return an error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		claims, err := jwt.Decode(ctx, "header.body", func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return nil, jwt.EdDSA, nil
		})
		assert.Nil(t, claims)
		assert.ErrorExact(t, err, "token must contain header, body, and signature")
	})

	t.Run("when decoding a token with four segments it should return an error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		claims, err := jwt.Decode(ctx, "header.body.signature.extra", func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return nil, jwt.EdDSA, nil
		})
		assert.Nil(t, claims)
		assert.ErrorExact(t, err, "token must contain header, body, and signature")
	})

	t.Run("when decoding a token with invalid base64 header it should return an error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		claims, err := jwt.Decode(ctx, "not-valid-base64!@#.body.signature", func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return nil, jwt.EdDSA, nil
		})
		assert.Nil(t, claims)
		assert.ErrorPart(t, err, "failed to decode header")
	})

	t.Run("when decoding a token with invalid JSON header it should return an error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		invalidHeader := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
		claims, err := jwt.Decode(ctx, invalidHeader+".body.signature", func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return nil, jwt.EdDSA, nil
		})
		assert.Nil(t, claims)
		assert.ErrorPart(t, err, "json unmarshal error")
	})

	t.Run("when key provider returns an error it should return an error", func(t *testing.T) {
		t.Parallel()
		claims := jwt.Claims{
			Issuer: ptr.Of("test-issuer"),
		}
		token, _, _, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		ctx := context.Background()
		decodedClaims, err := jwt.Decode(ctx, token, func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return nil, "", errors.New("key not found")
		})
		assert.Nil(t, decodedClaims)
		assert.ErrorPart(t, err, "failed to retrieve key")
		assert.ErrorPart(t, err, "key not found")
	})

	t.Run("when token algorithm does not match expected algorithm it should return an error", func(t *testing.T) {
		t.Parallel()
		claims := jwt.Claims{
			Issuer: ptr.Of("test-issuer"),
		}
		token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		ctx := context.Background()
		decodedClaims, err := jwt.Decode(ctx, token, func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return key, jwt.SignatureAlgorithm("HS256"), nil
		})
		assert.Nil(t, decodedClaims)
		assert.ErrorExact(t, err, "token algorithm does not match expected algorithm")
	})

	t.Run("when key provider returns unknown algorithm it should return an error", func(t *testing.T) {
		t.Parallel()
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"Unknown","typ":"JWT","kid":"test-key"}`))
		body := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
		signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
		token := header + "." + body + "." + signature

		ctx := context.Background()
		decodedClaims, err := jwt.Decode(ctx, token, func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return []byte("key"), jwt.SignatureAlgorithm("Unknown"), nil
		})
		assert.Nil(t, decodedClaims)
		assert.ErrorExact(t, err, "failed to resolve signature provider")
	})

	t.Run("when decoding a token with invalid base64 body it should return an error", func(t *testing.T) {
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

		ctx := context.Background()
		decodedClaims, err := jwt.Decode(ctx, token, func(ctx context.Context, reqKeyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return privateKey, jwt.EdDSA, nil
		})
		assert.Nil(t, decodedClaims)
		assert.ErrorPart(t, err, "failed to decode body")
	})

	t.Run("when decoding a token with invalid JSON body it should return an error", func(t *testing.T) {
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

		ctx := context.Background()
		decodedClaims, err := jwt.Decode(ctx, token, func(ctx context.Context, reqKeyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return privateKey, jwt.EdDSA, nil
		})
		assert.Nil(t, decodedClaims)
		assert.ErrorPart(t, err, "json unmarshal error")
	})

	t.Run("when decoding a valid token it should return the claims", func(t *testing.T) {
		t.Parallel()
		claims := jwt.Claims{
			Issuer:   ptr.Of("test-issuer"),
			Subject:  ptr.Of("test-subject"),
			Audience: ptr.Of("test-audience"),
			TokenID:  ptr.Of("test-token-id"),
		}
		token, key, keyID, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		ctx := context.Background()
		decodedClaims, err := jwt.Decode(ctx, token, func(ctx context.Context, reqKeyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			assert.Equals(t, reqKeyID, keyID)
			return key, jwt.EdDSA, nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, decodedClaims)
		assert.Equals(t, *decodedClaims.Issuer, "test-issuer")
		assert.Equals(t, *decodedClaims.Subject, "test-subject")
		assert.Equals(t, *decodedClaims.Audience, "test-audience")
		assert.Equals(t, *decodedClaims.TokenID, "test-token-id")
	})

	t.Run("when decoding with context it should pass context to key provider", func(t *testing.T) {
		t.Parallel()
		type contextKey string
		claims := jwt.Claims{
			Issuer: ptr.Of("test-issuer"),
		}
		token, key, _, err := jwt.Encode(claims, jwt.EdDSA)
		assert.NoError(t, err)

		ctx := context.WithValue(context.Background(), contextKey("test-key"), "test-value")
		decodedClaims, err := jwt.Decode(ctx, token, func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			val := ctx.Value(contextKey("test-key"))
			assert.NotNil(t, val)
			assert.Equals(t, val.(string), "test-value")
			return key, jwt.EdDSA, nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, decodedClaims)
	})

	t.Run("when signature is corrupted it should return an error", func(t *testing.T) {
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

		ctx := context.Background()
		decodedClaims, err := jwt.Decode(ctx, corruptedToken, func(ctx context.Context, keyID string) ([]byte, jwt.SignatureAlgorithm, error) {
			return key, jwt.EdDSA, nil
		})
		assert.Nil(t, decodedClaims)
		assert.ErrorPart(t, err, "failed to verify token")
	})
}
