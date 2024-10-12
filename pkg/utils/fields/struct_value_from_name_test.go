package fields_test

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/utils/fields"
)

func TestStructValueFromName(t *testing.T) {
	t.Parallel()

	type deepEmbedded struct {
		DeepEmbeddedField string
	}

	type embedded struct {
		deepEmbedded
		EmbeddedField string
	}

	type structValue struct {
		StructField string
	}

	type testStruct struct {
		embedded
		StructValue structValue
		Value       string
	}

	t.Run("when struct instance is nil it should return an error indicating struct cannot be nil", func(t *testing.T) {
		t.Parallel()
		var myStructPointer *testStruct = nil
		_, err := fields.StructValueFromName(myStructPointer, "Value")
		assert.ErrorExact(t, err, "struct instance cannot be nil")
	})

	t.Run("when type is not a struct it should return an error indicating type must be struct or pointer to struct", func(t *testing.T) {
		t.Parallel()
		nonStruct := 123
		_, err := fields.StructValueFromName(nonStruct, "Value")
		assert.ErrorExact(t, err, "type must be a struct or a pointer to a struct")
	})

	t.Run("when field does not exist it should return an error indicating the field does not exist", func(t *testing.T) {
		t.Parallel()
		myStruct := testStruct{Value: "value"}
		_, err := fields.StructValueFromName(myStruct, "NonExistentField")
		assert.ErrorPart(t, err, "field NonExistentField does not exist in the struct")
	})

	t.Run("when a field exists in a struct it should return the field value", func(t *testing.T) {
		t.Parallel()
		myStruct := testStruct{Value: "value"}
		value, err := fields.StructValueFromName(myStruct, "Value")
		assert.NoError(t, err)
		assert.Equals(t, "value", value.Interface())
	})

	t.Run("when field exists in pointer to struct it should return the field value", func(t *testing.T) {
		t.Parallel()
		myStruct := &testStruct{Value: "value"}
		value, err := fields.StructValueFromName(myStruct, "Value")
		assert.NoError(t, err)
		assert.Equals(t, "value", value.Interface())
	})

	t.Run("when the struct has an embedded field it should be able to get the value", func(t *testing.T) {
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
			Value: "Value",
		}
		allMetadata := fields.StructMetadata[testStruct]()
		assert.Equals(t, allMetadata.Size(), 4)
		for fieldName, _ := range allMetadata.Iterator() {
			_, err := fields.StructValueFromName(instance, fieldName)
			assert.NoError(t, err)
		}
		value, err := fields.StructValueFromName(instance, "Value")
		assert.NoError(t, err)
		assert.Equals(t, value.Interface(), "Value")
		value, err = fields.StructValueFromName(instance, "StructValue")
		assert.NoError(t, err)
		assert.Equals(t, value.Interface(), structValue{StructField: "StructField"})
		value, err = fields.StructValueFromName(instance, "EmbeddedField")
		assert.NoError(t, err)
		assert.Equals(t, value.Interface(), "EmbeddedField")
		value, err = fields.StructValueFromName(instance, "DeepEmbeddedField")
		assert.NoError(t, err)
		assert.Equals(t, value.Interface(), "DeepEmbeddedField")
	})
}
