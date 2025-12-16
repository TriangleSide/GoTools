package jwt

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
)

const (
	// EdDSA represents the EdDSA (Ed25519) signature algorithm.
	EdDSA SignatureAlgorithm = "EdDSA"
)

// eddsaProvider implements the signatureProvider interface using ed25519.
type eddsaProvider struct{}

// KeyGen generates a new ed25519 private key using cryptographically secure random bytes.
func (e eddsaProvider) KeyGen() ([]byte, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key: %w", err)
	}
	return privateKey, nil
}

// Sign signs the data using ed25519 with the provided private key.
func (e eddsaProvider) Sign(data []byte, key []byte) ([]byte, error) {
	privateKey, err := deriveEdDSAPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to use private key: %w", err)
	}
	signature := ed25519.Sign(privateKey, data)
	return signature, nil
}

// Verify verifies the ed25519 signature of the data using the provided key.
func (e eddsaProvider) Verify(data []byte, signature []byte, key []byte) (bool, error) {
	publicKey, err := deriveEdDSAPublicKey(key)
	if err != nil {
		return false, fmt.Errorf("failed to use public key: %w", err)
	}
	if len(signature) != ed25519.SignatureSize {
		return false, errors.New("eddsa signature must be 64 bytes")
	}
	if !ed25519.Verify(publicKey, data, signature) {
		return false, errors.New("token signature is invalid")
	}
	return true, nil
}

// deriveEdDSAPrivateKey validates and returns an ed25519 private key from the supplied bytes.
func deriveEdDSAPrivateKey(key []byte) (ed25519.PrivateKey, error) {
	if len(key) != ed25519.PrivateKeySize {
		return nil, errors.New("eddsa private key must be 64 bytes")
	}
	privateKey := ed25519.PrivateKey(key)
	return privateKey, nil
}

// deriveEdDSAPublicKey builds an ed25519 public key from either a public or private key input.
func deriveEdDSAPublicKey(key []byte) (ed25519.PublicKey, error) {
	switch len(key) {
	case ed25519.PublicKeySize:
		return ed25519.PublicKey(key), nil
	case ed25519.PrivateKeySize:
		privateKey := ed25519.PrivateKey(key)
		publicKey := privateKey.Public().(ed25519.PublicKey)
		return publicKey, nil
	default:
		return nil, errors.New("eddsa key must be 32 or 64 bytes")
	}
}
