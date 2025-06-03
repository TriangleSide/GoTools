package symmetric_test

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	mathrand "math/rand"
	"strconv"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/crypto/symmetric"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestSymmetricEncryption(t *testing.T) {
	t.Parallel()

	newEncryptor := func(t *testing.T) symmetric.Encryptor {
		t.Helper()
		encryptor, err := symmetric.New("encryptionKey" + strconv.Itoa(mathrand.Int()))
		assert.NoError(t, err)
		assert.NotNil(t, encryptor)
		return encryptor
	}

	t.Run("when the cipher provider returns an error it should return an error on creation", func(t *testing.T) {
		t.Parallel()
		encryptor, err := symmetric.New("encryptionKey"+strconv.Itoa(mathrand.Int()), symmetric.WithBlockCypherProvider(func(key []byte) (cipher.Block, error) {
			return nil, errors.New("block cipher provider error")
		}))
		assert.ErrorPart(t, err, "failed to create the block cipher (block cipher provider error)")
		assert.Nil(t, encryptor)
	})

	t.Run("where data of different size is generated it should be able to be encrypted and decrypted", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		for dataSize := 1; dataSize <= 1024; dataSize++ {
			data := make([]byte, dataSize)
			n, err := rand.Read(data)
			assert.NoError(t, err)
			assert.Equals(t, dataSize, n)
			encrypted, err := encryptor.Encrypt(data)
			assert.NoError(t, err)
			assert.NotEquals(t, data, encrypted)
			decrypted, err := encryptor.Decrypt(encrypted)
			assert.NoError(t, err)
			assert.Equals(t, data, decrypted)
		}
	})

	t.Run("when the same data is encrypted it should have different cypher-text", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		const dataSize = 16
		data := make([]byte, dataSize)
		n, err := rand.Read(data)
		assert.NoError(t, err)
		assert.Equals(t, dataSize, n)
		encrypted1, err := encryptor.Encrypt(data)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted1)
		encrypted2, err := encryptor.Encrypt(data)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted2)
		assert.NotEquals(t, encrypted1, encrypted2)
	})

	t.Run("when nil bytes are decrypted it should return an error", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		decrypted, err := encryptor.Decrypt(nil)
		assert.ErrorPart(t, err, "shorter than the minimum length")
		assert.Nil(t, decrypted)
	})

	t.Run("an empty slice of bytes are encrypted and decrypted it should return an empty slice", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		encrypted, err := encryptor.Encrypt([]byte{})
		assert.NoError(t, err)
		decrypted, err := encryptor.Decrypt(encrypted)
		assert.NoError(t, err)
		assert.Equals(t, len(decrypted), 0)
	})

	t.Run("when an encryptor is created with an empty key it should return an error", func(t *testing.T) {
		t.Parallel()
		encryptor, err := symmetric.New("")
		assert.ErrorPart(t, err, "invalid key")
		assert.Nil(t, encryptor)
	})

	t.Run("when the random data func fails it should return an error when encrypting", func(t *testing.T) {
		t.Parallel()
		encryptor, err := symmetric.New("key", symmetric.WithRandomDataFunc(func(buffer []byte) error {
			return errors.New("random data error")
		}))
		assert.NoError(t, err)
		assert.NotNil(t, encryptor)
		cypher, err := encryptor.Encrypt([]byte("test"))
		assert.ErrorPart(t, err, "failed to generate initialization vector (random data error)")
		assert.Nil(t, cypher)
	})
}
