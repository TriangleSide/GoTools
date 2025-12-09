package jwt_test

import (
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestClaimsVerifier(t *testing.T) {
	t.Parallel()

	t.Run("when created with no options it should use default values", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier()
		assert.NotNil(t, verifier)
		err := verifier.Verify(&jwt.Claims{})
		assert.NoError(t, err)
	})

	t.Run("when claims is nil it should return an error", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier()
		err := verifier.Verify(nil)
		assert.ErrorExact(t, err, "claims cannot be nil")
	})

	t.Run("when expected issuer is set and claim issuer is missing it should return an error", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedIssuer("expected-issuer"))
		err := verifier.Verify(&jwt.Claims{})
		assert.ErrorExact(t, err, "issuer claim is missing")
	})

	t.Run("when expected issuer is set and claim issuer does not match it should return an error", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedIssuer("expected-issuer"))
		issuer := "different-issuer"
		err := verifier.Verify(&jwt.Claims{Issuer: &issuer})
		assert.ErrorPart(t, err, "issuer claim mismatch")
		assert.ErrorPart(t, err, "expected-issuer")
		assert.ErrorPart(t, err, "different-issuer")
	})

	t.Run("when expected issuer is set and claim issuer matches it should succeed", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedIssuer("expected-issuer"))
		issuer := "expected-issuer"
		err := verifier.Verify(&jwt.Claims{Issuer: &issuer})
		assert.NoError(t, err)
	})

	t.Run("when expected subject is set and claim subject is missing it should return an error", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedSubject("expected-subject"))
		err := verifier.Verify(&jwt.Claims{})
		assert.ErrorExact(t, err, "subject claim is missing")
	})

	t.Run("when expected subject is set and claim subject does not match it should return an error", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedSubject("expected-subject"))
		subject := "different-subject"
		err := verifier.Verify(&jwt.Claims{Subject: &subject})
		assert.ErrorPart(t, err, "subject claim mismatch")
		assert.ErrorPart(t, err, "expected-subject")
		assert.ErrorPart(t, err, "different-subject")
	})

	t.Run("when expected subject is set and claim subject matches it should succeed", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedSubject("expected-subject"))
		subject := "expected-subject"
		err := verifier.Verify(&jwt.Claims{Subject: &subject})
		assert.NoError(t, err)
	})

	t.Run("when expected audience is set and claim audience is missing it should return an error", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedAudience("expected-audience"))
		err := verifier.Verify(&jwt.Claims{})
		assert.ErrorExact(t, err, "audience claim is missing")
	})

	t.Run("when expected audience is set and claim audience does not match it should return an error", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedAudience("expected-audience"))
		audience := "different-audience"
		err := verifier.Verify(&jwt.Claims{Audience: &audience})
		assert.ErrorPart(t, err, "audience claim mismatch")
		assert.ErrorPart(t, err, "expected-audience")
		assert.ErrorPart(t, err, "different-audience")
	})

	t.Run("when expected audience is set and claim audience matches it should succeed", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithExpectedAudience("expected-audience"))
		audience := "expected-audience"
		err := verifier.Verify(&jwt.Claims{Audience: &audience})
		assert.NoError(t, err)
	})

	t.Run("when token has expired it should return an error", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
		expiredTime := jwt.NewTimestamp(fixedTime.Add(-1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{ExpiresAt: &expiredTime})
		assert.ErrorExact(t, err, "token has expired")
	})

	t.Run("when token has not expired it should succeed", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
		futureTime := jwt.NewTimestamp(fixedTime.Add(1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{ExpiresAt: &futureTime})
		assert.NoError(t, err)
	})

	t.Run("when token has no expiration it should succeed", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier()
		err := verifier.Verify(&jwt.Claims{})
		assert.NoError(t, err)
	})

	t.Run("when expiration verification is disabled it should skip expiration check", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithVerifyExpiresAt(false),
		)
		expiredTime := jwt.NewTimestamp(fixedTime.Add(-1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{ExpiresAt: &expiredTime})
		assert.NoError(t, err)
	})

	t.Run("when token has expired but within clock skew it should succeed", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithClockSkew(2*time.Minute),
		)
		expiredTime := jwt.NewTimestamp(fixedTime.Add(-1 * time.Minute))
		err := verifier.Verify(&jwt.Claims{ExpiresAt: &expiredTime})
		assert.NoError(t, err)
	})

	t.Run("when token is not yet valid it should return an error", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
		futureTime := jwt.NewTimestamp(fixedTime.Add(1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{NotBefore: &futureTime})
		assert.ErrorExact(t, err, "token is not yet valid")
	})

	t.Run("when token is valid based on not-before it should succeed", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
		pastTime := jwt.NewTimestamp(fixedTime.Add(-1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{NotBefore: &pastTime})
		assert.NoError(t, err)
	})

	t.Run("when token has no not-before it should succeed", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier()
		err := verifier.Verify(&jwt.Claims{})
		assert.NoError(t, err)
	})

	t.Run("when not-before verification is disabled it should skip not-before check", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithVerifyNotBefore(false),
		)
		futureTime := jwt.NewTimestamp(fixedTime.Add(1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{NotBefore: &futureTime})
		assert.NoError(t, err)
	})

	t.Run("when token is not yet valid but within clock skew it should succeed", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithClockSkew(2*time.Minute),
		)
		futureTime := jwt.NewTimestamp(fixedTime.Add(1 * time.Minute))
		err := verifier.Verify(&jwt.Claims{NotBefore: &futureTime})
		assert.NoError(t, err)
	})

	t.Run("when issued-at verification is enabled and token was issued in the future it should return an error", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithVerifyIssuedAt(true),
		)
		futureTime := jwt.NewTimestamp(fixedTime.Add(1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{IssuedAt: &futureTime})
		assert.ErrorExact(t, err, "token was issued in the future")
	})

	t.Run("when issued-at verification is enabled and token was issued in the past it should succeed", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithVerifyIssuedAt(true),
		)
		pastTime := jwt.NewTimestamp(fixedTime.Add(-1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{IssuedAt: &pastTime})
		assert.NoError(t, err)
	})

	t.Run("when issued-at verification is enabled and token has no issued-at it should succeed", func(t *testing.T) {
		t.Parallel()
		verifier := jwt.NewClaimsVerifier(jwt.WithVerifyIssuedAt(true))
		err := verifier.Verify(&jwt.Claims{})
		assert.NoError(t, err)
	})

	t.Run("when issued-at verification is disabled it should skip issued-at check", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithVerifyIssuedAt(false),
		)
		futureTime := jwt.NewTimestamp(fixedTime.Add(1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{IssuedAt: &futureTime})
		assert.NoError(t, err)
	})

	t.Run("when issued-at verification is enabled and token was issued in the future but within clock skew it should succeed", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithVerifyIssuedAt(true),
			jwt.WithClockSkew(2*time.Minute),
		)
		futureTime := jwt.NewTimestamp(fixedTime.Add(1 * time.Minute))
		err := verifier.Verify(&jwt.Claims{IssuedAt: &futureTime})
		assert.NoError(t, err)
	})

	t.Run("when all claims are valid it should succeed", func(t *testing.T) {
		t.Parallel()
		fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		verifier := jwt.NewClaimsVerifier(
			jwt.WithExpectedIssuer("my-issuer"),
			jwt.WithExpectedSubject("my-subject"),
			jwt.WithExpectedAudience("my-audience"),
			jwt.WithTimeFunc(func() time.Time { return fixedTime }),
			jwt.WithVerifyIssuedAt(true),
		)
		issuer := "my-issuer"
		subject := "my-subject"
		audience := "my-audience"
		issuedAt := jwt.NewTimestamp(fixedTime.Add(-1 * time.Hour))
		notBefore := jwt.NewTimestamp(fixedTime.Add(-30 * time.Minute))
		expiresAt := jwt.NewTimestamp(fixedTime.Add(1 * time.Hour))
		err := verifier.Verify(&jwt.Claims{
			Issuer:    &issuer,
			Subject:   &subject,
			Audience:  &audience,
			IssuedAt:  &issuedAt,
			NotBefore: &notBefore,
			ExpiresAt: &expiresAt,
		})
		assert.NoError(t, err)
	})
}
