package jwt

import (
	"crypto/hmac"
	"encoding/base64"
	"fmt"
	"hash"
)

// hashData returns a base64 URL encoded HMAC string using the provided hash provider.
func hashData(data, secret string, provider func() hash.Hash) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("hash provider cannot be nil")
	}
	mac := hmac.New(provider, []byte(secret))
	if _, err := mac.Write([]byte(data)); err != nil {
		return "", fmt.Errorf("failed to write data to hash (%w)", err)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// verifyHash checks if the encoded hash matches the data and secret using the given provider.
func verifyHash(encoded, data, secret string, provider func() hash.Hash) (bool, error) {
	computed, err := hashData(data, secret, provider)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(computed), []byte(encoded)), nil
}
