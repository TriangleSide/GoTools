package jwt

import (
	"crypto/hmac"
	"encoding/base64"
	"fmt"
	"hash"
)

// hashData returns a base64 URL encoded HMAC string from the header and body using the provided hash provider.
func hashData(header Header, body Body, secret string, provider func() hash.Hash) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("hash provider cannot be nil")
	}

	encodedHeader, err := encodeHeader(header)
	if err != nil {
		return "", fmt.Errorf("failed to encode header (%w)", err)
	}

	encodedBody, err := encodeBody(body)
	if err != nil {
		return "", fmt.Errorf("failed to encode body (%w)", err)
	}

	mac := hmac.New(provider, []byte(secret))
	if _, err := mac.Write([]byte(encodedHeader + "." + encodedBody)); err != nil {
		return "", fmt.Errorf("failed to write data to hash (%w)", err)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// verifyHash compares the provided signature against the calculated hash for the header and body using the supplied provider.
func verifyHash(header Header, body Body, secret, signature string, provider func() hash.Hash) (bool, error) {
	if provider == nil {
		return false, fmt.Errorf("hash provider cannot be nil")
	}

	encodedHeader, err := encodeHeader(header)
	if err != nil {
		return false, fmt.Errorf("failed to encode header (%w)", err)
	}

	encodedBody, err := encodeBody(body)
	if err != nil {
		return false, fmt.Errorf("failed to encode body (%w)", err)
	}

	mac := hmac.New(provider, []byte(secret))
	if _, err := mac.Write([]byte(encodedHeader + "." + encodedBody)); err != nil {
		return false, fmt.Errorf("failed to write data to hash (%w)", err)
	}

	sig, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash (%w)", err)
	}

	return hmac.Equal(mac.Sum(nil), sig), nil
}
