package symmetric

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

// Encryptor holds the data needed to do AES symmetric encryption.
type Encryptor struct {
	aesBlock cipher.Block
}

// New allocates and configures an Encryptor.
func New(key string) (*Encryptor, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("invalid key")
	}
	hash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher (%s)", err.Error())
	}
	return &Encryptor{
		aesBlock: block,
	}, nil
}

// Encrypt takes a slice of data and returns an encrypted version of that data using the AES algorithm.
// It returns a cypher-text slice of data and an error if any occurs during the encryption process.
func (encryptor *Encryptor) Encrypt(data []byte) ([]byte, error) {
	ciphertext := make([]byte, aes.BlockSize+len(data))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate initialization vector (%s)", err.Error())
	}

	cfb := cipher.NewCFBEncrypter(encryptor.aesBlock, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

// Decrypt performs symmetric decryption on a slice of data using the AES algorithm.
// It returns a plain-text slice of data and an error if any occurs during the decryption process.
func (encryptor *Encryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("cipher-text of len %d is shorter then the minimum length of %d", len(encryptedData), aes.BlockSize)
	}

	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(encryptor.aesBlock, iv)
	cfb.XORKeyStream(encryptedData, encryptedData)

	return encryptedData, nil
}
