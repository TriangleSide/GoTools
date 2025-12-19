package jwt

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// SignatureAlgorithm is the name of the algorithm used to sign the JWT.
// This name is encoded into the JWT header.
type SignatureAlgorithm string

// PublicKey represents a public key used for verifying JWT signatures.
type PublicKey []byte

// PrivateKey represents a private key used for signing JWTs.
type PrivateKey []byte

// signatureProvider are the functions used to sign and verify JWTs that all hashing algorithms must implement.
type signatureProvider interface {
	KeyGen(randReader io.Reader) (PublicKey, PrivateKey, error)
	Sign(data []byte, key PrivateKey) ([]byte, error)
	Verify(data []byte, signature []byte, key PublicKey) (bool, error)
}

var (
	// signatureProviders maps SignatureAlgorithm values to their corresponding signatureProvider implementations.
	// Private secret algorithms must not be added because of the algorithm-confusion vulnerability.
	//
	// The JWT algorithm-confusion vulnerability is when:
	//  1. System uses EdDSA (asymmetric) for signing tokens with a private key.
	//  2. Attacker intercepts a valid token signed with EdDSA.
	//  3. Attacker modifies the header to change algorithm from EdDSA to HS512
	//  4. Attacker uses the PUBLIC key (which is public) as the HMAC-SHA512 secret
	//  5. Attacker can now forge any token because Decode will:
	//     - Read HS512 from the modified header
	//     - Select the HMAC provider
	//     - Verify using HMAC with the public key as the secret
	signatureProviders = map[SignatureAlgorithm]signatureProvider{
		EdDSA: eddsaProvider{},
	}
)

// keyGen generates a new cryptographically secure signing key pair and key ID for the specified algorithm.
// The key ID is derived from the SHA-256 hash of the public key.
func keyGen(provider signatureProvider, randReader io.Reader) (PublicKey, PrivateKey, string, error) {
	publicKey, privateKey, err := provider.KeyGen(randReader)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to generate the key for the JWT: %w", err)
	}
	keyHash := sha256.Sum256(publicKey)
	keyID := base64.RawURLEncoding.EncodeToString(keyHash[:])
	return publicKey, privateKey, keyID, nil
}
