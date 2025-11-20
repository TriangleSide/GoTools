package jwt

// SignatureAlgorithm is the name of the algorithm used to sign the JWT.
// This name is encoded into the JWT header.
type SignatureAlgorithm string

// signatureProvider are the functions used to sign and verify JWTs that all hashing algorithms must implement.
type signatureProvider interface {
	Sign(data []byte, key []byte) ([]byte, error)
	Verify(data []byte, signature []byte, key []byte) (bool, error)
}

var (
	// signatureProviders maps SignatureAlgorithm values to their corresponding signatureProvider implementations.
	signatureProviders = map[SignatureAlgorithm]signatureProvider{
		HS512: hmacSHA512Provider{},
		EdDSA: eddsaProvider{},
	}
)
