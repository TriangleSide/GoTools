package structs_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestMetadata_NonStructType_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		_ = structs.Metadata[int]()
	}, "type must be a struct or a pointer to a struct")
}

func TestMetadata_PointerToStruct_ReturnsStructMeta(t *testing.T) {
	t.Parallel()
	metadata := structs.Metadata[*struct{ Value int }]()
	assert.Equals(t, metadata.Size(), 1)
}

func TestMetadata_EmptyStruct_ReturnsEmptyMap(t *testing.T) {
	t.Parallel()
	metadata := structs.Metadata[struct{}]()
	assert.Equals(t, metadata.Size(), 0)
}

func TestMetadata_StringFieldWithoutTag_ReturnsFieldNameAndTypeWithoutMetadata(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string
	}
	metadata := structs.Metadata[testStruct]()
	assert.Equals(t, metadata.Size(), 1)
	valueField := metadata.Get("Value")
	assert.True(t, metadata.Has("Value"))
	assert.Equals(t, valueField.Type().Kind(), reflect.String)
	assert.Equals(t, valueField.Tags().Size(), 0)
	assert.Equals(t, len(valueField.Anonymous()), 0)
}

func TestMetadata_FieldWithTag_ReturnsFieldNameAndTypeWithTagMetadata(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value int `key:"Value"`
	}
	metadata := structs.Metadata[testStruct]()
	assert.Equals(t, metadata.Size(), 1)
	valueField := metadata.Get("Value")
	assert.True(t, metadata.Has("Value"))
	assert.Equals(t, valueField.Type().Kind(), reflect.Int)
	assert.Equals(t, valueField.Tags().Get("key"), "Value")
	assert.Equals(t, len(valueField.Anonymous()), 0)
}

func TestMetadata_FieldWithMultipleTags_ReturnsFieldNameAndTypeWithTagsMetadata(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value float32 `key1:"Value1" key2:"Value2"`
	}
	metadata := structs.Metadata[testStruct]()
	assert.Equals(t, metadata.Size(), 1)
	valueField := metadata.Get("Value")
	assert.True(t, metadata.Has("Value"))
	assert.Equals(t, valueField.Type().Kind(), reflect.Float32)
	assert.Equals(t, valueField.Tags().Get("key1"), "Value1")
	assert.Equals(t, valueField.Tags().Get("key2"), "Value2")
	assert.Equals(t, len(valueField.Anonymous()), 0)
}

func TestMetadata_MultipleFieldsWithTags_ReturnsFieldNamesAndTypesWithTagsMetadata(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value1 string `key2:"Value3" key4:"Value5"`
		Value6 string `key7:"Value8" key9:"Value10"`
	}

	metadata := structs.Metadata[testStruct]()
	assert.Equals(t, metadata.Size(), 2)
	value1Field := metadata.Get("Value1")
	assert.True(t, metadata.Has("Value1"))
	assert.Equals(t, value1Field.Type().Kind(), reflect.String)
	assert.Equals(t, value1Field.Tags().Get("key2"), "Value3")
	assert.Equals(t, value1Field.Tags().Get("key4"), "Value5")

	value6Field := metadata.Get("Value6")
	assert.True(t, metadata.Has("Value6"))
	assert.Equals(t, value6Field.Type().Kind(), reflect.String)
	assert.Equals(t, value6Field.Tags().Get("key7"), "Value8")
	assert.Equals(t, value6Field.Tags().Get("key9"), "Value10")
}

func TestMetadata_NestedAnonymousStructs_IncludesAnonymousStructsFields(t *testing.T) {
	t.Parallel()

	type deepStruct struct {
		DeepField string `key:"Deep"`
	}

	type embeddedStruct1 struct {
		deepStruct

		EmbeddedField1 string `key:"Embedded1"`
	}

	type embeddedStruct2 struct {
		EmbeddedField2 string `key:"Embedded2"`
	}

	type outerStruct struct {
		embeddedStruct1
		embeddedStruct2

		OuterField string `key:"Outer"`
	}

	metadata := structs.Metadata[outerStruct]()
	assert.Equals(t, metadata.Size(), 4)
	deepField := metadata.Get("DeepField")
	assert.True(t, metadata.Has("DeepField"))
	assert.Equals(t, deepField.Type().Kind(), reflect.String)
	assert.Equals(t, deepField.Tags().Get("key"), "Deep")
	assert.Equals(t, len(deepField.Anonymous()), 2)
	assert.Equals(t, deepField.Anonymous()[0], "embeddedStruct1")
	assert.Equals(t, deepField.Anonymous()[1], "deepStruct")

	embeddedField1 := metadata.Get("EmbeddedField1")
	assert.True(t, metadata.Has("EmbeddedField1"))
	assert.Equals(t, embeddedField1.Type().Kind(), reflect.String)
	assert.Equals(t, embeddedField1.Tags().Get("key"), "Embedded1")
	assert.Equals(t, len(embeddedField1.Anonymous()), 1)
	assert.Equals(t, embeddedField1.Anonymous()[0], "embeddedStruct1")

	embeddedField2 := metadata.Get("EmbeddedField2")
	assert.True(t, metadata.Has("EmbeddedField2"))
	assert.Equals(t, embeddedField2.Type().Kind(), reflect.String)
	assert.Equals(t, embeddedField2.Tags().Get("key"), "Embedded2")
	assert.Equals(t, len(embeddedField2.Anonymous()), 1)
	assert.Equals(t, embeddedField2.Anonymous()[0], "embeddedStruct2")

	outerField := metadata.Get("OuterField")
	assert.True(t, metadata.Has("OuterField"))
	assert.Equals(t, outerField.Type().Kind(), reflect.String)
	assert.Equals(t, outerField.Tags().Get("key"), "Outer")
	assert.Equals(t, len(outerField.Anonymous()), 0)
}

