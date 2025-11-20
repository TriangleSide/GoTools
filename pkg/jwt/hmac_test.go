package jwt //nolint:testpackage

import (
	"errors"
	"hash"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type failingHash struct {
	writeErr error
}

func (f *failingHash) Write([]byte) (int, error) { return 0, f.writeErr }
func (f *failingHash) Sum(b []byte) []byte       { return append(b, byte(0)) }
func (f *failingHash) Reset()                    {}
func (f *failingHash) Size() int                 { return 1 }
func (f *failingHash) BlockSize() int            { return 1 }

func TestHelperFunctionsForHMAC(t *testing.T) {
	t.Parallel()

	errWrite := errors.New("hash write failure")

	t.Run("when signing hash write fails it should return error", func(t *testing.T) {
		t.Parallel()
		_, err := signHMAC([]byte("data"), []byte("key"), func() hash.Hash {
			return &failingHash{writeErr: errWrite}
		})
		assert.ErrorPart(t, err, "failed to generate signature")
		assert.ErrorPart(t, err, "hash write failure")
	})

	t.Run("when verifying hash write fails it should return error", func(t *testing.T) {
		t.Parallel()
		match, err := verifyHMAC([]byte("data"), []byte("signature"), []byte("key"), func() hash.Hash {
			return &failingHash{writeErr: errWrite}
		})
		assert.ErrorPart(t, err, "failed to generate signature")
		assert.ErrorPart(t, err, "hash write failure")
		assert.Equals(t, match, false)
	})
}
