package validation_test

import (
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/ptr"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
	"github.com/TriangleSide/go-toolkit/pkg/validation"
)

func TestDiveValidator_NilValue_ReturnsNilError(t *testing.T) {
	t.Parallel()
	err := validation.Var(([]any)(nil), "dive")
	assert.ErrorPart(t, err, "value is nil")
}

func TestDiveValidator_SliceOfIntAllGreaterThanZero_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 2, 3}, "dive,gt=0")
	assert.NoError(t, err)
}

func TestDiveValidator_SliceOfIntWithOneLessThanOrEqualToZero_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 0, 3}, "dive,gt=0")
	assert.ErrorPart(t, err, "value 0 must be greater than 0")
}

func TestDiveValidator_SliceOfUintAllGreaterThanZero_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]uint{1, 2, 3}, "dive,gt=0")
	assert.NoError(t, err)
}

func TestDiveValidator_SliceOfFloat32AllGreaterThanZero_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]float32{1.1, 2.2, 3.3}, "dive,gt=0")
	assert.NoError(t, err)
}

func TestDiveValidator_SliceOfFloat32WithValueLessThanOrEqualToZero_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]float32{1.1, -1.0, 3.3}, "dive,gt=0")
	assert.ErrorPart(t, err, "value -1 must be greater than 0")
}

func TestDiveValidator_SliceOfPointerToIntAllGreaterThanZero_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]*int{ptr.Of(1), ptr.Of(2)}, "dive,gt=0")
	assert.NoError(t, err)
}

func TestDiveValidator_SliceOfPointerToIntWithOneLessThanOrEqualToZero_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]*int{ptr.Of(1), ptr.Of(0)}, "dive,gt=0")
	assert.ErrorPart(t, err, "value 0 must be greater than 0")
}

func TestDiveValidator_SliceOfIntAllRequired_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 2, 3}, "dive,required")
	assert.NoError(t, err)
}

func TestDiveValidator_SliceOfPointerToIntWithNilValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]*int{ptr.Of(1), nil}, "dive,required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestDiveValidator_SliceOfStringsWithEmptyString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]string{"a", "", "c"}, "dive,required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestDiveValidator_NestedDiveWithEmptyString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([][]string{{"a"}, {""}}, "dive,dive,required")
	assert.ErrorPart(t, err, "the value is the zero-value")
}

func TestDiveValidator_NestedDiveWithNilSlice_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([][]string{{"a"}, nil}, "dive,dive,required")
	assert.ErrorPart(t, err, "value is nil")
}

func TestDiveValidator_DiveAsOnlyArgument_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]string{"a", "b"}, "dive")
	assert.ErrorPart(t, err, "empty validate instructions")
}

func TestDiveValidator_SliceOfNonZeroIntValues_ReturnsNoError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 2, 3}, "dive,gt=0")
	assert.NoError(t, err)
}

func TestDiveValidator_NilSliceOfIntegers_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(([]int)(nil), "dive,gt=0")
	assert.ErrorPart(t, err, "value is nil")
}

func TestDiveValidator_NonSliceValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(10, "dive,gt=0")
	assert.ErrorPart(t, err, "dive validator only accepts slice values")
}

func TestDiveValidator_SliceWithInvalidType_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var([]any{1, "test"}, "dive,gt=0")
	assert.ErrorPart(t, err, "gt validation not supported for kind string")
}
