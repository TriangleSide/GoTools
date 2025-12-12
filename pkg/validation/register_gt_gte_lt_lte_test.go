package validation_test

import (
	"fmt"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestGtGteLtLteValidators(t *testing.T) {
	t.Parallel()

	type testCaseDefinition struct {
		Name             string
		Value            any
		Validation       string
		ExpectedErrorMsg string
	}

	testCases := []testCaseDefinition{
		{
			Name:             "int value greater than threshold",
			Value:            10,
			Validation:       "gt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int value equal to threshold",
			Value:            5,
			Validation:       "gt=5",
			ExpectedErrorMsg: "value 5 must be greater than 5",
		},
		{
			Name:             "int value less than threshold",
			Value:            4,
			Validation:       "gt=5",
			ExpectedErrorMsg: "value 4 must be greater than 5",
		},
		{
			Name:             "uint value greater than threshold",
			Value:            uint(10),
			Validation:       "gt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "uint value equal to threshold",
			Value:            uint(5),
			Validation:       "gt=5",
			ExpectedErrorMsg: "value 5 must be greater than 5",
		},
		{
			Name:             "float32 value greater than threshold",
			Value:            float32(5.1),
			Validation:       "gt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "float32 value equal to threshold",
			Value:            float32(5.0),
			Validation:       "gt=5",
			ExpectedErrorMsg: "value 5 must be greater than 5",
		},
		{
			Name:             "float64 value greater than threshold",
			Value:            float64(5.1),
			Validation:       "gt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "pointer to int greater than threshold",
			Value:            ptr.Of(10),
			Validation:       "gt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "pointer to int equal to threshold",
			Value:            ptr.Of(5),
			Validation:       "gt=5",
			ExpectedErrorMsg: "value 5 must be greater than 5",
		},
		{
			Name:             "nil pointer to int",
			Value:            (*int)(nil),
			Validation:       "gt=5",
			ExpectedErrorMsg: "value is nil",
		},
		{
			Name:             "invalid threshold parameter",
			Value:            10,
			Validation:       "gt=abc",
			ExpectedErrorMsg: "invalid parameters 'abc' for gt: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			Name:             "unsupported kind string",
			Value:            "test",
			Validation:       "gt=5",
			ExpectedErrorMsg: "gt validation not supported for kind string",
		},
		{
			Name:             "int value greater than threshold",
			Value:            10,
			Validation:       "gte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int value equal to threshold",
			Value:            5,
			Validation:       "gte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int value less than threshold",
			Value:            4,
			Validation:       "gte=5",
			ExpectedErrorMsg: "value 4 must be greater than or equal to 5",
		},
		{
			Name:             "float32 value equal to threshold",
			Value:            float32(5.0),
			Validation:       "gte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "float32 value less than threshold",
			Value:            float32(4.9),
			Validation:       "gte=5",
			ExpectedErrorMsg: "must be greater than or equal to 5",
		},
		{
			Name:             "pointer to int equal to threshold",
			Value:            ptr.Of(5),
			Validation:       "gte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "nil pointer to int",
			Value:            (*int)(nil),
			Validation:       "gte=5",
			ExpectedErrorMsg: "value is nil",
		},
		{
			Name:             "invalid threshold parameter",
			Value:            10,
			Validation:       "gte=abc",
			ExpectedErrorMsg: "invalid parameters 'abc' for gte: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			Name:             "unsupported kind string",
			Value:            "test",
			Validation:       "gte=5",
			ExpectedErrorMsg: "gte validation not supported for kind string",
		},
		{
			Name:             "int value less than threshold",
			Value:            4,
			Validation:       "lt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int value equal to threshold",
			Value:            5,
			Validation:       "lt=5",
			ExpectedErrorMsg: "value 5 must be less than 5",
		},
		{
			Name:             "int value greater than threshold",
			Value:            6,
			Validation:       "lt=5",
			ExpectedErrorMsg: "value 6 must be less than 5",
		},
		{
			Name:             "float32 value less than threshold",
			Value:            float32(4.9),
			Validation:       "lt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "float32 value equal to threshold",
			Value:            float32(5.0),
			Validation:       "lt=5",
			ExpectedErrorMsg: "value 5 must be less than 5",
		},
		{
			Name:             "pointer to int less than threshold",
			Value:            ptr.Of(4),
			Validation:       "lt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "nil pointer to int",
			Value:            (*int)(nil),
			Validation:       "lt=5",
			ExpectedErrorMsg: "value is nil",
		},
		{
			Name:             "invalid threshold parameter",
			Value:            10,
			Validation:       "lt=abc",
			ExpectedErrorMsg: "invalid parameters 'abc' for lt: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			Name:             "unsupported kind string",
			Value:            "test",
			Validation:       "lt=5",
			ExpectedErrorMsg: "lt validation not supported for kind string",
		},
		{
			Name:             "int value less than threshold",
			Value:            4,
			Validation:       "lte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int value equal to threshold",
			Value:            5,
			Validation:       "lte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int value greater than threshold",
			Value:            6,
			Validation:       "lte=5",
			ExpectedErrorMsg: "value 6 must be less than or equal to 5",
		},
		{
			Name:             "float32 value equal to threshold",
			Value:            float32(5.0),
			Validation:       "lte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "float32 value greater than threshold",
			Value:            float32(5.1),
			Validation:       "lte=5",
			ExpectedErrorMsg: "must be less than or equal to 5",
		},
		{
			Name:             "pointer to int equal to threshold",
			Value:            ptr.Of(5),
			Validation:       "lte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "nil pointer to int",
			Value:            (*int)(nil),
			Validation:       "lte=5",
			ExpectedErrorMsg: "value is nil",
		},
		{
			Name:             "invalid threshold parameter",
			Value:            10,
			Validation:       "lte=abc",
			ExpectedErrorMsg: "invalid parameters 'abc' for lte: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			Name:             "unsupported kind string",
			Value:            "test",
			Validation:       "lte=5",
			ExpectedErrorMsg: "lte validation not supported for kind string",
		},
		{
			Name:             "gt with negative threshold and positive value",
			Value:            5,
			Validation:       "gt=-10",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "gt with negative threshold and negative value above",
			Value:            -5,
			Validation:       "gt=-10",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "gt with negative threshold and negative value below",
			Value:            -15,
			Validation:       "gt=-10",
			ExpectedErrorMsg: "value -15 must be greater than -10",
		},
		{
			Name:             "gt with negative threshold and equal value",
			Value:            -10,
			Validation:       "gt=-10",
			ExpectedErrorMsg: "value -10 must be greater than -10",
		},
		{
			Name:             "gte with negative threshold and equal value",
			Value:            -10,
			Validation:       "gte=-10",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "lt with negative threshold and value below",
			Value:            -15,
			Validation:       "lt=-10",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "lt with negative threshold and value above",
			Value:            -5,
			Validation:       "lt=-10",
			ExpectedErrorMsg: "value -5 must be less than -10",
		},
		{
			Name:             "lte with negative threshold and equal value",
			Value:            -10,
			Validation:       "lte=-10",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "gt with zero threshold and positive value",
			Value:            1,
			Validation:       "gt=0",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "gt with zero threshold and zero value",
			Value:            0,
			Validation:       "gt=0",
			ExpectedErrorMsg: "value 0 must be greater than 0",
		},
		{
			Name:             "gt with zero threshold and negative value",
			Value:            -1,
			Validation:       "gt=0",
			ExpectedErrorMsg: "value -1 must be greater than 0",
		},
		{
			Name:             "gte with zero threshold and zero value",
			Value:            0,
			Validation:       "gte=0",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "lt with zero threshold and negative value",
			Value:            -1,
			Validation:       "lt=0",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "lt with zero threshold and zero value",
			Value:            0,
			Validation:       "lt=0",
			ExpectedErrorMsg: "value 0 must be less than 0",
		},
		{
			Name:             "lte with zero threshold and zero value",
			Value:            0,
			Validation:       "lte=0",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "gt with float threshold and value above",
			Value:            5.6,
			Validation:       "gt=5.5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "gt with float threshold and value equal",
			Value:            5.5,
			Validation:       "gt=5.5",
			ExpectedErrorMsg: "value 5.5 must be greater than 5.5",
		},
		{
			Name:             "gt with float threshold and value below",
			Value:            5.4,
			Validation:       "gt=5.5",
			ExpectedErrorMsg: "value 5.4 must be greater than 5.5",
		},
		{
			Name:             "gte with float threshold and value equal",
			Value:            5.5,
			Validation:       "gte=5.5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "lt with float threshold and value below",
			Value:            5.4,
			Validation:       "lt=5.5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "lt with float threshold and value equal",
			Value:            5.5,
			Validation:       "lt=5.5",
			ExpectedErrorMsg: "value 5.5 must be less than 5.5",
		},
		{
			Name:             "lte with float threshold and value equal",
			Value:            5.5,
			Validation:       "lte=5.5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int8 value greater than threshold",
			Value:            int8(10),
			Validation:       "gt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int8 value less than threshold",
			Value:            int8(4),
			Validation:       "gt=5",
			ExpectedErrorMsg: "value 4 must be greater than 5",
		},
		{
			Name:             "int16 value greater than threshold",
			Value:            int16(1000),
			Validation:       "gt=500",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int32 value greater than threshold",
			Value:            int32(100000),
			Validation:       "gt=50000",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "int64 value greater than threshold",
			Value:            int64(1000000000),
			Validation:       "gt=500000000",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "uint8 value greater than threshold",
			Value:            uint8(200),
			Validation:       "gt=100",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "uint16 value greater than threshold",
			Value:            uint16(60000),
			Validation:       "gt=50000",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "uint32 value greater than threshold",
			Value:            uint32(4000000000),
			Validation:       "gt=3000000000",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "uint64 value greater than threshold",
			Value:            uint64(10000000000),
			Validation:       "gt=5000000000",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "float64 value equal to threshold for gte",
			Value:            float64(5.0),
			Validation:       "gte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "float64 value less than threshold for lt",
			Value:            float64(4.9),
			Validation:       "lt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "float64 value equal to threshold for lte",
			Value:            float64(5.0),
			Validation:       "lte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "pointer to float32 greater than threshold",
			Value:            ptr.Of(float32(10.5)),
			Validation:       "gt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "pointer to float64 greater than threshold",
			Value:            ptr.Of(float64(10.5)),
			Validation:       "gt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "nil pointer to float32",
			Value:            (*float32)(nil),
			Validation:       "gt=5",
			ExpectedErrorMsg: "value is nil",
		},
		{
			Name:             "nil pointer to float64",
			Value:            (*float64)(nil),
			Validation:       "gt=5",
			ExpectedErrorMsg: "value is nil",
		},
		{
			Name:             "unsupported kind bool",
			Value:            true,
			Validation:       "gt=5",
			ExpectedErrorMsg: "gt validation not supported for kind bool",
		},
		{
			Name:             "unsupported kind slice",
			Value:            []int{1, 2, 3},
			Validation:       "gt=5",
			ExpectedErrorMsg: "gt validation not supported for kind slice",
		},
		{
			Name:             "unsupported kind map",
			Value:            map[string]int{"a": 1},
			Validation:       "gt=5",
			ExpectedErrorMsg: "gt validation not supported for kind map",
		},
		{
			Name:             "unsupported kind struct",
			Value:            struct{ X int }{X: 5},
			Validation:       "gt=5",
			ExpectedErrorMsg: "gt validation not supported for kind struct",
		},
		{
			Name:             "gt with empty parameter",
			Value:            10,
			Validation:       "gt=",
			ExpectedErrorMsg: "invalid parameters '' for gt",
		},
		{
			Name:             "uint value equal to threshold for gte",
			Value:            uint(5),
			Validation:       "gte=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "uint value less than threshold for lt",
			Value:            uint(4),
			Validation:       "lt=5",
			ExpectedErrorMsg: "",
		},
		{
			Name:             "uint value equal to threshold for lte",
			Value:            uint(5),
			Validation:       "lte=5",
			ExpectedErrorMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s (%s)", tc.Name, tc.Validation), func(t *testing.T) {
			t.Parallel()
			err := validation.Var(tc.Value, tc.Validation)
			if tc.ExpectedErrorMsg != "" {
				assert.ErrorPart(t, err, tc.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
