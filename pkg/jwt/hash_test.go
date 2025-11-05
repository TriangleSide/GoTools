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

// failingHash implements hash.Hash and always fails on Write.
type failingHash struct {
	v int
}

func (f *failingHash) Write(p []byte) (n int, err error) { return 0, errors.New("write fail") }
func (f *failingHash) Sum(b []byte) []byte               { return []byte{} }
func (f *failingHash) Reset()                            {}
func (f *failingHash) Size() int                         { return 1 }
func (f *failingHash) BlockSize() int                    { return 1 }

func TestHash(t *testing.T) {
	t.Parallel()

	t.Run("it should hash header and body data", func(t *testing.T) {
		t.Parallel()
		hashed, err := hashData("header", "body", "secret", sha256.New)
		assert.NoError(t, err)

		mac := hmac.New(sha256.New, []byte("secret"))
		_, err = mac.Write([]byte("header.body"))
		assert.NoError(t, err)
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		assert.Equals(t, hashed, expected)
	})

	t.Run("it should use a custom hash provider", func(t *testing.T) {
		t.Parallel()
		hashed, err := hashData("header", "body", "secret", sha512.New)
		assert.NoError(t, err)

		mac := hmac.New(sha512.New, []byte("secret"))
		_, err = mac.Write([]byte("header.body"))
		assert.NoError(t, err)
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		assert.Equals(t, hashed, expected)
	})

	t.Run("when the provider is nil it should return an error", func(t *testing.T) {
		t.Parallel()
		h, err := hashData("header", "body", "secret", nil)
		assert.ErrorPart(t, err, "hash provider cannot be nil")
		assert.Equals(t, h, "")
	})

	t.Run("when the hash provider returns an error on write it should return an error", func(t *testing.T) {
		t.Parallel()
		h, err := hashData("header", "body", "secret", func() hash.Hash { return &failingHash{} })
		assert.ErrorPart(t, err, "failed to write data to hash")
		assert.Equals(t, h, "")
	})
}
