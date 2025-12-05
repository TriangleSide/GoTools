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

// Option describes a functional option for customizing Encode and Decode behavior.
type Option func(*config)

// WithSignatureAlgorithm overrides the default signature algorithm used for signing and verification.
func WithSignatureAlgorithm(algorithm SignatureAlgorithm) Option {
	return func(cfg *config) {
		cfg.algorithm = algorithm
	}
}

// config contains the runtime configuration for Encode and Decode helpers.
type config struct {
	algorithm SignatureAlgorithm
}

// newConfig builds a config populated with the defaults and supplied options.
func newConfig(opts ...Option) *config {
	cfg := &config{
		algorithm: EdDSA,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Encode converts the provided claims into a signed JWT string using the supplied key and options.
func Encode(claims Claims, key []byte, keyId string, opts ...Option) (string, error) {
	cfg := newConfig(opts...)

	header := Header{Algorithm: string(cfg.algorithm), Type: "JWT", KeyID: keyId}
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

// Decode validates the supplied token string using the secret and returns the decoded claims.
func Decode(token string, keyProvider func(keyId string) ([]byte, error)) (*Claims, error) {
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

	key, err := keyProvider(header.KeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key (%w)", err)
	}

	provider, ok := signatureProviders[SignatureAlgorithm(header.Algorithm)]
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
