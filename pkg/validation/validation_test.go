package validation_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestVar_UnknownValidator_ReturnsError(t *testing.T) {
	t.Parallel()

	err := validation.Var("value", "does_not_exists")
	assert.ErrorPart(t, err, "validation with name 'does_not_exists' is not registered")
}

func TestVar_WhitespaceOnlyInstructions_ReturnsError(t *testing.T) {
	t.Parallel()

	err := validation.Var("value", " \t ")
	assert.ErrorPart(t, err, "empty validate instructions")
}

func TestVar_TrailingSeparator_ReturnsError(t *testing.T) {
	t.Parallel()

	err := validation.Var("value", string(validation.RequiredValidatorName)+validation.ValidatorsSep)
	assert.ErrorPart(t, err, "validation with name '' is not registered")
}

func TestVar_MalformedSecondInstruction_ReturnsError(t *testing.T) {
	t.Parallel()

	err := validation.Var("value", string(validation.RequiredValidatorName)+validation.ValidatorsSep+"oneof=one=two")
	assert.ErrorPart(t, err, "malformed validator and instruction")
}

func TestVar_ViolationStopsValidation_SkipsRemainingValidators(t *testing.T) {
	t.Parallel()

	firstName := validation.Validator("validation_test_violation_stops_remaining_first")
	secondName := validation.Validator("validation_test_violation_stops_remaining_second")

	firstCallback := func(parameters *validation.CallbackParameters) *validation.CallbackResult {
		return validation.NewCallbackResult().WithError(validation.NewFieldError(parameters, errors.New("first violation")))
	}
	validation.MustRegisterValidator(firstName, firstCallback)
	validation.MustRegisterValidator(secondName, func(*validation.CallbackParameters) *validation.CallbackResult {
		panic(errors.New("should not be called"))
	})

	err := validation.Var("anything", string(firstName)+validation.ValidatorsSep+string(secondName))
	assert.ErrorPart(t, err, "first violation")
}

func TestVar_RequiredWithNonZeroValue_ReturnsNoError(t *testing.T) {
	t.Parallel()

	err := validation.Var("value", string(validation.RequiredValidatorName))
	assert.NoError(t, err)
}

func TestStruct_NilValue_ReturnsError(t *testing.T) {
	t.Parallel()

	type testStruct struct{}
	var instance *testStruct
	err := validation.Struct(instance)
	assert.ErrorPart(t, err, "value is nil")
}

func TestStruct_NonStructParameter_ReturnsError(t *testing.T) {
	t.Parallel()

	err := validation.Struct[int](1)
	assert.ErrorPart(t, err, "validation parameter must be a struct but got int")
}

