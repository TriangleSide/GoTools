package jwt

import (
	"crypto/sha256"
	"crypto/sha512"
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

	t.Run("it should hash and verify data", func(t *testing.T) {
		t.Parallel()
		hashed, err := hashData("payload", "secret", sha256.New)
		assert.NoError(t, err)
		ok, err := verifyHash(hashed, "payload", "secret", sha256.New)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("when hashes do not match it should return false", func(t *testing.T) {
		t.Parallel()
		hashed, err := hashData("payload", "secret", sha256.New)
		assert.NoError(t, err)
		ok, err := verifyHash(hashed+"extra", "payload", "secret", sha256.New)
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("it should use a custom hash provider", func(t *testing.T) {
		t.Parallel()
		hashed, err := hashData("payload", "secret", sha512.New)
		assert.NoError(t, err)
		ok, err := verifyHash(hashed, "payload", "secret", sha512.New)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("when the provider is nil it should return an error", func(t *testing.T) {
		t.Parallel()
		h, err := hashData("payload", "secret", nil)
		assert.ErrorPart(t, err, "hash provider cannot be nil")
		assert.Equals(t, h, "")
	})

	t.Run("when the hash provider returns an error on write it should return an error", func(t *testing.T) {
		t.Parallel()
		h, err := hashData("payload", "secret", func() hash.Hash { return &failingHash{} })
		assert.ErrorPart(t, err, "failed to write data to hash")
		assert.Equals(t, h, "")
	})

	t.Run("when verify uses a failing hash provider it should return an error", func(t *testing.T) {
		t.Parallel()
		ok, err := verifyHash("", "payload", "secret", func() hash.Hash { return &failingHash{} })
		assert.ErrorPart(t, err, "failed to write data to hash")
		assert.False(t, ok)
	})

	t.Run("when verify is called with a nil provider it should return an error", func(t *testing.T) {
		t.Parallel()
		ok, err := verifyHash("", "payload", "secret", nil)
		assert.ErrorPart(t, err, "hash provider cannot be nil")
		assert.False(t, ok)
	})
}
