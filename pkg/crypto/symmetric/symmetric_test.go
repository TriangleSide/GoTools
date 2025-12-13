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

type invalidSizeBlock struct{}

func (invalidSizeBlock) BlockSize() int { return 8 }

func (invalidSizeBlock) Encrypt(dst, src []byte) { copy(dst, src) }

func (invalidSizeBlock) Decrypt(dst, src []byte) { copy(dst, src) }

func getRandomInt(t *testing.T) int {
	t.Helper()
	randomValueBig, err := rand.Int(rand.Reader, big.NewInt(1000000))
	assert.Nil(t, err)
	return int(randomValueBig.Int64())
}

func newEncryptor(t *testing.T) symmetric.Encryptor {
	t.Helper()
	encryptor, err := symmetric.New("encryptionKey" + strconv.Itoa(getRandomInt(t)))
	assert.NoError(t, err)
	assert.NotNil(t, encryptor)
	return encryptor
}

func TestNew_CustomBlockCipherProvider_ReceivesHashedKey(t *testing.T) {
	t.Parallel()
	key := "encryptionKey" + strconv.Itoa(getRandomInt(t))
	var providedKey []byte
	encryptor, err := symmetric.New(key, symmetric.WithBlockCipherProvider(func(keyBytes []byte) (cipher.Block, error) {
		providedKey = append([]byte(nil), keyBytes...)
		return aes.NewCipher(keyBytes)
	}))
	assert.NoError(t, err)
	assert.NotNil(t, encryptor)
	expectedHash := sha256.Sum256([]byte(key))
	assert.Equals(t, providedKey, expectedHash[:])
}

func TestNew_CipherProviderReturnsError_ReturnsError(t *testing.T) {
	t.Parallel()
	encryptor, err := symmetric.New("encryptionKey"+strconv.Itoa(getRandomInt(t)), symmetric.WithBlockCipherProvider(func([]byte) (cipher.Block, error) {
		return nil, errors.New("block cipher provider error")
	}))
	assert.ErrorPart(t, err, "failed to create the block cipher (block cipher provider error)")
	assert.Nil(t, encryptor)
}

func TestNew_AEADModeCannotBeCreated_ReturnsError(t *testing.T) {
	t.Parallel()
	encryptor, err := symmetric.New("encryptionKey"+strconv.Itoa(getRandomInt(t)), symmetric.WithBlockCipherProvider(func([]byte) (cipher.Block, error) {
		return invalidSizeBlock{}, nil
	}))
	assert.ErrorPart(t, err, "failed to configure AEAD mode")
	assert.Nil(t, encryptor)
}

func TestEncryptDecrypt_VariousDataSizes_SuccessfullyRoundTrips(t *testing.T) {
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
}

func TestNew_CustomRandomDataFunc_ControlsNonce(t *testing.T) {
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
}

func TestEncrypt_SameDataSameInstance_DifferentCipherText(t *testing.T) {
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
}

func TestEncrypt_SameDataDifferentInstances_DifferentCipherText(t *testing.T) {
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
}

func TestDecrypt_NilBytes_ReturnsError(t *testing.T) {
	t.Parallel()
	encryptor := newEncryptor(t)
	decrypted, err := encryptor.Decrypt(nil)
	assert.ErrorPart(t, err, "shorter than the minimum length")
	assert.Nil(t, decrypted)
}

func TestDecrypt_CipherTextShorterThanNonceSize_ReturnsError(t *testing.T) {
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
}

func TestDecrypt_CipherTextExactlyNonceSize_ReturnsError(t *testing.T) {
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
	exactNonceSizeCipher := make([]byte, nonceSize)
	decrypted, err := encryptor.Decrypt(exactNonceSizeCipher)
	assert.ErrorPart(t, err, "failed to decrypt cipher-text")
	assert.Nil(t, decrypted)
}

func TestDecrypt_TamperedCipherText_ReturnsError(t *testing.T) {
	t.Parallel()
	encryptor := newEncryptor(t)
	encrypted, err := encryptor.Encrypt([]byte("super secret data"))
	assert.NoError(t, err)
	assert.True(t, len(encrypted) > 0)
	encrypted[len(encrypted)-1]++
	decrypted, err := encryptor.Decrypt(encrypted)
	assert.ErrorPart(t, err, "failed to decrypt cipher-text")
	assert.Nil(t, decrypted)
}

func TestEncryptDecrypt_EmptySlice_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	encryptor := newEncryptor(t)
	encrypted, err := encryptor.Encrypt([]byte{})
	assert.NoError(t, err)
	decrypted, err := encryptor.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equals(t, len(decrypted), 0)
}

func TestEncryptDecrypt_NilPlaintext_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()
	encryptor := newEncryptor(t)
	encrypted, err := encryptor.Encrypt(nil)
	assert.NoError(t, err)
	decrypted, err := encryptor.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equals(t, len(decrypted), 0)
}

func TestNew_EmptyKey_ReturnsError(t *testing.T) {
	t.Parallel()
	encryptor, err := symmetric.New("")
	assert.ErrorPart(t, err, "invalid key")
	assert.Nil(t, encryptor)
}

func TestEncrypt_RandomDataFuncFails_ReturnsError(t *testing.T) {
	t.Parallel()
	encryptor, err := symmetric.New("key", symmetric.WithRandomDataFunc(func([]byte) error {
		return errors.New("random data error")
	}))
	assert.NoError(t, err)
	assert.NotNil(t, encryptor)
	cypher, err := encryptor.Encrypt([]byte("test"))
	assert.ErrorPart(t, err, "failed to generate nonce (random data error)")
	assert.Nil(t, cypher)
}

func TestEncryptDecrypt_Concurrent_HandlesMultipleGoroutines(t *testing.T) {
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
}
