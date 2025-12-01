package jwt

import (
	"crypto/hmac"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
)

const (
	// HS512 represents the HMAC-SHA512 hash algorithm.
	HS512 SignatureAlgorithm = "HS512"
)

// signHMAC is a helper function to sign data using HMAC with the specified hash function.
func signHMAC(data []byte, key []byte, hashFunc func() hash.Hash) ([]byte, error) {
	mac := hmac.New(hashFunc, key)
	if _, err := mac.Write(data); err != nil {
		return nil, fmt.Errorf("failed to generate signature (%w)", err)
	}
	return mac.Sum(nil), nil
}

// verifyHMAC is a helper function to verify HMAC signatures with the specified hash function.
func verifyHMAC(data []byte, signature []byte, key []byte, hashFunc func() hash.Hash) (bool, error) {
	mac := hmac.New(hashFunc, key)
	if _, err := mac.Write(data); err != nil {
		return false, fmt.Errorf("failed to generate signature (%w)", err)
	}
	if !hmac.Equal(mac.Sum(nil), signature) {
		return false, errors.New("token signature is invalid")
	}
	return true, nil
}

// hmacSHA512Provider implements the signatureProvider interface for HMAC-SHA512.
type hmacSHA512Provider struct{}

// Sign signs the data using HMAC-SHA512 with the provided key.
func (h hmacSHA512Provider) Sign(data []byte, key []byte) ([]byte, error) {
	return signHMAC(data, key, sha512.New)
}

// Verify verifies the HMAC-SHA512 signature of the data using the provided key.
func (h hmacSHA512Provider) Verify(data []byte, signature []byte, key []byte) (bool, error) {
	return verifyHMAC(data, signature, key, sha512.New)
}
