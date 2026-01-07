package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestGtValidator_IntValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(10, "gt=5")
	assert.NoError(t, err)
}

func TestGtValidator_IntValueEqualToThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5, "gt=5")
	assert.ErrorPart(t, err, "value 5 must be greater than 5")
}

func TestGtValidator_IntValueLessThanThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(4, "gt=5")
	assert.ErrorPart(t, err, "value 4 must be greater than 5")
}

func TestGtValidator_UintValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(10), "gt=5")
	assert.NoError(t, err)
}

func TestGtValidator_UintValueEqualToThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(5), "gt=5")
	assert.ErrorPart(t, err, "value 5 must be greater than 5")
}

func TestGtValidator_Float32ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(5.1), "gt=5")
	assert.NoError(t, err)
}

func TestGtValidator_Float32ValueEqualToThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(5.0), "gt=5")
	assert.ErrorPart(t, err, "value 5 must be greater than 5")
}

func TestGtValidator_Float64ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float64(5.1), "gt=5")
	assert.NoError(t, err)
}

func TestGtValidator_PointerToIntGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(10), "gt=5")
	assert.NoError(t, err)
}

func TestGtValidator_PointerToIntEqualToThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(5), "gt=5")
	assert.ErrorPart(t, err, "value 5 must be greater than 5")
}

func TestGtValidator_NilPointerToInt_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*int)(nil), "gt=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestGtValidator_InvalidThresholdParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(10, "gt=abc")
	assert.ErrorPart(t, err, "invalid parameters 'abc' for gt: strconv.ParseFloat: parsing \"abc\": invalid syntax")
}

func TestGtValidator_UnsupportedKindString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("test", "gt=5")
	assert.ErrorPart(t, err, "gt validation not supported for kind string")
}

func TestGtValidator_NegativeThresholdPositiveValue_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5, "gt=-10")
	assert.NoError(t, err)
}

func TestGtValidator_NegativeThresholdNegativeValueAbove_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-5, "gt=-10")
	assert.NoError(t, err)
}

func TestGtValidator_NegativeThresholdNegativeValueBelow_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-15, "gt=-10")
	assert.ErrorPart(t, err, "value -15 must be greater than -10")
}

func TestGtValidator_NegativeThresholdEqualValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-10, "gt=-10")
	assert.ErrorPart(t, err, "value -10 must be greater than -10")
}

func TestGtValidator_ZeroThresholdPositiveValue_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(1, "gt=0")
	assert.NoError(t, err)
}

func TestGtValidator_ZeroThresholdZeroValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "gt=0")
	assert.ErrorPart(t, err, "value 0 must be greater than 0")
}

func TestGtValidator_ZeroThresholdNegativeValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-1, "gt=0")
	assert.ErrorPart(t, err, "value -1 must be greater than 0")
}

func TestGtValidator_FloatThresholdValueAbove_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5.6, "gt=5.5")
	assert.NoError(t, err)
}

func TestGtValidator_FloatThresholdValueEqual_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5.5, "gt=5.5")
	assert.ErrorPart(t, err, "value 5.5 must be greater than 5.5")
}

func TestGtValidator_FloatThresholdValueBelow_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5.4, "gt=5.5")
	assert.ErrorPart(t, err, "value 5.4 must be greater than 5.5")
}

func TestGtValidator_Int8ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(int8(10), "gt=5")
	assert.NoError(t, err)
}

func TestGtValidator_Int8ValueLessThanThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(int8(4), "gt=5")
	assert.ErrorPart(t, err, "value 4 must be greater than 5")
}

func TestGtValidator_Int16ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(int16(1000), "gt=500")
	assert.NoError(t, err)
}

func TestGtValidator_Int32ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(int32(100000), "gt=50000")
	assert.NoError(t, err)
}

func TestGtValidator_Int64ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(int64(1000000000), "gt=500000000")
	assert.NoError(t, err)
}