func TestStruct_EmbeddedFields_ValidatesAllFields(t *testing.T) {
	t.Parallel()

	type deepEmbedded struct {
		DeepEmbeddedField string `validate:"required"`
	}

	type embedded struct {
		deepEmbedded

		EmbeddedField string `validate:"required"`
	}

	type structValue struct {
		StructField string `validate:"required"`
	}

	type testStruct struct {
		embedded

		StructValue structValue
		Value       string `validate:"required"`
	}

	type testCase struct {
		name              string
		instance          *testStruct
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name: "deep_embedded_field_is_zero_value",
			instance: &testStruct{
				embedded: embedded{
					deepEmbedded: deepEmbedded{
						DeepEmbeddedField: "",
					},
					EmbeddedField: "EmbeddedField",
				},
				StructValue: structValue{
					StructField: "StructField",
				},
				Value: "Value",
			},
			expectedErrorPart: "validation failed on field 'DeepEmbeddedField' " +
				"with validator 'required' because the value is the zero-value",
		},
		{
			name: "embedded_field_is_zero_value",
			instance: &testStruct{
				embedded: embedded{
					deepEmbedded: deepEmbedded{
						DeepEmbeddedField: "DeepEmbeddedField",
					},
					EmbeddedField: "",
				},
				StructValue: structValue{
					StructField: "StructField",
				},
				Value: "Value",
			},
			expectedErrorPart: "validation failed on field 'EmbeddedField' " +
				"with validator 'required' because the value is the zero-value",
		},
		{
			name: "struct_field_is_zero_value",
			instance: &testStruct{
				embedded: embedded{
					deepEmbedded: deepEmbedded{
						DeepEmbeddedField: "DeepEmbeddedField",
					},
					EmbeddedField: "EmbeddedField",
				},
				StructValue: structValue{
					StructField: "",
				},
				Value: "Value",
			},
			expectedErrorPart: "validation failed on field 'StructField' with validator 'required'" +
				" because the value is the zero-value",
		},
		{
			name: "value_is_zero_value",
			instance: &testStruct{
				embedded: embedded{
					deepEmbedded: deepEmbedded{
						DeepEmbeddedField: "DeepEmbeddedField",
					},
					EmbeddedField: "EmbeddedField",
				},
				StructValue: structValue{
					StructField: "StructField",
				},
				Value: "",
			},
			expectedErrorPart: "validation failed on field 'Value' with validator 'required'" +
				" because the value is the zero-value",
		},
		{
			name: "all_fields_set",
			instance: &testStruct{
				embedded: embedded{
					deepEmbedded: deepEmbedded{
						DeepEmbeddedField: "DeepEmbeddedField",
					},
					EmbeddedField: "EmbeddedField",
				},
				StructValue: structValue{
					StructField: "StructField",
				},
				Value: "Value",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.Struct(testCase.instance)
			if testCase.expectedErrorPart == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestStruct_InvalidNestedStructFieldInstruction_ReturnsError(t *testing.T) {
	t.Parallel()

	type StructField struct {
		Field string `validate:"required_if=NotExists"`
	}
	type testStruct struct {
		StructField StructField
		Value       string `validate:"required"`
	}
	instance := &testStruct{
		StructField: StructField{
			Field: "Value",
		},
		Value: "Value",
	}
	err := validation.Struct(instance)
	assert.ErrorPart(t, err, "required_if requires a field name and a value to compare")
}

func TestStruct_MalformedValidatorInstruction_ReturnsError(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Value string `validate:"oneof=one=two"`
	}
	err := validation.Struct(&testStruct{
		Value: "one",
	})
	assert.ErrorPart(t, err, "malformed validator and instruction")
}

func TestStruct_EmptyValidateTag_ReturnsError(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Value string `validate:""`
	}
	err := validation.Struct(&testStruct{
		Value: "one",
	})
	assert.ErrorPart(t, err, "empty validate instructions")
}

func TestStruct_NestedStructFieldViolation_ReturnsError(t *testing.T) {
	t.Parallel()

	type fieldStruct struct {
		FieldStructValue int `validate:"gt=0"`
	}
	type testStruct struct {
		Value fieldStruct `validate:"required"`
	}

	instance := &testStruct{Value: fieldStruct{FieldStructValue: -1}}
	err := validation.Struct(instance)
	assert.ErrorPart(t, err, "validation failed on field 'FieldStructValue'")

	err = validation.Var(instance, string(validation.RequiredValidatorName))
	assert.ErrorPart(t, err, "validation failed on field 'FieldStructValue'")
}

func TestVar_InterfaceHoldingStruct_InvalidInnerField_ReturnsError(t *testing.T) {
	t.Parallel()

	type inner struct {
		InnerValue string `validate:"required"`
	}
	type outer struct {
		Any any
	}

	type testCase struct {
		name              string
		instance          outer
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name:     "valid_inner_struct",
			instance: outer{Any: inner{InnerValue: "ok"}},
		},
		{
			name:              "invalid_inner_struct",
			instance:          outer{Any: inner{InnerValue: ""}},
			expectedErrorPart: "validation failed on field 'InnerValue'",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.Var(testCase.instance, string(validation.RequiredValidatorName))
			if testCase.expectedErrorPart == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestStruct_SliceOfStructs_Violation_ReturnsError(t *testing.T) {
	t.Parallel()

	type testSliceStruct struct {
		SliceStructValue int `validate:"gt=0"`
	}
	type testStruct struct {
		Slice []testSliceStruct `validate:"required"`
	}

	err := validation.Struct(&testStruct{
		Slice: []testSliceStruct{{SliceStructValue: 1}, {SliceStructValue: 0}},
	})
	assert.ErrorPart(t, err, "validation failed on field 'SliceStructValue' with validator 'gt'"+
		" and parameters '0' because the value 0 must be greater than 0")
}

func TestStruct_SliceOfStructs_UnknownValidator_ReturnsError(t *testing.T) {
	t.Parallel()

	type testSliceStruct struct {
		SliceStructValue int `validate:"not_exist"`
	}
	type testStruct struct {
		Slice []testSliceStruct `validate:"required"`
	}

	err := validation.Struct(&testStruct{
		Slice: []testSliceStruct{{SliceStructValue: 1}},
	})
	assert.ErrorPart(t, err, "validation with name 'not_exist' is not registered")
}

func TestStruct_MapOfStructs_Violation_ReturnsError(t *testing.T) {
	t.Parallel()

	type testMapStruct struct {
		SliceStructValue int `validate:"gt=0"`
	}
	type testStruct struct {
		Map map[testMapStruct]testMapStruct `validate:"required"`
	}

	type testCase struct {
		name              string
		mapValue          map[testMapStruct]testMapStruct
		expectedErrorPart string
	}

	testCases := []testCase{
		{
			name:     "invalid_value_struct",
			mapValue: map[testMapStruct]testMapStruct{{SliceStructValue: 1}: {SliceStructValue: -1}},
			expectedErrorPart: "validation failed on field 'SliceStructValue' with validator 'gt' and parameters '0'" +
				" because the value -1 must be greater than 0",
		},
		{
			name:     "invalid_key_struct",
			mapValue: map[testMapStruct]testMapStruct{{SliceStructValue: -2}: {SliceStructValue: 1}},
			expectedErrorPart: "validation failed on field 'SliceStructValue' with validator 'gt' and parameters '0'" +
				" because the value -2 must be greater than 0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			instance := &testStruct{Map: testCase.mapValue}

			err := validation.Struct(instance)
			assert.ErrorPart(t, err, testCase.expectedErrorPart)

			err = validation.Var(instance, string(validation.RequiredValidatorName))
			assert.ErrorPart(t, err, testCase.expectedErrorPart)
		})
	}
}

func TestStruct_MapKeyUnknownValidator_ReturnsError(t *testing.T) {
	t.Parallel()

	type testMapStruct struct {
		SliceStructValue int `validate:"not_exist"`
	}
	type testStruct struct {
		Map map[testMapStruct]int `validate:"required"`
	}

	mapValue := map[testMapStruct]int{{SliceStructValue: 1}: 0}
	instance := &testStruct{Map: mapValue}

	err := validation.Struct(instance)
	assert.ErrorPart(t, err, "validation with name 'not_exist' is not registered")

	err = validation.Var(instance, string(validation.RequiredValidatorName))
	assert.ErrorPart(t, err, "validation with name 'not_exist' is not registered")
}

func TestStruct_MapValueUnknownValidator_ReturnsError(t *testing.T) {
	t.Parallel()

	type testMapStruct struct {
		SliceStructValue int `validate:"not_exist"`
	}
	type testStruct struct {
		Map map[string]testMapStruct `validate:"required"`
	}

	mapValue := map[string]testMapStruct{"test": {SliceStructValue: 1}}
	instance := &testStruct{Map: mapValue}

	err := validation.Struct(instance)
	assert.ErrorPart(t, err, "validation with name 'not_exist' is not registered")

	err = validation.Var(instance, string(validation.RequiredValidatorName))
	assert.ErrorPart(t, err, "validation with name 'not_exist' is not registered")
}

func TestStruct_CallbackResultNotFilled_ReturnsError(t *testing.T) {
	t.Parallel()

	validatorName := validation.Validator("validation_test_not_filled")
	validation.MustRegisterValidator(validatorName, func(*validation.CallbackParameters) *validation.CallbackResult {
		return validation.NewCallbackResult()
	})

	type testStruct struct {
		Value string `validate:"validation_test_not_filled"`
	}

	err := validation.Struct(&testStruct{Value: "test"})
	assert.ErrorPart(t, err, "callback response is not correctly filled")
}

func TestStruct_CycleInStruct_ReturnsError(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Value *testStruct `validate:"required"`
	}
	value := &testStruct{}
	value.Value = value

	err := validation.Struct(value)
	assert.ErrorPart(t, err, "cycle found in the validation")

	err = validation.Var(value, string(validation.RequiredValidatorName))
	assert.ErrorPart(t, err, "cycle found in the validation")
}

func TestStruct_ConcurrentValidations_ReturnExpectedResults(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Value string `validate:"required"`
	}

	const workers = 32
	errs := make(chan error, workers)

	var waitGroup sync.WaitGroup
	for workerIdx := range workers {
		waitGroup.Go(func() {
			if workerIdx%2 == 0 {
				errs <- validation.Struct(&testStruct{Value: ""})
				return
			}
			errs <- validation.Struct(&testStruct{Value: "ok"})
		})
	}
	waitGroup.Wait()
	close(errs)

	var gotError bool
	var gotNoError bool
	var firstErr error
	for err := range errs {
		if err == nil {
			gotNoError = true
			continue
		}
		gotError = true
		if firstErr == nil {
			firstErr = err
		}
	}

	assert.True(t, gotError)
	assert.True(t, gotNoError)
	assert.ErrorPart(t, firstErr, "zero-value")
}
