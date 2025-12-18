package jwt_test

import (
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/jwt"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/timestamp"
)

func TestNewClaimsVerifier_NoOptions_UsesDefaultValues(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier()
	assert.NotNil(t, verifier)
	err := verifier.Verify(&jwt.Claims{})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_NilClaims_ReturnsError(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier()
	err := verifier.Verify(nil)
	assert.ErrorExact(t, err, "claims cannot be nil")
}

func TestClaimsVerifierVerify_ExpectedIssuerSetAndClaimIssuerMissing_ReturnsError(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedIssuer("expected-issuer"))
	err := verifier.Verify(&jwt.Claims{})
	assert.ErrorExact(t, err, "issuer claim is missing")
}

func TestClaimsVerifierVerify_ExpectedIssuerSetAndClaimIssuerMismatch_ReturnsError(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedIssuer("expected-issuer"))
	issuer := "different-issuer"
	err := verifier.Verify(&jwt.Claims{Issuer: &issuer})
	assert.ErrorPart(t, err, "issuer claim mismatch")
	assert.ErrorPart(t, err, "expected-issuer")
	assert.ErrorPart(t, err, "different-issuer")
}

func TestClaimsVerifierVerify_ExpectedIssuerSetAndClaimIssuerMatches_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedIssuer("expected-issuer"))
	issuer := "expected-issuer"
	err := verifier.Verify(&jwt.Claims{Issuer: &issuer})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_ExpectedSubjectSetAndClaimSubjectMissing_ReturnsError(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedSubject("expected-subject"))
	err := verifier.Verify(&jwt.Claims{})
	assert.ErrorExact(t, err, "subject claim is missing")
}

func TestClaimsVerifierVerify_ExpectedSubjectSetAndClaimSubjectMismatch_ReturnsError(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedSubject("expected-subject"))
	subject := "different-subject"
	err := verifier.Verify(&jwt.Claims{Subject: &subject})
	assert.ErrorPart(t, err, "subject claim mismatch")
	assert.ErrorPart(t, err, "expected-subject")
	assert.ErrorPart(t, err, "different-subject")
}

func TestClaimsVerifierVerify_ExpectedSubjectSetAndClaimSubjectMatches_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedSubject("expected-subject"))
	subject := "expected-subject"
	err := verifier.Verify(&jwt.Claims{Subject: &subject})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_ExpectedAudienceSetAndClaimAudienceMissing_ReturnsError(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedAudience("expected-audience"))
	err := verifier.Verify(&jwt.Claims{})
	assert.ErrorExact(t, err, "audience claim is missing")
}

func TestClaimsVerifierVerify_ExpectedAudienceSetAndClaimAudienceMismatch_ReturnsError(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedAudience("expected-audience"))
	audience := "different-audience"
	err := verifier.Verify(&jwt.Claims{Audience: &audience})
	assert.ErrorPart(t, err, "audience claim mismatch")
	assert.ErrorPart(t, err, "expected-audience")
	assert.ErrorPart(t, err, "different-audience")
}