func TestGtValidator_Uint8ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint8(200), "gt=100")
	assert.NoError(t, err)
}

func TestGtValidator_Uint16ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint16(60000), "gt=50000")
	assert.NoError(t, err)
}

func TestGtValidator_Uint32ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint32(4000000000), "gt=3000000000")
	assert.NoError(t, err)
}

func TestGtValidator_Uint64ValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint64(10000000000), "gt=5000000000")
	assert.NoError(t, err)
}

func TestGtValidator_PointerToFloat32GreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(float32(10.5)), "gt=5")
	assert.NoError(t, err)
}

func TestGtValidator_PointerToFloat64GreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(float64(10.5)), "gt=5")
	assert.NoError(t, err)
}

func TestGtValidator_NilPointerToFloat32_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*float32)(nil), "gt=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestGtValidator_NilPointerToFloat64_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*float64)(nil), "gt=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestGtValidator_UnsupportedKindBool_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(true, "gt=5")
	assert.ErrorPart(t, err, "gt validation not supported for kind bool")
}

func TestGtValidator_UnsupportedKindSlice_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 2, 3}, "gt=5")
	assert.ErrorPart(t, err, "gt validation not supported for kind slice")
}

func TestGtValidator_UnsupportedKindMap_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(map[string]int{"a": 1}, "gt=5")
	assert.ErrorPart(t, err, "gt validation not supported for kind map")
}

func TestGtValidator_UnsupportedKindStruct_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(struct{ X int }{X: 5}, "gt=5")
	assert.ErrorPart(t, err, "gt validation not supported for kind struct")
}

func TestGtValidator_EmptyParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(10, "gt=")
	assert.ErrorPart(t, err, "invalid parameters '' for gt")
}

func TestGteValidator_IntValueGreaterThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(10, "gte=5")
	assert.NoError(t, err)
}

func TestGteValidator_IntValueEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5, "gte=5")
	assert.NoError(t, err)
}

func TestGteValidator_IntValueLessThanThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(4, "gte=5")
	assert.ErrorPart(t, err, "value 4 must be greater than or equal to 5")
}

func TestGteValidator_Float32ValueEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(5.0), "gte=5")
	assert.NoError(t, err)
}

func TestGteValidator_Float32ValueLessThanThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(4.9), "gte=5")
	assert.ErrorPart(t, err, "must be greater than or equal to 5")
}

func TestGteValidator_PointerToIntEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(5), "gte=5")
	assert.NoError(t, err)
}

func TestGteValidator_NilPointerToInt_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*int)(nil), "gte=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestGteValidator_InvalidThresholdParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(10, "gte=abc")
	assert.ErrorPart(t, err, "invalid parameters 'abc' for gte: strconv.ParseFloat: parsing \"abc\": invalid syntax")
}

func TestGteValidator_UnsupportedKindString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("test", "gte=5")
	assert.ErrorPart(t, err, "gte validation not supported for kind string")
}

func TestGteValidator_NegativeThresholdEqualValue_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-10, "gte=-10")
	assert.NoError(t, err)
}

func TestGteValidator_ZeroThresholdZeroValue_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "gte=0")
	assert.NoError(t, err)
}

func TestGteValidator_FloatThresholdValueEqual_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5.5, "gte=5.5")
	assert.NoError(t, err)
}

func TestGteValidator_Float64ValueEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float64(5.0), "gte=5")
	assert.NoError(t, err)
}

func TestGteValidator_UintValueEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(5), "gte=5")
	assert.NoError(t, err)
}

func TestLtValidator_IntValueLessThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(4, "lt=5")
	assert.NoError(t, err)
}

func TestLtValidator_IntValueEqualToThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5, "lt=5")
	assert.ErrorPart(t, err, "value 5 must be less than 5")
}

func TestLtValidator_IntValueGreaterThanThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(6, "lt=5")
	assert.ErrorPart(t, err, "value 6 must be less than 5")
}

