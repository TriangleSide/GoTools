package jwt

import (
	"errors"
	"fmt"
	"time"
)

// ClaimsVerifier validates JWT claims against configured constraints.
type ClaimsVerifier struct {
	expectedIssuer   *string
	expectedSubject  *string
	expectedAudience *string
	clockSkew        time.Duration
	timeFunc         func() time.Time
	verifyExpiresAt  bool
	verifyNotBefore  bool
	verifyIssuedAt   bool
}

// VerifierOption configures a ClaimsVerifier.
type VerifierOption func(*ClaimsVerifier)

// NewClaimsVerifier creates a new ClaimsVerifier with the provided options.
func NewClaimsVerifier(opts ...VerifierOption) *ClaimsVerifier {
	verifier := &ClaimsVerifier{
		clockSkew:       0,
		timeFunc:        time.Now,
		verifyExpiresAt: true,
		verifyNotBefore: true,
		verifyIssuedAt:  false,
	}
	for _, opt := range opts {
		opt(verifier)
	}
	return verifier
}

// WithExpectedIssuer configures the verifier to require the claim's issuer to match the expected value.
func WithExpectedIssuer(issuer string) VerifierOption {
	return func(v *ClaimsVerifier) {
		v.expectedIssuer = &issuer
	}
}

// WithExpectedSubject configures the verifier to require the claim's subject to match the expected value.
func WithExpectedSubject(subject string) VerifierOption {
	return func(v *ClaimsVerifier) {
		v.expectedSubject = &subject
	}
}

// WithExpectedAudience configures the verifier to require the claim's audience to match the expected value.
func WithExpectedAudience(audience string) VerifierOption {
	return func(v *ClaimsVerifier) {
		v.expectedAudience = &audience
	}
}

// WithClockSkew configures the verifier to allow a time tolerance for clock differences between systems.
func WithClockSkew(skew time.Duration) VerifierOption {
	return func(v *ClaimsVerifier) {
		v.clockSkew = skew
	}
}

// WithTimeFunc configures the verifier to use a custom function for obtaining the current time.
func WithTimeFunc(fn func() time.Time) VerifierOption {
	return func(v *ClaimsVerifier) {
		v.timeFunc = fn
	}
}

// WithVerifyExpiresAt configures the verifier to check or skip the expiration time validation.
func WithVerifyExpiresAt(verify bool) VerifierOption {
	return func(v *ClaimsVerifier) {
		v.verifyExpiresAt = verify
	}
}

// WithVerifyNotBefore configures the verifier to check or skip the not-before time validation.
func WithVerifyNotBefore(verify bool) VerifierOption {
	return func(v *ClaimsVerifier) {
		v.verifyNotBefore = verify
	}
}

// WithVerifyIssuedAt configures the verifier to check or skip the issued-at time validation.
func WithVerifyIssuedAt(verify bool) VerifierOption {
	return func(v *ClaimsVerifier) {
		v.verifyIssuedAt = verify
	}
}

// Verify validates the provided claims against the configured constraints.
func (v *ClaimsVerifier) Verify(claims *Claims) error {
	if claims == nil {
		return errors.New("claims cannot be nil")
	}

	if err := v.verifyIssuer(claims); err != nil {
		return err
	}

	if err := v.verifySubject(claims); err != nil {
		return err
	}

	if err := v.verifyAudience(claims); err != nil {
		return err
	}

	now := v.timeFunc()

	if err := v.verifyExpiration(claims, now); err != nil {
		return err
	}

	if err := v.verifyNotBeforeTime(claims, now); err != nil {
		return err
	}

	if err := v.verifyIssuedAtTime(claims, now); err != nil {
		return err
	}

	return nil
}

// verifyIssuer checks the issuer claim if an expected value is configured.
func (v *ClaimsVerifier) verifyIssuer(claims *Claims) error {
	if v.expectedIssuer == nil {
		return nil
	}
	if claims.Issuer == nil {
		return errors.New("issuer claim is missing")
	}
	if *claims.Issuer != *v.expectedIssuer {
		return fmt.Errorf("issuer claim mismatch (expected %q, got %q)", *v.expectedIssuer, *claims.Issuer)
	}
	return nil
}

// verifySubject checks the subject claim if an expected value is configured.
func (v *ClaimsVerifier) verifySubject(claims *Claims) error {
	if v.expectedSubject == nil {
		return nil
	}
	if claims.Subject == nil {
		return errors.New("subject claim is missing")
	}
	if *claims.Subject != *v.expectedSubject {
		return fmt.Errorf("subject claim mismatch (expected %q, got %q)", *v.expectedSubject, *claims.Subject)
	}
	return nil
}

// verifyAudience checks the audience claim if an expected value is configured.
func (v *ClaimsVerifier) verifyAudience(claims *Claims) error {
	if v.expectedAudience == nil {
		return nil
	}
	if claims.Audience == nil {
		return errors.New("audience claim is missing")
	}
	if *claims.Audience != *v.expectedAudience {
		return fmt.Errorf("audience claim mismatch (expected %q, got %q)", *v.expectedAudience, *claims.Audience)
	}
	return nil
}

// verifyExpiration checks that the token has not expired.
func (v *ClaimsVerifier) verifyExpiration(claims *Claims, now time.Time) error {
	if !v.verifyExpiresAt {
		return nil
	}
	if claims.ExpiresAt == nil {
		return nil
	}
	expWithSkew := claims.ExpiresAt.Time().Add(v.clockSkew)
	if now.After(expWithSkew) {
		return errors.New("token has expired")
	}
	return nil
}

// verifyNotBeforeTime checks that the token is valid for use at the current time.
func (v *ClaimsVerifier) verifyNotBeforeTime(claims *Claims, now time.Time) error {
	if !v.verifyNotBefore {
		return nil
	}
	if claims.NotBefore == nil {
		return nil
	}
	nbfWithSkew := claims.NotBefore.Time().Add(-v.clockSkew)
	if now.Before(nbfWithSkew) {
		return errors.New("token is not yet valid")
	}
	return nil
}

// verifyIssuedAtTime checks that the token was not issued in the future.
func (v *ClaimsVerifier) verifyIssuedAtTime(claims *Claims, now time.Time) error {
	if !v.verifyIssuedAt {
		return nil
	}
	if claims.IssuedAt == nil {
		return nil
	}
	iatWithSkew := claims.IssuedAt.Time().Add(-v.clockSkew)
	if now.Before(iatWithSkew) {
		return errors.New("token was issued in the future")
	}
	return nil
}
