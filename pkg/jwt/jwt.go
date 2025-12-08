package jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	// tokenSegmentCount represents the number of segments in a JWT string.
	tokenSegmentCount = 3
)

// Header represents the header portion of a JSON Web Token.
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

// Claims represents the claims (body) of a JSON Web Token.
type Claims struct {
	Issuer    *string    `json:"iss"`
	Subject   *string    `json:"sub"`
	Audience  *string    `json:"aud"`
	ExpiresAt *Timestamp `json:"exp"`
	NotBefore *Timestamp `json:"nbf"`
	IssuedAt  *Timestamp `json:"iat"`
	TokenID   *string    `json:"jti"`
}

// Encode converts the provided claims into a signed JWT string using the supplied key and algorithm.
func Encode(claims Claims, key []byte, keyId string, algorithm SignatureAlgorithm) (string, error) {
	header := Header{Algorithm: string(algorithm), Type: "JWT", KeyID: keyId}
	headerJson := marshalToStableJSON(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(headerJson))

	bodyJson := marshalToStableJSON(claims)
	encodedBody := base64.RawURLEncoding.EncodeToString([]byte(bodyJson))

	provider, ok := signatureProviders[SignatureAlgorithm(header.Algorithm)]
	if !ok {
		return "", errors.New("failed to resolve signature provider")
	}
	signatureBytes, err := provider.Sign([]byte(encodedHeader+"."+encodedBody), key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token (%w)", err)
	}
	encodedSignature := base64.RawURLEncoding.EncodeToString(signatureBytes)

	return strings.Join([]string{encodedHeader, encodedBody, encodedSignature}, "."), nil
}

// Decode validates the supplied token string using the key and algorithm from the provider and returns the decoded claims.
func Decode(token string, keyProvider func(keyId string) ([]byte, SignatureAlgorithm, error)) (*Claims, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	parts := strings.Split(token, ".")
	if len(parts) != tokenSegmentCount {
		return nil, errors.New("token must contain header, body, and signature")
	}

	headerJson, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header (%w)", err)
	}
	var header Header
	if err := json.Unmarshal(headerJson, &header); err != nil {
		return nil, fmt.Errorf("json unmarshal error (%w)", err)
	}

	key, algorithm, err := keyProvider(header.KeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key (%w)", err)
	}

	if SignatureAlgorithm(header.Algorithm) != algorithm {
		return nil, errors.New("token algorithm does not match expected algorithm")
	}

	provider, ok := signatureProviders[algorithm]
	if !ok {
		return nil, errors.New("failed to resolve signature provider")
	}
	existingSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature (%w)", err)
	}
	signaturesMatch, err := provider.Verify([]byte(parts[0]+"."+parts[1]), existingSignature, key)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token (%w)", err)
	}
	if !signaturesMatch {
		return nil, errors.New("token signature is invalid")
	}

	bodyJson, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode body (%w)", err)
	}
	var claims Claims
	if err := json.Unmarshal(bodyJson, &claims); err != nil {
		return nil, fmt.Errorf("json unmarshal error (%w)", err)
	}

	return &claims, nil
}