func TestLtValidator_Float32ValueLessThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(4.9), "lt=5")
	assert.NoError(t, err)
}

func TestLtValidator_Float32ValueEqualToThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(5.0), "lt=5")
	assert.ErrorPart(t, err, "value 5 must be less than 5")
}

func TestLtValidator_PointerToIntLessThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(4), "lt=5")
	assert.NoError(t, err)
}

func TestLtValidator_NilPointerToInt_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*int)(nil), "lt=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestLtValidator_InvalidThresholdParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(10, "lt=abc")
	assert.ErrorPart(t, err, "invalid parameters 'abc' for lt: strconv.ParseFloat: parsing \"abc\": invalid syntax")
}

func TestLtValidator_UnsupportedKindString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("test", "lt=5")
	assert.ErrorPart(t, err, "lt validation not supported for kind string")
}

func TestLtValidator_NegativeThresholdValueBelow_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-15, "lt=-10")
	assert.NoError(t, err)
}

func TestLtValidator_NegativeThresholdValueAbove_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-5, "lt=-10")
	assert.ErrorPart(t, err, "value -5 must be less than -10")
}

func TestLtValidator_ZeroThresholdNegativeValue_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-1, "lt=0")
	assert.NoError(t, err)
}

func TestLtValidator_ZeroThresholdZeroValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "lt=0")
	assert.ErrorPart(t, err, "value 0 must be less than 0")
}

func TestLtValidator_FloatThresholdValueBelow_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5.4, "lt=5.5")
	assert.NoError(t, err)
}

func TestLtValidator_FloatThresholdValueEqual_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5.5, "lt=5.5")
	assert.ErrorPart(t, err, "value 5.5 must be less than 5.5")
}

func TestLtValidator_Float64ValueLessThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float64(4.9), "lt=5")
	assert.NoError(t, err)
}

func TestLtValidator_UintValueLessThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(4), "lt=5")
	assert.NoError(t, err)
}

func TestLteValidator_IntValueLessThanThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(4, "lte=5")
	assert.NoError(t, err)
}

func TestLteValidator_IntValueEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5, "lte=5")
	assert.NoError(t, err)
}

func TestLteValidator_IntValueGreaterThanThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(6, "lte=5")
	assert.ErrorPart(t, err, "value 6 must be less than or equal to 5")
}

func TestLteValidator_Float32ValueEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(5.0), "lte=5")
	assert.NoError(t, err)
}

func TestLteValidator_Float32ValueGreaterThanThreshold_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(5.1), "lte=5")
	assert.ErrorPart(t, err, "must be less than or equal to 5")
}

func TestLteValidator_PointerToIntEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(5), "lte=5")
	assert.NoError(t, err)
}

func TestLteValidator_NilPointerToInt_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*int)(nil), "lte=5")
	assert.ErrorPart(t, err, "value is nil")
}

func TestLteValidator_InvalidThresholdParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(10, "lte=abc")
	assert.ErrorPart(t, err, "invalid parameters 'abc' for lte: strconv.ParseFloat: parsing \"abc\": invalid syntax")
}

func TestLteValidator_UnsupportedKindString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("test", "lte=5")
	assert.ErrorPart(t, err, "lte validation not supported for kind string")
}

func TestLteValidator_NegativeThresholdEqualValue_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-10, "lte=-10")
	assert.NoError(t, err)
}

func TestLteValidator_ZeroThresholdZeroValue_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "lte=0")
	assert.NoError(t, err)
}

func TestLteValidator_FloatThresholdValueEqual_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(5.5, "lte=5.5")
	assert.NoError(t, err)
}

func TestLteValidator_Float64ValueEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(float64(5.0), "lte=5")
	assert.NoError(t, err)
}

func TestLteValidator_UintValueEqualToThreshold_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(5), "lte=5")
	assert.NoError(t, err)
}
