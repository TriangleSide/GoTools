package validation_test

import (
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/ptr"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
	"github.com/TriangleSide/go-toolkit/pkg/validation"
)

func TestLenValidator_ExactLength_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "len=5")
	assert.NoError(t, err)
}

func TestLenValidator_IncorrectLength_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "len=3")
	assert.ErrorPart(t, err, "length 5 must be exactly 3")
}

func TestLenValidator_PointerToStringWithCorrectLength_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("hello"), "len=5")
	assert.NoError(t, err)
}

func TestLenValidator_PointerToStringWithIncorrectLength_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("hello"), "len=3")
	assert.ErrorPart(t, err, "length 5 must be exactly 3")
}

func TestLenValidator_NilPointer_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var((*string)(nil), "len=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestLenValidator_EmptyStringWithZeroLength_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "len=0")
	assert.NoError(t, err)
}

func TestLenValidator_EmptyStringWithNonZeroLength_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "len=1")
	assert.ErrorPart(t, err, "length 0 must be exactly 1")
}

func TestLenValidator_InvalidParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "len=abc")
	assert.ErrorPart(t, err, "invalid instruction 'abc' for len")
}

func TestLenValidator_NonStringValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(12345, "len=5")
	assert.ErrorPart(t, err, "value must be a string for the len validator")
}

func TestLenValidator_InterfaceContainingString_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(any("hello"), "len=5")
	assert.NoError(t, err)
}

func TestLenValidator_InterfaceContainingNonString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any(12345), "len=5")
	assert.ErrorPart(t, err, "value must be a string for the len validator")
}

func TestLenValidator_EmptyParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "len=")
	assert.ErrorPart(t, err, "invalid instruction '' for len")
}

func TestLenValidator_UnicodeString_ChecksByteLength(t *testing.T) {
	t.Parallel()
	err := validation.Var("héllo", "len=5")
	assert.ErrorPart(t, err, "length 6 must be exactly 5")
}

func TestLenValidator_NegativeParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "len=-1")
	assert.ErrorPart(t, err, "the length parameter can't be negative")
}

func TestLenValidator_PointerToEmptyStringWithZeroLength_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(""), "len=0")
	assert.NoError(t, err)
}

func TestMinValidator_LengthEqualsMinimum_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "min=5")
	assert.NoError(t, err)
}

func TestMinValidator_LengthGreaterThanMinimum_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "min=3")
	assert.NoError(t, err)
}

func TestMinValidator_LengthLessThanMinimum_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "min=6")
	assert.ErrorPart(t, err, "length 5 must be at least 6")
}

func TestMinValidator_ZeroLength_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "min=0")
	assert.NoError(t, err)
}

func TestMinValidator_NegativeParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "min=-1")
	assert.ErrorPart(t, err, "the length parameter can't be negative")
}

func TestMinValidator_InvalidParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "min=abc")
	assert.ErrorPart(t, err, "invalid instruction 'abc' for min")
}

func TestMinValidator_EmptyParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "min=")
	assert.ErrorPart(t, err, "invalid instruction '' for min")
}

func TestMinValidator_NilPointer_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var((*string)(nil), "min=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestMinValidator_NonStringValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(12345, "min=5")
	assert.ErrorPart(t, err, "value must be a string for the min validator")
}

func TestMinValidator_PointerToString_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("hello"), "min=3")
	assert.NoError(t, err)
}

func TestMaxValidator_LengthEqualsMaximum_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "max=5")
	assert.NoError(t, err)
}

func TestMaxValidator_LengthLessThanMaximum_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "max=10")
	assert.NoError(t, err)
}

func TestMaxValidator_LengthGreaterThanMaximum_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "max=4")
	assert.ErrorPart(t, err, "length 5 must be at most 4")
}

func TestMaxValidator_NegativeParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "max=-1")
	assert.ErrorPart(t, err, "the length parameter can't be negative")
}

func TestMaxValidator_InvalidParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "max=abc")
	assert.ErrorPart(t, err, "invalid instruction 'abc' for max")
}

func TestMaxValidator_EmptyParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "max=")
	assert.ErrorPart(t, err, "invalid instruction '' for max")
}

func TestMaxValidator_NilPointer_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var((*string)(nil), "max=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestMaxValidator_NonStringValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(12345, "max=5")
	assert.ErrorPart(t, err, "value must be a string for the max validator")
}

func TestMaxValidator_PointerToString_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("hello"), "max=10")
	assert.NoError(t, err)
}

func TestMaxValidator_ZeroLengthOnEmptyString_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "max=0")
	assert.NoError(t, err)
}

func TestMaxValidator_ZeroLengthOnNonEmptyString_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("a", "max=0")
	assert.ErrorPart(t, err, "length 1 must be at most 0")
}
