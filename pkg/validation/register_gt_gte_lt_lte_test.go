package validation_test

import (
	"fmt"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestComparisonValidations(t *testing.T) {
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
