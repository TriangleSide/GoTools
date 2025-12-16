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
	blockCipherProvider func(key []byte) (cipher.Block, error)
	randomDataFunc      func(buffer []byte) error
}

// Option is optional configuration of the encryptor.
type Option func(*config)

// WithBlockCipherProvider overwrites the provider for the block cipher.
func WithBlockCipherProvider(provider func(key []byte) (cipher.Block, error)) Option {
	return func(c *config) {
		c.blockCipherProvider = provider
	}
}

// WithRandomDataFunc overwrites the random data function.
func WithRandomDataFunc(randomDataFunc func(buffer []byte) error) Option {
	return func(c *config) {
		c.randomDataFunc = randomDataFunc
	}
}

// Cipher provides AES-GCM encryption and decryption.
type Cipher struct {
	aead           cipher.AEAD
	randomDataFunc func(buffer []byte) error
}

// New allocates and configures a Cipher.
func New(key string, opts ...Option) (*Cipher, error) {
	cfg := &config{
		blockCipherProvider: aes.NewCipher,
		randomDataFunc: func(buffer []byte) error {
			if _, err := io.ReadFull(rand.Reader, buffer); err != nil {
				return fmt.Errorf("failed to read random data (%w)", err)
			}
			return nil
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if len(key) == 0 {
		return nil, errors.New("invalid key")
	}
	hash := sha256.Sum256([]byte(key))

	block, err := cfg.blockCipherProvider(hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create the block cipher (%w)", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to configure AEAD mode (%w)", err)
	}

	return &Cipher{
		aead:           aead,
		randomDataFunc: cfg.randomDataFunc,
	}, nil
}

// Encrypt takes a slice of data and returns an encrypted version using Cipher-GCM with a unique nonce.
// It returns the nonce-prefixed ciphertext and an error if any occurs during the encryption process.
func (cipher *Cipher) Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, cipher.aead.NonceSize())
	if err := cipher.randomDataFunc(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce (%w)", err)
	}

	return cipher.aead.Seal(nonce, nonce, data, nil), nil
}

// Decrypt performs Cipher-GCM decryption on a nonce-prefixed ciphertext.
// It returns the recovered plaintext and an error if any occurs during the decryption process.
func (cipher *Cipher) Decrypt(encryptedData []byte) ([]byte, error) {
	nonceSize := cipher.aead.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf(
			"cipher-text of len %d is shorter than the minimum length of %d", len(encryptedData), nonceSize)
	}

	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	plaintext, err := cipher.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt cipher-text (%w)", err)
	}

	return plaintext, nil
}
