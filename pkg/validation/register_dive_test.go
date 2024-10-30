package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/ptr"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

func TestDiveValidatorWithValidations(t *testing.T) {
	t.Parallel()

	type testCaseDefinition struct {
		Name             string
		Value            any
		Validation       string
		ExpectedErrorMsg string
	}

	t.Run("when using 'dive' with validations it should...", func(t *testing.T) {
		t.Parallel()

		testCases := []testCaseDefinition{
			{
				Name:             "dive is used on a nil value",
				Value:            nil,
				Validation:       "dive",
				ExpectedErrorMsg: "the value could not be dereferenced",
			},
			{
				Name:             "slice of int values all greater than 0",
				Value:            []int{1, 2, 3},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "",
			},
			{
				Name:             "slice of int values with one less than or equal to 0",
				Value:            []int{1, 0, 3},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "value 0 must be greater than 0",
			},
			{
				Name:             "slice of uint values all greater than 0",
				Value:            []uint{1, 2, 3},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "",
			},
			{
				Name:             "slice of float32 values all greater than 0",
				Value:            []float32{1.1, 2.2, 3.3},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "",
			},
			{
				Name:             "slice of float32 with a value less than or equal to 0",
				Value:            []float32{1.1, -1.0, 3.3},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "value -1 must be greater than 0",
			},
			{
				Name:             "slice of pointer to int values all greater than 0",
				Value:            []*int{ptr.Of(1), ptr.Of(2)},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "",
			},
			{
				Name:             "slice of pointer to int with one less than or equal to 0",
				Value:            []*int{ptr.Of(1), ptr.Of(0)},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "value 0 must be greater than 0",
			},
			{
				Name:             "slice of int values all required",
				Value:            []int{1, 2, 3},
				Validation:       "dive,required",
				ExpectedErrorMsg: "",
			},
			{
				Name:             "slice of pointer to int with nil value",
				Value:            []*int{ptr.Of(1), nil},
				Validation:       "dive,required",
				ExpectedErrorMsg: "the value could not be dereferenced",
			},
			{
				Name:             "slice of strings with empty string",
				Value:            []string{"a", "", "c"},
				Validation:       "dive,required",
				ExpectedErrorMsg: "the value is the zero-value",
			},
			{
				Name:             "dive is called twice with an empty string",
				Value:            [][]string{{"a"}, {""}},
				Validation:       "dive,dive,required",
				ExpectedErrorMsg: "the value is the zero-value",
			},
			{
				Name:             "dive is called twice with a nil slice value",
				Value:            [][]string{{"a"}, nil},
				Validation:       "dive,dive,required",
				ExpectedErrorMsg: "the value could not be dereferenced",
			},
			{
				Name:             "dive is the only argument",
				Value:            []string{"a", "b"},
				Validation:       "dive",
				ExpectedErrorMsg: "empty validate instructions",
			},
			{
				Name:             "slice of non-zero int values",
				Value:            []int{1, 2, 3},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "",
			},
			{
				Name:             "nil slice of integers",
				Value:            ([]int)(nil),
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "the value could not be dereferenced",
			},
			{
				Name:             "non-slice value with dive",
				Value:            10,
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "dive validator only accepts slice values",
			},
			{
				Name:             "slice with one invalid type",
				Value:            []interface{}{1, "test"},
				Validation:       "dive,gt=0",
				ExpectedErrorMsg: "gt validation not supported for kind string",
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				t.Parallel()
				err := validation.Var(tc.Value, tc.Validation)
				if tc.ExpectedErrorMsg != "" {
					assert.ErrorPart(t, err, tc.ExpectedErrorMsg)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}
