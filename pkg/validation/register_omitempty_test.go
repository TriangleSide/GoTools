package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestOmitemptyValidator_StrEmpty_SkipsRequired(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_StrSet_RunsLenEquals4(t *testing.T) {
	t.Parallel()
	err := validation.Var("test", "omitempty,len=4")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_StrSet_LenEquals5Error(t *testing.T) {
	t.Parallel()
	err := validation.Var("test", "omitempty,len=5")
	assert.ErrorPart(t, err, "length 4 must be exactly 5")
}

func TestOmitemptyValidator_IntZero_SkipsGt(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_IntFive_GtOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(5, "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_IntNegative_GtError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-1, "omitempty,gt=0")
	assert.ErrorPart(t, err, "value -1 must be greater than 0")
}

func TestOmitemptyValidator_DiveEmptySlice_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{}, "dive,omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DiveSliceWithZeros_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{0, 0}, "dive,omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DiveSliceWithPositives_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, 2, 3}, "dive,omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DiveSliceWithNegative_Error(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{1, -1, 3}, "dive,omitempty,gt=0")
	assert.ErrorPart(t, err, "value -1 must be greater than 0")
}

func TestOmitemptyValidator_DiveSliceZeroAndNegative_Error(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{0, -1, 2}, "dive,omitempty,gt=0")
	assert.ErrorPart(t, err, "value -1 must be greater than 0")
}

func TestOmitemptyValidator_NilPointer_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var((*int)(nil), "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_IntOne_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(1, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_IntZeroValue_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_EmptyIntSlice_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var([]int{}, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DiveNilPointers_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var([]*int{nil, nil}, "dive,omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DiveMixedPointers_Error(t *testing.T) {
	t.Parallel()
	err := validation.Var([]*int{ptr.Of(-1), nil, ptr.Of(1)}, "dive,omitempty,gt=0")
	assert.ErrorPart(t, err, "value -1 must be greater than 0")
}

func TestOmitemptyValidator_StrHello_LenEquals5Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "omitempty,len=5")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_StrHello_LenEquals4Error(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "omitempty,len=4")
	assert.ErrorPart(t, err, "length 5 must be exactly 4")
}

func TestOmitemptyValidator_MapEmpty_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(map[string]int{}, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_MapSet_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(map[string]int{"a": 1}, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_MapNil_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var((map[string]int)(nil), "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_BoolFalse_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(false, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_BoolTrue_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(true, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_FloatZero_SkipsGt(t *testing.T) {
	t.Parallel()
	err := validation.Var(0.0, "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_FloatPositive_GtOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(1.5, "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_FloatNegative_GtError(t *testing.T) {
	t.Parallel()
	err := validation.Var(-1.5, "omitempty,gt=0")
	assert.ErrorPart(t, err, "value -1.5 must be greater than 0")
}

func TestOmitemptyValidator_ChanNil_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var((chan int)(nil), "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_ChanSet_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(make(chan int), "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_FuncNil_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var((func())(nil), "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_FuncSet_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(func() {}, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_StructZero_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(struct{ A int }{}, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_StructSet_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(struct{ A int }{A: 1}, "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DoublePointerNil_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var((**int)(nil), "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DoublePointerToZero_SkipsGt(t *testing.T) {
	t.Parallel()
	i := 0
	p := &i
	err := validation.Var(&p, "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DoublePointerToFive_GtOk(t *testing.T) {
	t.Parallel()
	i := 5
	p := &i
	err := validation.Var(&p, "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_ComplexZero_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(complex(0, 0), "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_ComplexSet_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(complex(1, 1), "omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_OnlyEmptyStr_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "omitempty")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_OnlyStrValue_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var("hello", "omitempty")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_OnlyIntZero_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var(0, "omitempty")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_OnlyIntValue_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var(42, "omitempty")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_StrPointerEmpty_SkipsLen(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(""), "omitempty,len=5")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_StrPointerHello_LenEquals5Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("hello"), "omitempty,len=5")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_StrPointerHello_LenEquals4Error(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("hello"), "omitempty,len=4")
	assert.ErrorPart(t, err, "length 5 must be exactly 4")
}

func TestOmitemptyValidator_IntPointerZero_SkipsGt(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(0), "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_IntPointerFive_GtOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(5), "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_IntPointerNegative_GtError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(-5), "omitempty,gt=0")
	assert.ErrorPart(t, err, "value -5 must be greater than 0")
}

func TestOmitemptyValidator_DiveStructSliceZero_Ok(t *testing.T) {
	t.Parallel()
	err := validation.Var([]struct{ A int }{{A: 0}, {A: 0}}, "dive,omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_DiveStructSliceSet_RequiredOk(t *testing.T) {
	t.Parallel()
	err := validation.Var([]struct{ A int }{{A: 1}, {A: 2}}, "dive,omitempty,required")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_UintZero_SkipsGt(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(0), "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_UintFive_GtOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(uint(5), "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_Int8Zero_SkipsGt(t *testing.T) {
	t.Parallel()
	err := validation.Var(int8(0), "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_Int8Five_GtOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(int8(5), "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_Float32Zero_SkipsGt(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(0), "omitempty,gt=0")
	assert.NoError(t, err)
}

func TestOmitemptyValidator_Float32Positive_GtOk(t *testing.T) {
	t.Parallel()
	err := validation.Var(float32(1.5), "omitempty,gt=0")
	assert.NoError(t, err)
}
