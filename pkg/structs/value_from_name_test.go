package structs_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
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
		_, err := structs.ValueFromName(myStructPointer, "Value")
		assert.ErrorExact(t, err, "struct instance cannot be nil")
	})

	t.Run("when type is not a struct it should return an error indicating type must be struct or pointer to struct", func(t *testing.T) {
		t.Parallel()
		nonStruct := 123
		_, err := structs.ValueFromName(nonStruct, "Value")
		assert.ErrorExact(t, err, "type must be a struct or a pointer to a struct")
	})

	t.Run("when field does not exist it should return an error indicating the field does not exist", func(t *testing.T) {
		t.Parallel()
		myStruct := testStruct{Value: "value"}
		_, err := structs.ValueFromName(myStruct, "NonExistentField")
		assert.ErrorPart(t, err, "field NonExistentField does not exist in the struct")
	})

	t.Run("when a field exists in a struct it should return the field value", func(t *testing.T) {
		t.Parallel()
		myStruct := testStruct{Value: "value"}
		value, err := structs.ValueFromName(myStruct, "Value")
		assert.NoError(t, err)
		assert.Equals(t, "value", value.Interface())
	})

	t.Run("when field exists in pointer to struct it should return the field value", func(t *testing.T) {
		t.Parallel()
		myStruct := &testStruct{Value: "value"}
		value, err := structs.ValueFromName(myStruct, "Value")
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
				StructField: "AssignToField",
			},
			Value: "Value",
		}
		allMetadata := structs.Metadata[testStruct]()
		assert.Equals(t, allMetadata.Size(), 4)
		for fieldName := range allMetadata.All() {
			_, err := structs.ValueFromName(instance, fieldName)
			assert.NoError(t, err)
		}
		value, err := structs.ValueFromName(instance, "Value")
		assert.NoError(t, err)
		assert.Equals(t, value.Interface(), "Value")
		value, err = structs.ValueFromName(instance, "StructValue")
		assert.NoError(t, err)
		assert.Equals(t, value.Interface(), structValue{StructField: "AssignToField"})
		value, err = structs.ValueFromName(instance, "EmbeddedField")
		assert.NoError(t, err)
		assert.Equals(t, value.Interface(), "EmbeddedField")
		value, err = structs.ValueFromName(instance, "DeepEmbeddedField")
		assert.NoError(t, err)
		assert.Equals(t, value.Interface(), "DeepEmbeddedField")
	})
}
