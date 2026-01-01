package structs_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestValueFromName_NilStructInstance_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string
	}
	var myStructPointer *testStruct
	_, err := structs.ValueFromName(myStructPointer, "Value")
	assert.ErrorExact(t, err, "struct instance cannot be nil")
}

func TestValueFromName_NonStructType_ReturnsError(t *testing.T) {
	t.Parallel()
	nonStruct := 123
	_, err := structs.ValueFromName(nonStruct, "Value")
	assert.ErrorExact(t, err, "type must be a struct or a pointer to a struct")
}

func TestValueFromName_PointerToNonStructType_ReturnsError(t *testing.T) {
	t.Parallel()
	nonStruct := 123
	_, err := structs.ValueFromName(&nonStruct, "Value")
	assert.ErrorExact(t, err, "type must be a struct or a pointer to a struct")
}

func TestValueFromName_FieldDoesNotExist_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string
	}
	myStruct := testStruct{Value: "value"}
	_, err := structs.ValueFromName(myStruct, "NonExistentField")
	assert.ErrorPart(t, err, "field NonExistentField does not exist in the struct")
}

func TestValueFromName_FieldExistsInStruct_ReturnsFieldValue(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string
	}
	myStruct := testStruct{Value: "value"}
	value, err := structs.ValueFromName(myStruct, "Value")
	assert.NoError(t, err)
	assert.Equals(t, "value", value.Interface())
}

func TestValueFromName_FieldExistsInPointerToStruct_ReturnsFieldValue(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string
	}
	myStruct := &testStruct{Value: "value"}
	value, err := structs.ValueFromName(myStruct, "Value")
	assert.NoError(t, err)
	assert.Equals(t, "value", value.Interface())
}

func TestValueFromName_EmbeddedFields_ReturnsFieldValues(t *testing.T) {
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
	assert.Equals(t, len(allMetadata), 4)
	for fieldName := range allMetadata {
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
}
