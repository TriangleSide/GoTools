package jwt

import (
	"crypto/rand"
	"io"
)

// encodeOptions holds configuration options for JWT encoding.
type encodeOptions struct {
	randReader io.Reader
}

// EncodeOption configures JWT encoding behavior.
type EncodeOption func(*encodeOptions)

// defaultEncodeOptions returns the default encoding options.
func defaultEncodeOptions() *encodeOptions {
	return &encodeOptions{
		randReader: rand.Reader,
	}
}

// WithRandomReader configures the random data source for key generation.
// This is primarily intended for testing to allow deterministic key generation.
// The default is crypto/rand.Reader which should be used in production.
func WithRandomReader(reader io.Reader) EncodeOption {
	return func(opts *encodeOptions) {
		opts.randReader = reader
	}
}
