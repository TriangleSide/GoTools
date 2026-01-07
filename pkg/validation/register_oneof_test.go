package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestOneOf_ValueMatchesAllowedString_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("apple", "oneof=apple banana cherry")
	assert.NoError(t, err)
}

func TestOneOf_SingleAllowedValueMatches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("only", "oneof=only")
	assert.NoError(t, err)
}

func TestOneOf_SingleAllowedValueNotMatches_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("other", "oneof=only")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_ValueNotMatchAllowedStrings_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("orange", "oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_IntegerMatchesAllowedValue_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(42, "oneof=42 100 200")
	assert.NoError(t, err)
}

func TestOneOf_IntegerNotMatchAllowedValues_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(50, "oneof=42 100 200")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_PointerMatchesAllowedValue_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("banana"), "oneof=apple banana cherry")
	assert.NoError(t, err)
}

func TestOneOf_PointerNotMatchAllowedValues_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("grape"), "oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_NilValue_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(any(nil), "oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is nil")
}

func TestOneOf_TypedNilStringPointer_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var((*string)(nil), "oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is nil")
}

func TestOneOf_TypedNilIntPointer_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var((*int)(nil), "oneof=1 2 3")
	assert.ErrorPart(t, err, "value is nil")
}

func TestOneOf_InterfaceMatchesAllowedValue_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(any("cherry"), "oneof=apple banana cherry")
	assert.NoError(t, err)
}

func TestOneOf_InterfaceNotMatchAllowedValues_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(any("pear"), "oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_BooleanTrueMatchesStringRepresentation_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(true, "oneof=true false")
	assert.NoError(t, err)
}

func TestOneOf_BooleanFalseMatchesStringRepresentation_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(false, "oneof=true false")
	assert.NoError(t, err)
}

func TestOneOf_BooleanNotMatchStringRepresentation_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(true, "oneof=yes no")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_EmptyAllowedValuesList_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("anything", "oneof=")
	assert.ErrorPart(t, err, "no parameters provided")
}

func TestOneOf_DiveWithAllElementsMatching_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var([]string{"apple", "banana"}, "dive,oneof=apple banana cherry")
	assert.NoError(t, err)
}

func TestOneOf_DiveWithElementNotMatching_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var([]string{"apple", "orange"}, "dive,oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_SliceWithoutDive_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var([]string{"apple", "banana"}, "oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_EmptyStringWithEmptyParams_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "oneof=")
	assert.ErrorPart(t, err, "no parameters")
}

func TestOneOf_EmptyStringNotAllowed_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_OmitemptyWithEmptyValue_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "omitempty,oneof=apple banana cherry")
	assert.NoError(t, err)
}

func TestOneOf_OmitemptyWithInvalidNonEmptyValue_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("orange", "omitempty,oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_IntegerMatchesStringRepresentation_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(100, "oneof=100 200 300")
	assert.NoError(t, err)
}

func TestOneOf_IntegerNotMatchStringRepresentation_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(150, "oneof=100 200 300")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_NegativeIntegerMatches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(-5, "oneof=-5 -10 -15")
	assert.NoError(t, err)
}

func TestOneOf_NegativeIntegerNotMatches_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(-7, "oneof=-5 -10 -15")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_Int8Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(int8(42), "oneof=42 100")
	assert.NoError(t, err)
}

func TestOneOf_Int8NotMatches_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(int8(50), "oneof=42 100")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_Int16Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(int16(1000), "oneof=1000 2000")
	assert.NoError(t, err)
}

func TestOneOf_Int32Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(int32(100000), "oneof=100000 200000")
	assert.NoError(t, err)
}

func TestOneOf_Int64Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(int64(9223372036854775807), "oneof=9223372036854775807 -9223372036854775808")
	assert.NoError(t, err)
}

func TestOneOf_UintMatches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(42), "oneof=42 100")
	assert.NoError(t, err)
}

func TestOneOf_UintNotMatches_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(50), "oneof=42 100")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_Uint8Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint8(255), "oneof=255 128")
	assert.NoError(t, err)
}

func TestOneOf_Uint16Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint16(65535), "oneof=65535 32768")
	assert.NoError(t, err)
}

func TestOneOf_Uint32Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint32(4294967295), "oneof=4294967295 2147483648")
	assert.NoError(t, err)
}

