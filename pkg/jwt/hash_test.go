package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"hash"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

// failingHash implements hash.Hash and always fails on Write.
type failingHash struct {
	v int
}

func (f *failingHash) Write(p []byte) (n int, err error) { return 0, errors.New("write fail") }
func (f *failingHash) Sum(b []byte) []byte               { return []byte{} }
func (f *failingHash) Reset()                            {}
func (f *failingHash) Size() int                         { return 1 }
func (f *failingHash) BlockSize() int                    { return 1 }

func TestHash(t *testing.T) {
	t.Run("it should hash header and body data", func(t *testing.T) {
		t.Parallel()
		header := Header{Algorithm: "HS256", Type: "JWT"}
		body := Body{Subject: "user"}

		hashed, err := hashData(header, body, "secret", sha256.New)
		assert.NoError(t, err)

		encodedHeader, err := encodeHeader(header)
		assert.NoError(t, err)

		encodedBody, err := encodeBody(body)
		assert.NoError(t, err)

		mac := hmac.New(sha256.New, []byte("secret"))
		_, err = mac.Write([]byte(encodedHeader + "." + encodedBody))
		assert.NoError(t, err)
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		assert.Equals(t, hashed, expected)
	})

	t.Run("it should use a custom hash provider", func(t *testing.T) {
		t.Parallel()
		header := Header{Algorithm: "HS512", Type: "JWT"}
		body := Body{Subject: "admin"}

		hashed, err := hashData(header, body, "secret", sha512.New)
		assert.NoError(t, err)

		encodedHeader, err := encodeHeader(header)
		assert.NoError(t, err)

		encodedBody, err := encodeBody(body)
		assert.NoError(t, err)

		mac := hmac.New(sha512.New, []byte("secret"))
		_, err = mac.Write([]byte(encodedHeader + "." + encodedBody))
		assert.NoError(t, err)
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		assert.Equals(t, hashed, expected)
	})

	t.Run("when the provider is nil it should return an error", func(t *testing.T) {
		t.Parallel()
		h, err := hashData(Header{}, Body{}, "secret", nil)
		assert.ErrorPart(t, err, "hash provider cannot be nil")
		assert.Equals(t, h, "")
	})

	t.Run("when the hash provider returns an error on write it should return an error", func(t *testing.T) {
		t.Parallel()
		h, err := hashData(Header{}, Body{}, "secret", func() hash.Hash { return &failingHash{} })
		assert.ErrorPart(t, err, "failed to write data to hash")
		assert.Equals(t, h, "")
	})
}

func TestHashEncodingFailures(t *testing.T) {
	originalMarshal := marshalFunc
	defer func() {
		marshalFunc = originalMarshal
	}()

	t.Run("when encoding the header fails it should return an error", func(t *testing.T) {
		marshalFunc = func(v any) ([]byte, error) {
			return nil, errors.New("marshal fail")
		}

		hashed, err := hashData(Header{}, Body{}, "secret", sha256.New)
		assert.ErrorPart(t, err, "failed to encode header")
		assert.Equals(t, hashed, "")
	})

	t.Run("when encoding the body fails it should return an error", func(t *testing.T) {
		marshalFunc = originalMarshal
		callCount := 0
		marshalFunc = func(v any) ([]byte, error) {
			callCount++
			if callCount == 1 {
				return originalMarshal(v)
			}
			return nil, errors.New("marshal fail")
		}

		hashed, err := hashData(Header{}, Body{}, "secret", sha256.New)
		assert.ErrorPart(t, err, "failed to encode body")
		assert.Equals(t, hashed, "")
	})
}

func TestVerifyHash(t *testing.T) {
	t.Run("it should verify matching hashes", func(t *testing.T) {
		t.Parallel()
		header := Header{Algorithm: "HS256", Type: "JWT"}
		body := Body{Subject: "user"}

		encodedHeader, err := encodeHeader(header)
		assert.NoError(t, err)

		encodedBody, err := encodeBody(body)
		assert.NoError(t, err)

		mac := hmac.New(sha256.New, []byte("secret"))
		_, err = mac.Write([]byte(encodedHeader + "." + encodedBody))
		assert.NoError(t, err)

		signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		valid, err := verifyHash(header, body, "secret", signature, sha256.New)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("it should return false when the signature does not match", func(t *testing.T) {
		t.Parallel()
		header := Header{Algorithm: "HS256"}
		body := Body{Subject: "user"}

		encodedHeader, err := encodeHeader(header)
		assert.NoError(t, err)

		encodedBody, err := encodeBody(body)
		assert.NoError(t, err)

		mac := hmac.New(sha256.New, []byte("other-secret"))
		_, err = mac.Write([]byte(encodedHeader + "." + encodedBody))
		assert.NoError(t, err)

		signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		valid, err := verifyHash(header, body, "secret", signature, sha256.New)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("when the signature is not valid base64 it should return an error", func(t *testing.T) {
		t.Parallel()
		header := Header{Algorithm: "HS256"}
		body := Body{}

		valid, err := verifyHash(header, body, "secret", "!not-base64!", sha256.New)
		assert.ErrorPart(t, err, "failed to decode hash")
		assert.False(t, valid)
	})

	t.Run("when the provider is nil it should return an error", func(t *testing.T) {
		t.Parallel()
		valid, err := verifyHash(Header{}, Body{}, "secret", "", nil)
		assert.ErrorPart(t, err, "hash provider cannot be nil")
		assert.False(t, valid)
	})

	t.Run("when the hash provider returns an error on write it should return an error", func(t *testing.T) {
		t.Parallel()
		valid, err := verifyHash(Header{}, Body{}, "secret", "", func() hash.Hash { return &failingHash{} })
		assert.ErrorPart(t, err, "failed to write data to hash")
		assert.False(t, valid)
	})
}

func TestVerifyHashEncodingFailures(t *testing.T) {
	originalMarshal := marshalFunc
	defer func() {
		marshalFunc = originalMarshal
	}()

	t.Run("when encoding the header fails it should return an error", func(t *testing.T) {
		marshalFunc = func(v any) ([]byte, error) {
			return nil, errors.New("marshal fail")
		}

		valid, err := verifyHash(Header{}, Body{}, "secret", "", sha256.New)
		assert.ErrorPart(t, err, "failed to encode header")
		assert.False(t, valid)
	})

	t.Run("when encoding the body fails it should return an error", func(t *testing.T) {
		marshalFunc = originalMarshal
		callCount := 0
		marshalFunc = func(v any) ([]byte, error) {
			callCount++
			if callCount == 1 {
				return originalMarshal(v)
			}
			return nil, errors.New("marshal fail")
		}

		valid, err := verifyHash(Header{}, Body{}, "secret", "", sha256.New)
		assert.ErrorPart(t, err, "failed to encode body")
		assert.False(t, valid)
	})
}
