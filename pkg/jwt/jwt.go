package jwt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	// tokenSegmentCount represents the number of segments in a JWT string.
	tokenSegmentCount = 3
	// jwtHeaderType is the standard type value for JWT headers.
	jwtHeaderType = "JWT"
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

// Encode converts the provided claims into a signed JWT string using the specified algorithm.
// It generates a new key internally and returns the encoded JWT, the key used, the key ID, and any error.
// The key and key ID must be persisted by the caller for future verification.
func Encode(claims Claims, algorithm SignatureAlgorithm) (string, []byte, string, error) {
	provider, ok := signatureProviders[algorithm]
	if !ok {
		return "", nil, "", errors.New("failed to resolve signature provider")
	}

	key, keyId, err := keyGen(provider)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to generate signing key (%w)", err)
	}

	header := Header{Algorithm: string(algorithm), Type: jwtHeaderType, KeyID: keyId}
	headerJson := marshalToStableJSON(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString([]byte(headerJson))

	bodyJson := marshalToStableJSON(claims)
	encodedBody := base64.RawURLEncoding.EncodeToString([]byte(bodyJson))

	signatureBytes, err := provider.Sign([]byte(encodedHeader+"."+encodedBody), key)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to sign token (%w)", err)
	}

	encodedSignature := base64.RawURLEncoding.EncodeToString(signatureBytes)
	jwt := strings.Join([]string{encodedHeader, encodedBody, encodedSignature}, ".")

	return jwt, key, keyId, nil
}

// KeyProvider is a function type that retrieves the signing key and algorithm based on the provided context and key ID.
type KeyProvider func(ctx context.Context, keyId string) ([]byte, SignatureAlgorithm, error)

// Decode validates the supplied token string using the key and algorithm from the provider and returns the decoded claims.
func Decode(ctx context.Context, token string, keyProvider KeyProvider) (*Claims, error) {
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

	key, algorithm, err := keyProvider(ctx, header.KeyID)
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
