package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"hash"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

const (
	testEncodedHeader = "encoded-header"
	testEncodedBody   = "encoded-body"
)

// failingHash implements hash.Hash and always fails on Write.
type failingHash struct{ _ byte }

func (f *failingHash) Write(p []byte) (n int, err error) { return 0, errors.New("write fail") }
func (f *failingHash) Sum(b []byte) []byte               { return []byte{} }
func (f *failingHash) Reset()                            {}
func (f *failingHash) Size() int                         { return 1 }
func (f *failingHash) BlockSize() int                    { return 1 }

func TestHash(t *testing.T) {
	t.Parallel()

	t.Run("it should hash header and body data", func(t *testing.T) {
		t.Parallel()
		secret := t.Name()

		hashed, err := hashData(testEncodedHeader, testEncodedBody, secret, sha256.New)
		assert.NoError(t, err)

		mac := hmac.New(sha256.New, []byte(secret))
		_, err = mac.Write([]byte(testEncodedHeader + "." + testEncodedBody))
		assert.NoError(t, err)
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		assert.Equals(t, hashed, expected)
	})

	t.Run("it should use a custom hash provider", func(t *testing.T) {
		t.Parallel()
		secret := t.Name() + "-custom"

		hashed, err := hashData(testEncodedHeader, testEncodedBody, secret, sha512.New)
		assert.NoError(t, err)

		mac := hmac.New(sha512.New, []byte(secret))
		_, err = mac.Write([]byte(testEncodedHeader + "." + testEncodedBody))
		assert.NoError(t, err)
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		assert.Equals(t, hashed, expected)
	})

	t.Run("when the provider is nil it should return an error", func(t *testing.T) {
		t.Parallel()
		secret := t.Name()

		h, err := hashData(testEncodedHeader, testEncodedBody, secret, nil)
		assert.ErrorPart(t, err, "hash provider cannot be nil")
		assert.Equals(t, h, "")
	})

	t.Run("when the hash provider returns an error on write it should return an error", func(t *testing.T) {
		t.Parallel()
		secret := t.Name()

		h, err := hashData(testEncodedHeader, testEncodedBody, secret, func() hash.Hash { return &failingHash{} })
		assert.ErrorPart(t, err, "failed to write data to hash")
		assert.Equals(t, h, "")
	})

	t.Run("it should verify matching hashes", func(t *testing.T) {
		t.Parallel()
		secret := t.Name()

		mac := hmac.New(sha256.New, []byte(secret))
		_, err := mac.Write([]byte(testEncodedHeader + "." + testEncodedBody))
		assert.NoError(t, err)

		signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		valid, err := verifyHash(testEncodedHeader, testEncodedBody, signature, secret, sha256.New)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("it should return false when the signature does not match", func(t *testing.T) {
		t.Parallel()
		secret := t.Name()
		otherSecret := secret + "-other"

		mac := hmac.New(sha256.New, []byte(otherSecret))
		_, err := mac.Write([]byte(testEncodedHeader + "." + testEncodedBody))
		assert.NoError(t, err)

		signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		valid, err := verifyHash(testEncodedHeader, testEncodedBody, signature, secret, sha256.New)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("when the signature is not valid base64 it should return an error", func(t *testing.T) {
		t.Parallel()
		secret := t.Name()

		valid, err := verifyHash(testEncodedHeader, testEncodedBody, "!not-base64!", secret, sha256.New)
		assert.ErrorPart(t, err, "failed to decode hash")
		assert.False(t, valid)
	})

	t.Run("when the provider is nil it should return an error", func(t *testing.T) {
		t.Parallel()
		secret := t.Name()

		valid, err := verifyHash(testEncodedHeader, testEncodedBody, "", secret, nil)
		assert.ErrorPart(t, err, "hash provider cannot be nil")
		assert.False(t, valid)
	})

	t.Run("when the hash provider returns an error on write it should return an error", func(t *testing.T) {
		t.Parallel()
		secret := t.Name()

		valid, err := verifyHash(testEncodedHeader, testEncodedBody, "", secret, func() hash.Hash { return &failingHash{} })
		assert.ErrorPart(t, err, "failed to write data to hash")
		assert.False(t, valid)
	})
}
