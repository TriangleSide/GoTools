package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestOmitemptyValidator_VariousInputs_ReturnsExpectedErrors(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		value         any
		validation    string
		expectedError string
	}

	testCases := []testCase{
		{
			name:          "str empty: omitempty skips required",
			value:         "",
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "str set: omitempty runs len=4",
			value:         "test",
			validation:    "omitempty,len=4",
			expectedError: "",
		},
		{
			name:          "str set: len=5 error",
			value:         "test",
			validation:    "omitempty,len=5",
			expectedError: "length 4 must be exactly 5",
		},
		{
			name:          "int 0: omitempty skips gt",
			value:         0,
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "int 5: gt ok",
			value:         5,
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "int -1: gt error",
			value:         -1,
			validation:    "omitempty,gt=0",
			expectedError: "value -1 must be greater than 0",
		},
		{
			name:          "dive []int{}: ok",
			value:         []int{},
			validation:    "dive,omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "dive []int{0,0}: ok",
			value:         []int{0, 0},
			validation:    "dive,omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "dive []int{1,2,3}: ok",
			value:         []int{1, 2, 3},
			validation:    "dive,omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "dive []int{1,-1,3}: error",
			value:         []int{1, -1, 3},
			validation:    "dive,omitempty,gt=0",
			expectedError: "value -1 must be greater than 0",
		},
		{
			name:          "dive []int{0,-1,2}: error",
			value:         []int{0, -1, 2},
			validation:    "dive,omitempty,gt=0",
			expectedError: "value -1 must be greater than 0",
		},
		{
			name:          "nil: omitempty required ok",
			value:         nil,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "int 1: required ok",
			value:         1,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "int 0: required ok",
			value:         0,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "[]int{}: required ok",
			value:         []int{},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "dive []*int{nil,nil}: ok",
			value:         []*int{nil, nil},
			validation:    "dive,omitempty,gt=0",
			expectedError: "",
		},
		{
			name: "dive []*int{-1,nil,1}: error",
			value: []*int{
				ptr.Of(-1),
				nil,
				ptr.Of(1),
			},
			validation:    "dive,omitempty,gt=0",
			expectedError: "value -1 must be greater than 0",
		},
		{
			name:          "str hello: len=5 ok",
			value:         "hello",
			validation:    "omitempty,len=5",
			expectedError: "",
		},
		{
			name:          "str hello: len=4 error",
			value:         "hello",
			validation:    "omitempty,len=4",
			expectedError: "length 5 must be exactly 4",
		},
		{
			name:          "map empty: required ok",
			value:         map[string]int{},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "map set: required ok",
			value:         map[string]int{"a": 1},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "map nil: required ok",
			value:         (map[string]int)(nil),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "bool false: required ok",
			value:         false,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "bool true: required ok",
			value:         true,
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "float 0: omitempty skips gt",
			value:         0.0,
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "float 1.5: gt ok",
			value:         1.5,
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "float -1.5: gt error",
			value:         -1.5,
			validation:    "omitempty,gt=0",
			expectedError: "value -1.5 must be greater than 0",
		},
		{
			name:          "chan nil: required ok",
			value:         (chan int)(nil),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "chan set: required ok",
			value:         make(chan int),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "func nil: required ok",
			value:         (func())(nil),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "func set: required ok",
			value:         func() {},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "struct zero: required ok",
			value:         struct{ A int }{},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "struct set: required ok",
			value:         struct{ A int }{A: 1},
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "**int nil: required ok",
			value:         (**int)(nil),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name: "**int -> 0: omitempty skips gt",
			value: func() **int {
				i := 0
				p := &i
				return &p
			}(),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name: "**int -> 5: gt ok",
			value: func() **int {
				i := 5
				p := &i
				return &p
			}(),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "complex 0: required ok",
			value:         complex(0, 0),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "complex set: required ok",
			value:         complex(1, 1),
			validation:    "omitempty,required",
			expectedError: "",
		},
		{
			name:          "omitempty only: empty str ok",
			value:         "",
			validation:    "omitempty",
			expectedError: "",
		},
		{
			name:          "omitempty only: str ok",
			value:         "hello",
			validation:    "omitempty",
			expectedError: "",
		},
		{
			name:          "omitempty only: int 0 ok",
			value:         0,
			validation:    "omitempty",
			expectedError: "",
		},
		{
			name:          "omitempty only: int 42 ok",
			value:         42,
			validation:    "omitempty",
			expectedError: "",
		},
		{
			name:          "*str empty: omitempty skips len",
			value:         ptr.Of(""),
			validation:    "omitempty,len=5",
			expectedError: "",
		},
		{
			name:          "*str hello: len=5 ok",
			value:         ptr.Of("hello"),
			validation:    "omitempty,len=5",
			expectedError: "",
		},
		{
			name:          "*str hello: len=4 error",
			value:         ptr.Of("hello"),
			validation:    "omitempty,len=4",
			expectedError: "length 5 must be exactly 4",
		},
		{
			name:          "*int 0: omitempty skips gt",
			value:         ptr.Of(0),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "*int 5: gt ok",
			value:         ptr.Of(5),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "*int -5: gt error",
			value:         ptr.Of(-5),
			validation:    "omitempty,gt=0",
			expectedError: "value -5 must be greater than 0",
		},
		{
			name: "dive []struct{A=0}: ok",
			value: []struct{ A int }{
				{A: 0},
				{A: 0},
			},
			validation:    "dive,omitempty,required",
			expectedError: "",
		},
		{
			name: "dive []struct{A>0}: required ok",
			value: []struct{ A int }{
				{A: 1},
				{A: 2},
			},
			validation:    "dive,omitempty,required",
			expectedError: "",
		},
		{
			name:          "uint 0: omitempty skips gt",
			value:         uint(0),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "uint 5: gt ok",
			value:         uint(5),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "int8 0: omitempty skips gt",
			value:         int8(0),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "int8 5: gt ok",
			value:         int8(5),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "float32 0: omitempty skips gt",
			value:         float32(0),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
		{
			name:          "float32 1.5: gt ok",
			value:         float32(1.5),
			validation:    "omitempty,gt=0",
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validation.Var(tc.value, tc.validation)
			if tc.expectedError != "" {
				assert.ErrorPart(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
