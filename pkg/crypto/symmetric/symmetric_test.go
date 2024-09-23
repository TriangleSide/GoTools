package symmetric_test

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	mathrand "math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/crypto/symmetric"
)

func TestSymmetricEncryption(t *testing.T) {
	t.Parallel()

	newEncryptor := func(t *testing.T) symmetric.Encryptor {
		encryptor, err := symmetric.New("encryptionKey" + strconv.Itoa(mathrand.Int()))
		if err != nil {
			t.Fatalf("should not return an error")
		}
		if encryptor == nil {
			t.Fatalf("the encryptor should not be nil")
		}
		return encryptor
	}

	t.Run("when the cipher provider returns an error it should return an error on creation", func(t *testing.T) {
		t.Parallel()
		encryptor, err := symmetric.New("encryptionKey"+strconv.Itoa(mathrand.Int()), symmetric.WithBlockCypherProvider(func(key []byte) (cipher.Block, error) {
			return nil, errors.New("block cipher provider error")
		}))
		if err == nil {
			t.Fatalf("should return an error")
		}
		if encryptor != nil {
			t.Fatalf("the encryptor should be nil")
		}
		if !strings.Contains(err.Error(), "failed to create the block cipher (block cipher provider error)") {
			t.Fatalf("error message is not correct (%s)", err.Error())
		}
	})

	t.Run("where data of different size is generated it should be able to be encrypted and decrypted", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		for dataSize := 1; dataSize <= 1024; dataSize++ {
			data := make([]byte, dataSize)
			n, err := rand.Read(data)
			if err != nil {
				t.Fatalf("rand returned an error (%s)", err.Error())
			}
			if n != dataSize {
				t.Fatalf("failed to read %d bytes from rand", n)
			}
			encrypted, err := encryptor.Encrypt(data)
			if err != nil {
				t.Fatalf("encryption returned an error (%s)", err.Error())
			}
			if reflect.DeepEqual(data, encrypted) {
				t.Fatalf("encrypted and decrypted are the same")
			}
			decrypted, err := encryptor.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("decryption returned an error (%s)", err.Error())
			}
			if !reflect.DeepEqual(data, decrypted) {
				t.Fatalf("decrypted data and the original should be the same")
			}
		}
	})

	t.Run("when the same data is encrypted it should have different cypher-text", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		const dataSize = 16
		data := make([]byte, dataSize)
		n, err := rand.Read(data)
		if err != nil {
			t.Fatalf("rand returned an error (%s)", err.Error())
		}
		if n != dataSize {
			t.Fatalf("failed to read %d bytes from rand", n)
		}
		encrypted1, err := encryptor.Encrypt(data)
		if err != nil {
			t.Fatalf("encryption returned an error (%s)", err.Error())
		}
		encrypted2, err := encryptor.Encrypt(data)
		if err != nil {
			t.Fatalf("encryption returned an error (%s)", err.Error())
		}
		if reflect.DeepEqual(encrypted1, encrypted2) {
			t.Fatalf("cipher texts are the same but should not be")
		}
	})

	t.Run("when nil bytes are decrypted it should return an error", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		decrypted, err := encryptor.Decrypt(nil)
		if err == nil {
			t.Fatalf("should return an error")
		}
		if decrypted != nil {
			t.Fatalf("decrypted data should be nil")
		}
		if !strings.Contains(err.Error(), "shorter then the minimum length") {
			t.Fatalf("error message is not correct (%s)", err.Error())
		}
	})

	t.Run("an empty slice of bytes are encrypted and decrypted it should return an empty slice", func(t *testing.T) {
		t.Parallel()
		encryptor := newEncryptor(t)
		encrypted, err := encryptor.Encrypt([]byte{})
		if err != nil {
			t.Fatalf("encryption returned an error (%s)", err.Error())
		}
		decrypted, err := encryptor.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("decryption returned an error (%s)", err.Error())
		}
		if len(decrypted) != 0 {
			t.Fatalf("decrypted data should be empty")
		}
	})

	t.Run("when an encryptor is created with an empty key it should return an error", func(t *testing.T) {
		t.Parallel()
		encryptor, err := symmetric.New("")
		if err == nil {
			t.Fatalf("should return an error")
		}
		if encryptor != nil {
			t.Fatalf("the encryptor should be nil")
		}
		if !strings.Contains(err.Error(), "invalid key") {
			t.Fatalf("error message is not correct (%s)", err.Error())
		}
	})

	t.Run("when the random data func fails it should return an error when encrypting", func(t *testing.T) {
		t.Parallel()
		encryptor, err := symmetric.New("key", symmetric.WithRandomDataFunc(func(buffer []byte) error {
			return errors.New("random data error")
		}))
		if err != nil {
			t.Fatalf("should not return an error")
		}
		if encryptor == nil {
			t.Fatalf("the encryptor should not be nil")
		}
		cypher, err := encryptor.Encrypt([]byte("test"))
		if err == nil {
			t.Fatalf("should return an error")
		}
		if cypher != nil {
			t.Fatalf("the cypher should be nil")
		}
		if !strings.Contains(err.Error(), "failed to generate initialization vector (random data error)") {
			t.Fatalf("error message is not correct (%s)", err.Error())
		}
	})

}
