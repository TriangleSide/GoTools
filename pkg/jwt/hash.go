package jwt

import (
	"crypto/hmac"
	"encoding/base64"
	"fmt"
	"hash"
)

// hashData returns a base64 URL encoded HMAC string from the provided encoded header and body using the hash provider.
func hashData(encodedHeader, encodedBody, secret string, provider func() hash.Hash) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("hash provider cannot be nil")
	}
	mac := hmac.New(provider, []byte(secret))
	if _, err := mac.Write([]byte(encodedHeader + "." + encodedBody)); err != nil {
		return "", fmt.Errorf("failed to write data to hash (%w)", err)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// verifyHash compares the provided signature against the calculated hash for the encoded header and body using the supplied provider.
func verifyHash(encodedHeader, encodedBody, encodedSignature, secret string, provider func() hash.Hash) (bool, error) {
	if provider == nil {
		return false, fmt.Errorf("hash provider cannot be nil")
	}
  
	mac := hmac.New(provider, []byte(secret))
	if _, err := mac.Write([]byte(encodedHeader + "." + encodedBody)); err != nil {
		return false, fmt.Errorf("failed to write data to hash (%w)", err)
	}

	sig, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash (%w)", err)
	}

	return hmac.Equal(mac.Sum(nil), sig), nil
}
