package symmetric_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
	"strconv"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/crypto/symmetric"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

// invalidSizeBlock is a cipher.Block that uses an unsupported block size.
type invalidSizeBlock struct{}

// BlockSize returns an invalid block size to trigger errors constructing the AEAD mode.
func (invalidSizeBlock) BlockSize() int { return 8 }

// Encrypt is a no-op required to satisfy cipher.Block.
func (invalidSizeBlock) Encrypt(dst, src []byte) { copy(dst, src) }

// Decrypt is a no-op required to satisfy cipher.Block.
func (invalidSizeBlock) Decrypt(dst, src []byte) { copy(dst, src) }

// getRandomInt returns a pseudo-random number to create unique test keys.
func getRandomInt(t *testing.T) int {
	t.Helper()
	randomValueBig, err := rand.Int(rand.Reader, big.NewInt(1000000))
	assert.Nil(t, err)
	return int(randomValueBig.Int64())
}

// TestSymmetricEncryption validates configuration, encryption, and decryption behaviors.
func TestSymmetricEncryption(t *testing.T) {
	t.Parallel()

	newEncryptor := func(t *testing.T) symmetric.Encryptor {
		t.Helper()
		encryptor, err := symmetric.New("encryptionKey" + strconv.Itoa(getRandomInt(t)))
		assert.NoError(t, err)
		assert.NotNil(t, encryptor)
		return encryptor
	}

	t.Run("when a custom block cipher provider is used it should receive the hashed key", func(t *testing.T) {
		t.Parallel()
		key := "encryptionKey" + strconv.Itoa(getRandomInt(t))
		var providedKey []byte
		encryptor, err := symmetric.New(key, symmetric.WithBlockCypherProvider(func(keyBytes []byte) (cipher.Block, error) {
			providedKey = append([]byte(nil), keyBytes...)
			return aes.NewCipher(keyBytes)
		}))
		assert.NoError(t, err)
		assert.NotNil(t, encryptor)
		expectedHash := sha256.Sum256([]byte(key))
		assert.Equals(t, providedKey, expectedHash[:])
	})

	t.Run("when the cipher provider returns an error it should return an error on creation", func(t *testing.T) {
		t.Parallel()
		encryptor, err := symmetric.New("encryptionKey"+strconv.Itoa(getRandomInt(t)), symmetric.WithBlockCypherProvider(func(key []byte) (cipher.Block, error) {
			return nil, errors.New("block cipher provider error")
		}))
		assert.ErrorPart(t, err, "failed to create the block cipher (block cipher provider error)")
		assert.Nil(t, encryptor)
	})

	t.Run("when the AEAD mode cannot be created it should return an error on creation", func(t *testing.T) {
		t.Parallel()
		encryptor, err := symmetric.New("encryptionKey"+strconv.Itoa(getRandomInt(t)), symmetric.WithBlockCypherProvider(func(key []byte) (cipher.Block, error) {
			return invalidSizeBlock{}, nil
		}))
		assert.ErrorPart(t, err, "failed to configure AEAD mode")
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

	t.Run("when a custom random data func is provided it should control the nonce", func(t *testing.T) {
		t.Parallel()
		var generatedNonce []byte
		encryptor, err := symmetric.New("key"+strconv.Itoa(getRandomInt(t)), symmetric.WithRandomDataFunc(func(buffer []byte) error {
			for i := range buffer {
				buffer[i] = byte(i + 1)
			}
			generatedNonce = append([]byte(nil), buffer...)
			return nil
		}))
		assert.NoError(t, err)
		assert.NotNil(t, encryptor)
		ciphertext, err := encryptor.Encrypt([]byte("data"))
		assert.NoError(t, err)
		assert.True(t, len(generatedNonce) > 0)
		assert.Equals(t, generatedNonce, ciphertext[:len(generatedNonce)])
	})

	t.Run("when the same data is encrypted using the same instance it should have different cypher-text", func(t *testing.T) {
		t.Parallel()
		const dataSize = 16
		data := make([]byte, dataSize)
		n, err := rand.Read(data)
		assert.NoError(t, err)
		assert.Equals(t, dataSize, n)
		encryptor := newEncryptor(t)
		encrypted1, err := encryptor.Encrypt(data)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted1)
		encrypted2, err := encryptor.Encrypt(data)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted2)
		assert.NotEquals(t, encrypted1, encrypted2)
	})

	t.Run("when the same data is encrypted using the different instances it should have different cypher-text", func(t *testing.T) {
		t.Parallel()
		const dataSize = 16
		data := make([]byte, dataSize)
		n, err := rand.Read(data)
		assert.NoError(t, err)
		assert.Equals(t, dataSize, n)
		encryptor := newEncryptor(t)
		encrypted1, err := encryptor.Encrypt(data)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted1)
		encryptor = newEncryptor(t)
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

	t.Run("when cipher-text shorter than the nonce size is decrypted it should return an error", func(t *testing.T) {
		t.Parallel()
		var nonceSize int
		encryptor, err := symmetric.New("key"+strconv.Itoa(getRandomInt(t)), symmetric.WithRandomDataFunc(func(buffer []byte) error {
			nonceSize = len(buffer)
			_, readErr := rand.Read(buffer)
			return readErr
		}))
		assert.NoError(t, err)
		assert.NotNil(t, encryptor)
		_, err = encryptor.Encrypt([]byte("prime the nonce size"))
		assert.NoError(t, err)
		assert.True(t, nonceSize > 0)
		shortCipher := make([]byte, nonceSize-1)
		decrypted, err := encryptor.Decrypt(shortCipher)
		assert.ErrorPart(t, err, "shorter than the minimum length")
		assert.ErrorPart(t, err, strconv.Itoa(nonceSize))
		assert.Nil(t, decrypted)
	})

	t.Run("when cipher-text is tampered it should fail to decrypt", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		encrypted, err := encryptor.Encrypt([]byte("super secret data"))
		assert.NoError(t, err)
		assert.True(t, len(encrypted) > 0)
		encrypted[len(encrypted)-1]++
		decrypted, err := encryptor.Decrypt(encrypted)
		assert.ErrorPart(t, err, "failed to decrypt cipher-text")
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
		assert.ErrorPart(t, err, "failed to generate nonce (random data error)")
		assert.Nil(t, cypher)
	})

	t.Run("when encrypting and decrypting concurrently it should handle multiple goroutines", func(t *testing.T) {
		t.Parallel()
		const (
			goroutines = 8
			iterations = 4096
		)
		encryptor := newEncryptor(t)
		var wg sync.WaitGroup
		for range goroutines {
			wg.Go(func() {
				for range iterations {
					data := []byte("payload-" + strconv.Itoa(getRandomInt(t)))
					encrypted, err := encryptor.Encrypt(data)
					assert.NoError(t, err)
					decrypted, err := encryptor.Decrypt(encrypted)
					assert.NoError(t, err)
					assert.Equals(t, data, decrypted)
				}
			})
		}
		wg.Wait()
	})
}
