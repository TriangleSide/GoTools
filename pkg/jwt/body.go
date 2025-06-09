package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Body represents the body (claims) of a JSON Web Token.
type Body struct {
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	TokenID   string `json:"jti,omitempty"`
}

func encodeBody(b Body) (string, error) {
	data, err := marshalFunc(b)
	if err != nil {
		return "", fmt.Errorf("json marshal error (%w)", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeBody(encoded string) (*Body, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error (%w)", err)
	}
	var b Body
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("json unmarshal error (%w)", err)
	}
	return &b, nil
}
