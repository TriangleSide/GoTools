package validation

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestValidation(t *testing.T) {
	t.Parallel()

	t.Run("when Var is called with a validator that does not exist it should return an error", func(t *testing.T) {
		t.Parallel()
		assert.ErrorPart(t, Var("value", "does_not_exists"), "validation with name 'does_not_exists' is not registered")
	})

	t.Run("when the struct validation has nil as a parameter it should return an error", func(t *testing.T) {
		t.Parallel()
		type testStruct struct{}
		var instance *testStruct = nil
		assert.ErrorPart(t, Struct(instance), "found nil while dereferencing")
	})

	t.Run("when the struct parameter is not a struct it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			_ = Struct[int](1)
		}, "validation parameter must be a struct but got int")
	})

	t.Run("when a struct has embedded fields...", func(t *testing.T) {
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

		t.Run("it should validate the deep embedded struct", func(t *testing.T) {
			t.Parallel()
			instance := &testStruct{
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
			}
			assert.ErrorPart(t, Struct(instance), "validation failed on field 'DeepEmbeddedField' with validator 'required' because the value is the zero-value")
		})

		t.Run("it should validate the embedded struct", func(t *testing.T) {
			t.Parallel()
			instance := &testStruct{
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
			}
			assert.ErrorPart(t, Struct(instance), "validation failed on field 'EmbeddedField' with validator 'required' because the value is the zero-value")
		})

		t.Run("it should validate the value that is a struct", func(t *testing.T) {
			t.Parallel()
			instance := &testStruct{
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
			}
			assert.ErrorPart(t, Struct(instance), "validation failed on field 'StructField' with validator 'required' because the value is the zero-value")
		})

		t.Run("it should validate the value", func(t *testing.T) {
			t.Parallel()
			instance := &testStruct{
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
			}
			assert.ErrorPart(t, Struct(instance), "validation failed on field 'Value' with validator 'required' because the value is the zero-value")
		})
	})

	t.Run("when struct validation has a field that is a struct it should fail if the validation instruction is not correct", func(t *testing.T) {
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
		assert.ErrorPart(t, Struct(instance), "required_if requires a field name and a value to compare")
	})

	t.Run("when the validator has incorrect parts it should fail", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Value string `validate:"oneof=one=two"`
		}
		assert.ErrorPart(t, Struct(&testStruct{
			Value: "one",
		}), "malformed validator and instruction")
	})

	t.Run("when the validate tag is empty it should fail", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Value string `validate:"        "` // nolint:tagalign
		}
		assert.ErrorPart(t, Struct(&testStruct{
			Value: "one",
		}), "empty validate instructions")
	})

	t.Run("when a struct a struct field and its validation fails it should return an error", func(t *testing.T) {
		t.Parallel()
		type fieldStruct struct {
			FieldStructValue int `validate:"gt=0"`
		}
		type testStruct struct {
			Value fieldStruct `validate:"required"`
		}
		assert.ErrorPart(t, Struct(&testStruct{Value: fieldStruct{FieldStructValue: -1}}), "validation failed on field 'FieldStructValue'")
		assert.ErrorPart(t, Var(&testStruct{Value: fieldStruct{FieldStructValue: -1}}, "required"), "validation failed on field 'FieldStructValue'")
	})

	t.Run("when a struct has a slice of structs and one of their validation fails it should return an error", func(t *testing.T) {
		t.Parallel()
		type testSliceStruct struct {
			SliceStructValue int `validate:"gt=0"`
		}
		type testStruct struct {
			Slice []testSliceStruct `validate:"required"`
		}
		assert.ErrorPart(t, Struct(&testStruct{
			Slice: []testSliceStruct{{SliceStructValue: 1}, {SliceStructValue: 0}},
		}), "validation failed on field 'SliceStructValue' with validator 'gt' and parameters '0' because the value 0 must be greater than 0")
	})

	t.Run("when a struct has a slice of structs and one of their validations is incorrectly formatted it should return an error", func(t *testing.T) {
		t.Parallel()
		type testSliceStruct struct {
			SliceStructValue int `validate:"not_exist"`
		}
		type testStruct struct {
			Slice []testSliceStruct `validate:"required"`
		}
		assert.ErrorPart(t, Struct(&testStruct{
			Slice: []testSliceStruct{{SliceStructValue: 1}},
		}), "validation with name 'not_exist' is not registered")
	})

	t.Run("when a struct has a map of structs and one of their validation fails it should return an error", func(t *testing.T) {
		t.Parallel()
		type testMapStruct struct {
			SliceStructValue int `validate:"gt=0"`
		}
		type testStruct struct {
			Map map[testMapStruct]testMapStruct `validate:"required"`
		}
		mapValue := map[testMapStruct]testMapStruct{{SliceStructValue: 1}: {SliceStructValue: -1}}
		assert.ErrorPart(t, Struct(&testStruct{Map: mapValue}), "validation failed on field 'SliceStructValue' with validator 'gt' and parameters '0' because the value -1 must be greater than 0")
		assert.ErrorPart(t, Var(&testStruct{Map: mapValue}, "required"), "validation failed on field 'SliceStructValue' with validator 'gt' and parameters '0' because the value -1 must be greater than 0")
		mapValue = map[testMapStruct]testMapStruct{{SliceStructValue: -2}: {SliceStructValue: 1}}
		assert.ErrorPart(t, Struct(&testStruct{Map: mapValue}), "validation failed on field 'SliceStructValue' with validator 'gt' and parameters '0' because the value -2 must be greater than 0")
		assert.ErrorPart(t, Var(&testStruct{Map: mapValue}, "required"), "validation failed on field 'SliceStructValue' with validator 'gt' and parameters '0' because the value -2 must be greater than 0")
	})

	t.Run("when a struct has a map of structs and the key validation is incorrectly formatted it should return an error", func(t *testing.T) {
		t.Parallel()
		type testMapStruct struct {
			SliceStructValue int `validate:"not_exist"`
		}
		type testStruct struct {
			Map map[testMapStruct]int `validate:"required"`
		}
		mapValue := map[testMapStruct]int{{SliceStructValue: 1}: 0}
		assert.ErrorPart(t, Struct(&testStruct{Map: mapValue}), "validation with name 'not_exist' is not registered")
		assert.ErrorPart(t, Var(&testStruct{Map: mapValue}, "required"), "validation with name 'not_exist' is not registered")
	})

	t.Run("when a struct has a map of structs and the value validation is incorrectly formatted it should return an error", func(t *testing.T) {
		t.Parallel()
		type testMapStruct struct {
			SliceStructValue int `validate:"not_exist"`
		}
		type testStruct struct {
			Map map[string]testMapStruct `validate:"required"`
		}
		mapValue := map[string]testMapStruct{"test": {SliceStructValue: 1}}
		assert.ErrorPart(t, Struct(&testStruct{Map: mapValue}), "validation with name 'not_exist' is not registered")
		assert.ErrorPart(t, Var(&testStruct{Map: mapValue}, "required"), "validation with name 'not_exist' is not registered")
	})

	t.Run("when the callback result is not correctly filled it should return an error", func(t *testing.T) {
		t.Parallel()
		MustRegisterValidator("test_not_filled", func(parameters *CallbackParameters) *CallbackResult { return NewCallbackResult() })
		type testStruct struct {
			Value string `validate:"test_not_filled"`
		}
		assert.ErrorPart(t, Struct(&testStruct{Value: "test"}), "callback response is not correctly filled")
	})

	t.Run("when a cycle is created in a struct it should return an error", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Value *testStruct `validate:"required"`
		}
		value := &testStruct{}
		value.Value = value
		assert.ErrorPart(t, Struct(value), "cycle found in the validation")
		assert.ErrorPart(t, Var(value, "required"), "cycle found in the validation")
	})
}