func TestMetadata_AmbiguousFieldName_Panics(t *testing.T) {
	t.Parallel()

	type embeddedStruct struct {
		Field string
	}

	type outerStruct struct {
		embeddedStruct

		Field string
	}

	assert.PanicPart(t, func() {
		_ = structs.Metadata[outerStruct]()
	}, "field Field is ambiguous")
}

func TestMetadata_FieldWithEmptyTagValue_ReturnsFieldWithEmptyTagValue(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string `key:""`
	}
	metadata := structs.Metadata[testStruct]()
	assert.Equals(t, metadata.Size(), 1)
	valueField := metadata.Get("Value")
	assert.True(t, metadata.Has("Value"))
	assert.Equals(t, valueField.Tags().Size(), 1)
	assert.Equals(t, valueField.Tags().Get("key"), "")
}

func TestMetadata_EmbeddedPointerToStruct_IncludesPointerStructFields(t *testing.T) {
	t.Parallel()

	type embeddedStruct struct {
		EmbeddedField string `key:"Embedded"`
	}

	type outerStruct struct {
		*embeddedStruct

		OuterField string `key:"Outer"`
	}

	metadata := structs.Metadata[outerStruct]()
	assert.Equals(t, metadata.Size(), 2)

	embeddedField := metadata.Get("EmbeddedField")
	assert.True(t, metadata.Has("EmbeddedField"))
	assert.Equals(t, embeddedField.Type().Kind(), reflect.String)
	assert.Equals(t, embeddedField.Tags().Get("key"), "Embedded")
	assert.Equals(t, len(embeddedField.Anonymous()), 1)
	assert.Equals(t, embeddedField.Anonymous()[0], "embeddedStruct")

	outerField := metadata.Get("OuterField")
	assert.True(t, metadata.Has("OuterField"))
	assert.Equals(t, outerField.Type().Kind(), reflect.String)
	assert.Equals(t, outerField.Tags().Get("key"), "Outer")
	assert.Equals(t, len(outerField.Anonymous()), 0)
}

func TestMetadataFromType_ValidStruct_ReturnsMetadata(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string `key:"value"`
	}
	metadata := structs.MetadataFromType(reflect.TypeFor[testStruct]())
	assert.Equals(t, metadata.Size(), 1)
	valueField := metadata.Get("Value")
	assert.True(t, metadata.Has("Value"))
	assert.Equals(t, valueField.Type().Kind(), reflect.String)
	assert.Equals(t, valueField.Tags().Get("key"), "value")
}

func TestMetadataFromType_NonStructType_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		_ = structs.MetadataFromType(reflect.TypeFor[int]())
	}, "type must be a struct or a pointer to a struct")
}

func TestMetadataFromType_PointerToStruct_ReturnsMetadata(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value int
	}
	metadata := structs.MetadataFromType(reflect.TypeFor[*testStruct]())
	assert.Equals(t, metadata.Size(), 1)
	assert.True(t, metadata.Has("Value"))
}

func TestMetadata_ConcurrentAccess_NoErrors(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Value float32 `key1:"Value1" key2:"Value2"`
	}

	const threadCount = 8
	const loopCount = 1000
	var waitGroup sync.WaitGroup
	waitChan := make(chan struct{})

	for range threadCount {
		waitGroup.Go(func() {
			<-waitChan
			for range loopCount {
				metadata := structs.Metadata[testStruct]()
				assert.Equals(t, metadata.Size(), 1)
				valueField := metadata.Get("Value")
				assert.True(t, metadata.Has("Value"))
				assert.Equals(t, valueField.Type().Kind(), reflect.Float32)
				assert.Equals(t, valueField.Tags().Get("key1"), "Value1")
				assert.Equals(t, valueField.Tags().Get("key2"), "Value2")
				assert.Equals(t, len(valueField.Anonymous()), 0)
			}
		})
	}

	close(waitChan)
	waitGroup.Wait()
}
