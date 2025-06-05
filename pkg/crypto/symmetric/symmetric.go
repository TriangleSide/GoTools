package symmetric

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

// config is the configuration for the encryptor.
type config struct {
	blockCypherProvider func(key []byte) (cipher.Block, error)
	randomDataFunc      func(buffer []byte) error
}

// Option is optional configuration of the encryptor.
type Option func(*config)

// WithBlockCypherProvider overwrites the provider for the block cipher.
func WithBlockCypherProvider(provider func(key []byte) (cipher.Block, error)) Option {
	return func(c *config) {
		c.blockCypherProvider = provider
	}
}

// WithRandomDataFunc overwrites the random data function.
func WithRandomDataFunc(randomDataFunc func(buffer []byte) error) Option {
	return func(c *config) {
		c.randomDataFunc = randomDataFunc
	}
}

// Encryptor does symmetric encryption and decryption.
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// aesEncryptor holds the data needed to do AES symmetric encryption.
type aesEncryptor struct {
	aesBlock       cipher.Block
	randomDataFunc func(buffer []byte) error
}

// New allocates and configures an Encryptor.
func New(key string, opts ...Option) (Encryptor, error) {
	cfg := &config{
		blockCypherProvider: aes.NewCipher,
		randomDataFunc: func(buffer []byte) error {
			_, err := io.ReadFull(rand.Reader, buffer)
			return err
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if len(key) == 0 {
		return nil, errors.New("invalid key")
	}
	hash := sha256.Sum256([]byte(key))

	block, err := cfg.blockCypherProvider(hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create the block cipher (%w)", err)
	}

	return &aesEncryptor{
		aesBlock:       block,
		randomDataFunc: cfg.randomDataFunc,
	}, nil
}

// Encrypt takes a slice of data and returns an encrypted version of that data using the AES algorithm.
// It returns a ciphertext slice of data and an error if any occurs during the encryption process.
func (encryptor *aesEncryptor) Encrypt(data []byte) ([]byte, error) {
	ciphertext := make([]byte, aes.BlockSize+len(data))

	iv := ciphertext[:aes.BlockSize]
	if err := encryptor.randomDataFunc(iv); err != nil {
		return nil, fmt.Errorf("failed to generate initialization vector (%w)", err)
	}

	cfb := cipher.NewCFBEncrypter(encryptor.aesBlock, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

// Decrypt performs symmetric decryption on a slice of data using the AES algorithm.
// It returns a plain-text slice of data and an error if any occurs during the decryption process.
func (encryptor *aesEncryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("cipher-text of len %d is shorter than the minimum length of %d", len(encryptedData), aes.BlockSize)
	}

	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(encryptor.aesBlock, iv)
	cfb.XORKeyStream(encryptedData, encryptedData)

	return encryptedData, nil
}