func TestOneOf_Uint64Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint64(18446744073709551615), "oneof=18446744073709551615 0")
	assert.NoError(t, err)
}

func TestOneOf_DiveWithPointersAllMatching_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var([]*string{ptr.Of("apple"), ptr.Of("banana")}, "dive,oneof=apple banana cherry")
	assert.NoError(t, err)
}

func TestOneOf_DiveWithNilPointerElement_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var([]*string{ptr.Of("apple"), nil, ptr.Of("cherry")}, "dive,oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is nil")
}

func TestOneOf_DiveWithPointerElementNotMatching_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var([]*string{ptr.Of("apple"), ptr.Of("grape")}, "dive,oneof=apple banana cherry")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_AllowedValuesWithSpecialCharacters_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("@pple!", "oneof=@pple! #banana$ %cherry%")
	assert.NoError(t, err)
}

func TestOneOf_NumericStringMatchesAllowedValue_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("42", "oneof=42 100 200")
	assert.NoError(t, err)
}

func TestOneOf_MultipleValidatorsWithOneOfPassing_ContinuesToNext(t *testing.T) {
	t.Parallel()
	err := validation.Var("apple", "oneof=apple banana,len=5")
	assert.NoError(t, err)
}

func TestOneOf_MultipleValidatorsWithOneOfFailing_StopsAtOneOf(t *testing.T) {
	t.Parallel()
	err := validation.Var("orange", "oneof=apple banana,len=6")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_FloatMatchesAllowedValue_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(3.14, "oneof=3.14 2.71")
	assert.NoError(t, err)
}

func TestOneOf_FloatNotMatchAllowedValues_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(1.62, "oneof=3.14 2.71")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_UnicodeStringMatches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("日本語", "oneof=日本語 中文 한국어")
	assert.NoError(t, err)
}

func TestOneOf_UnicodeStringNotMatches_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("español", "oneof=日本語 中文 한국어")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_EmojiMatches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("🎉", "oneof=🎉 🚀 ✨")
	assert.NoError(t, err)
}

func TestOneOf_EmojiNotMatches_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("🔥", "oneof=🎉 🚀 ✨")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_Float32Matches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(1.5), "oneof=1.5 2.5")
	assert.NoError(t, err)
}

func TestOneOf_Float32NotMatches_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(3.5), "oneof=1.5 2.5")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_ZeroInAllowedList_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "oneof=0 1 2")
	assert.NoError(t, err)
}

func TestOneOf_ZeroNotInAllowedList_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "oneof=1 2 3")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_DiveWithIntegersAllMatching_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 2, 3}, "dive,oneof=1 2 3")
	assert.NoError(t, err)
}

func TestOneOf_DiveWithIntegerNotMatching_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 2, 4}, "dive,oneof=1 2 3")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_DiveWithEmptySlice_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var([]string{}, "dive,oneof=a b c")
	assert.NoError(t, err)
}

func TestOneOf_PointerToIntegerMatching_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(42), "oneof=42 100")
	assert.NoError(t, err)
}

func TestOneOf_PointerToIntegerNotMatching_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(50), "oneof=42 100")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}

func TestOneOf_ManyOptionsMatchesLastValue_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("z", "oneof=a b c d e f g h i j k l m n o p q r s t u v w x y z")
	assert.NoError(t, err)
}

func TestOneOf_ManyOptionsMatchesFirstValue_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var("a", "oneof=a b c d e f g h i j k l m n o p q r s t u v w x y z")
	assert.NoError(t, err)
}

func TestOneOf_ParametersOnlyWhitespace_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var("test", "oneof=   ")
	assert.ErrorPart(t, err, "no parameters provided")
}

func TestOneOf_ByteMatches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(byte(65), "oneof=65 66 67")
	assert.NoError(t, err)
}

func TestOneOf_RuneMatches_Passes(t *testing.T) {
	t.Parallel()
	err := validation.Var(rune(65), "oneof=65 66 67")
	assert.NoError(t, err)
}

func TestOneOf_EmptySliceWithoutDive_Fails(t *testing.T) {
	t.Parallel()
	err := validation.Var([]string{}, "oneof=a b c")
	assert.ErrorPart(t, err, "value is not one of the allowed values")
}
