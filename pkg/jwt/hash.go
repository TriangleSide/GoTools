package jwt

import (
	"crypto/hmac"
	"encoding/base64"
	"fmt"
	"hash"
)

// hashData returns a base64 URL encoded HMAC string from the header and body using the provided hash provider.
func hashData(header, body, secret string, provider func() hash.Hash) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("hash provider cannot be nil")
	}
	mac := hmac.New(provider, []byte(secret))
	if _, err := mac.Write([]byte(header + "." + body)); err != nil {
		return "", fmt.Errorf("failed to write data to hash (%w)", err)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
