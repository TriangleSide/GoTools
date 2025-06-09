package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// MarshalFunc is used for JSON marshaling and can be overwritten in tests.
var MarshalFunc = json.Marshal

// Header represents the header portion of a JSON Web Token.
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ,omitempty"`
	KeyID     string `json:"kid,omitempty"`
}

// encodeHeader serializes the Header and returns a base64 URL encoded string without padding.
func encodeHeader(h Header) (string, error) {
	data, err := MarshalFunc(h)
	if err != nil {
		return "", fmt.Errorf("json marshal error (%w)", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// decodeHeader decodes a base64 URL encoded header string into a Header struct.
func decodeHeader(encoded string) (*Header, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error (%w)", err)
	}
	var h Header
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("json unmarshal error (%w)", err)
	}
	return &h, nil
}