func TestClaimsVerifierVerify_ExpectedAudienceSetAndClaimAudienceMatches_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedAudience("expected-audience"))
	audience := "expected-audience"
	err := verifier.Verify(&jwt.Claims{Audience: &audience})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_TokenExpired_ReturnsError(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
	expiredTime := timestamp.New(fixedTime.Add(-1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{ExpiresAt: &expiredTime})
	assert.ErrorExact(t, err, "token has expired")
}

func TestClaimsVerifierVerify_TokenNotExpired_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
	futureTime := timestamp.New(fixedTime.Add(1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{ExpiresAt: &futureTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_NoExpiration_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier()
	err := verifier.Verify(&jwt.Claims{})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_ExpirationVerificationDisabled_SkipsExpirationCheck(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithVerifyExpiresAt(false),
	)
	expiredTime := timestamp.New(fixedTime.Add(-1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{ExpiresAt: &expiredTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_TokenExpiredWithinClockSkew_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithClockSkew(2*time.Minute),
	)
	expiredTime := timestamp.New(fixedTime.Add(-1 * time.Minute))
	err := verifier.Verify(&jwt.Claims{ExpiresAt: &expiredTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_TokenNotYetValid_ReturnsError(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
	futureTime := timestamp.New(fixedTime.Add(1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{NotBefore: &futureTime})
	assert.ErrorExact(t, err, "token is not yet valid")
}

func TestClaimsVerifierVerify_TokenValidBasedOnNotBefore_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
	pastTime := timestamp.New(fixedTime.Add(-1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{NotBefore: &pastTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_NoNotBefore_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier()
	err := verifier.Verify(&jwt.Claims{})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_NotBeforeVerificationDisabled_SkipsNotBeforeCheck(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithVerifyNotBefore(false),
	)
	futureTime := timestamp.New(fixedTime.Add(1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{NotBefore: &futureTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_TokenNotYetValidWithinClockSkew_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithClockSkew(2*time.Minute),
	)
	futureTime := timestamp.New(fixedTime.Add(1 * time.Minute))
	err := verifier.Verify(&jwt.Claims{NotBefore: &futureTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_IssuedAtEnabledAndTokenIssuedInFuture_ReturnsError(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithVerifyIssuedAt(true),
	)
	futureTime := timestamp.New(fixedTime.Add(1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{IssuedAt: &futureTime})
	assert.ErrorExact(t, err, "token was issued in the future")
}

func TestClaimsVerifierVerify_IssuedAtEnabledAndTokenIssuedInPast_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithVerifyIssuedAt(true),
	)
	pastTime := timestamp.New(fixedTime.Add(-1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{IssuedAt: &pastTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_IssuedAtEnabledAndNoIssuedAt_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithVerifyIssuedAt(true))
	err := verifier.Verify(&jwt.Claims{})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_IssuedAtDisabled_SkipsIssuedAtCheck(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithVerifyIssuedAt(false),
	)
	futureTime := timestamp.New(fixedTime.Add(1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{IssuedAt: &futureTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_IssuedAtEnabledAndTokenIssuedInFutureWithinClockSkew_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithVerifyIssuedAt(true),
		jwt.WithClockSkew(2*time.Minute),
	)
	futureTime := timestamp.New(fixedTime.Add(1 * time.Minute))
	err := verifier.Verify(&jwt.Claims{IssuedAt: &futureTime})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_AllClaimsValid_Succeeds(t *testing.T) {
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
	issuedAt := timestamp.New(fixedTime.Add(-1 * time.Hour))
	notBefore := timestamp.New(fixedTime.Add(-30 * time.Minute))
	expiresAt := timestamp.New(fixedTime.Add(1 * time.Hour))
	err := verifier.Verify(&jwt.Claims{
		Issuer:    &issuer,
		Subject:   &subject,
		Audience:  &audience,
		IssuedAt:  &issuedAt,
		NotBefore: &notBefore,
		ExpiresAt: &expiresAt,
	})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_TokenExpiresExactlyAtCurrentTime_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
	expiresAt := timestamp.New(fixedTime)
	err := verifier.Verify(&jwt.Claims{ExpiresAt: &expiresAt})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_NotBeforeExactlyAtCurrentTime_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(jwt.WithTimeFunc(func() time.Time { return fixedTime }))
	notBefore := timestamp.New(fixedTime)
	err := verifier.Verify(&jwt.Claims{NotBefore: &notBefore})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_IssuedAtExactlyAtCurrentTime_Succeeds(t *testing.T) {
	t.Parallel()
	fixedTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	verifier := jwt.NewClaimsVerifier(
		jwt.WithTimeFunc(func() time.Time { return fixedTime }),
		jwt.WithVerifyIssuedAt(true),
	)
	issuedAt := timestamp.New(fixedTime)
	err := verifier.Verify(&jwt.Claims{IssuedAt: &issuedAt})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_EmptyStringIssuerMatches_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedIssuer(""))
	issuer := ""
	err := verifier.Verify(&jwt.Claims{Issuer: &issuer})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_EmptyStringSubjectMatches_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedSubject(""))
	subject := ""
	err := verifier.Verify(&jwt.Claims{Subject: &subject})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_EmptyStringAudienceMatches_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier(jwt.WithExpectedAudience(""))
	audience := ""
	err := verifier.Verify(&jwt.Claims{Audience: &audience})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_NoExpectedIssuerWithClaimIssuerSet_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier()
	issuer := "some-issuer"
	err := verifier.Verify(&jwt.Claims{Issuer: &issuer})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_NoExpectedSubjectWithClaimSubjectSet_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier()
	subject := "some-subject"
	err := verifier.Verify(&jwt.Claims{Subject: &subject})
	assert.NoError(t, err)
}

func TestClaimsVerifierVerify_NoExpectedAudienceWithClaimAudienceSet_Succeeds(t *testing.T) {
	t.Parallel()
	verifier := jwt.NewClaimsVerifier()
	audience := "some-audience"
	err := verifier.Verify(&jwt.Claims{Audience: &audience})
	assert.NoError(t, err)
}
