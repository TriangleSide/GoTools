package jwt

import (
	"crypto/rand"
	"io"
)

// encodeOptions holds the configuration for encoding a JWT.
type encodeOptions struct {
	randReader io.Reader
}

// defaultEncodeOptions returns the default options for encoding.
func defaultEncodeOptions() *encodeOptions {
	return &encodeOptions{
		randReader: rand.Reader,
	}
}

// EncodeOption configures the Encode function.
type EncodeOption func(*encodeOptions)

// WithRandomReader sets the random data source for key generation.
// This is for testing purposes to allow deterministic key generation.
// The default is crypto/rand.Reader which should be used in production.
func WithRandomReader(reader io.Reader) EncodeOption {
	return func(opts *encodeOptions) {
		opts.randReader = reader
	}
}
