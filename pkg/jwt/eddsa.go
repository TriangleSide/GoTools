package jwt

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
)

const (
	// EdDSA represents the EdDSA (Ed25519) signature algorithm.
	EdDSA SignatureAlgorithm = "EdDSA"
)

// eddsaProvider implements the signatureProvider interface using ed25519.
type eddsaProvider struct{}

// KeyGen generates a new ed25519 key pair using the provided random reader.
func (e eddsaProvider) KeyGen(randReader io.Reader) (PublicKey, PrivateKey, error) {
	pubKey, privKey, err := ed25519.GenerateKey(randReader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ed25519 key: %w", err)
	}
	return PublicKey(pubKey), PrivateKey(privKey), nil
}

// Sign signs the data using ed25519 with the provided private key.
func (e eddsaProvider) Sign(data []byte, key PrivateKey) ([]byte, error) {
	privateKey, err := deriveEdDSAPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to use private key: %w", err)
	}
	signature := ed25519.Sign(privateKey, data)
	return signature, nil
}

// Verify verifies the ed25519 signature of the data using the provided key.
func (e eddsaProvider) Verify(data []byte, signature []byte, key PublicKey) (bool, error) {
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
func deriveEdDSAPrivateKey(key PrivateKey) (ed25519.PrivateKey, error) {
	if len(key) != ed25519.PrivateKeySize {
		return nil, errors.New("eddsa private key must be 64 bytes")
	}
	return ed25519.PrivateKey(key), nil
}

// deriveEdDSAPublicKey validates and returns an ed25519 public key from the supplied bytes.
func deriveEdDSAPublicKey(key PublicKey) (ed25519.PublicKey, error) {
	if len(key) != ed25519.PublicKeySize {
		return nil, errors.New("eddsa public key must be 32 bytes")
	}
	return ed25519.PublicKey(key), nil
}
