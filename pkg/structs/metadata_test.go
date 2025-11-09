package structs_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestStructMetadata(t *testing.T) {
	t.Parallel()

	t.Run("when the type is not a struct it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			_ = structs.Metadata[int]()
		}, "Type must be a struct or a pointer to a struct")
	})

	t.Run("when the type is a pointer to a struct it return the structs meta", func(t *testing.T) {
		t.Parallel()
		metadata := structs.Metadata[*struct{ Value int }]()
		assert.Equals(t, metadata.Size(), 1)
	})

	t.Run("when the struct is empty it should return an empty map", func(t *testing.T) {
		t.Parallel()
		metadata := structs.Metadata[struct{}]()
		assert.Equals(t, metadata.Size(), 0)
	})

	t.Run("when a struct has a string field called Value and no tag it should return the field name and its type without metadata", func(t *testing.T) {
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
		assert.Equals(t, valueField.Anonymous().Size(), 0)
	})

	t.Run("when a struct has a string field called Value and a tag it should return the field name and its type with the tag metadata", func(t *testing.T) {
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
		assert.Equals(t, valueField.Anonymous().Size(), 0)
	})

	t.Run("when a struct has a string field called Value and tags with multiple fields it should return the field name and its type with the tags metadata", func(t *testing.T) {
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
		assert.Equals(t, valueField.Anonymous().Size(), 0)
	})

	t.Run("when a struct has multiple fields with tags, it should return their field names and type with their tags metadata", func(t *testing.T) {
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
	})

	t.Run("when a struct has nested anonymous structs with fields and tags it should include the anonymous structs fields", func(t *testing.T) {
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
		assert.Equals(t, deepField.Anonymous().Size(), 2)
		assert.Equals(t, deepField.Anonymous().At(0), "embeddedStruct1")
		assert.Equals(t, deepField.Anonymous().At(1), "deepStruct")

		embeddedField1 := metadata.Get("EmbeddedField1")
		assert.True(t, metadata.Has("EmbeddedField1"))
		assert.Equals(t, embeddedField1.Type().Kind(), reflect.String)
		assert.Equals(t, embeddedField1.Tags().Get("key"), "Embedded1")
		assert.Equals(t, embeddedField1.Anonymous().Size(), 1)
		assert.Equals(t, embeddedField1.Anonymous().At(0), "embeddedStruct1")

		embeddedField2 := metadata.Get("EmbeddedField2")
		assert.True(t, metadata.Has("EmbeddedField2"))
		assert.Equals(t, embeddedField2.Type().Kind(), reflect.String)
		assert.Equals(t, embeddedField2.Tags().Get("key"), "Embedded2")
		assert.Equals(t, embeddedField2.Anonymous().Size(), 1)
		assert.Equals(t, embeddedField2.Anonymous().At(0), "embeddedStruct2")

		outerField := metadata.Get("OuterField")
		assert.True(t, metadata.Has("OuterField"))
		assert.Equals(t, outerField.Type().Kind(), reflect.String)
		assert.Equals(t, outerField.Tags().Get("key"), "Outer")
		assert.Equals(t, outerField.Anonymous().Size(), 0)
	})

	t.Run("when a struct and a nested struct both have fields with the same name it should panic", func(t *testing.T) {
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
	})

	t.Run("when StructMedata is called concurrently is should have no errors", func(t *testing.T) {
		t.Parallel()

		type testStruct struct {
			Value float32 `key1:"Value1" key2:"Value2"`
		}

		const threadCount = 8
		const loopCount = 1000
		wg := sync.WaitGroup{}
		waitChan := make(chan struct{})

		for range threadCount {
			wg.Go(func() {
				<-waitChan
				for range loopCount {
					metadata := structs.Metadata[testStruct]()
					assert.Equals(t, metadata.Size(), 1)
					valueField := metadata.Get("Value")
					assert.True(t, metadata.Has("Value"))
					assert.Equals(t, valueField.Type().Kind(), reflect.Float32)
					assert.Equals(t, valueField.Tags().Get("key1"), "Value1")
					assert.Equals(t, valueField.Tags().Get("key2"), "Value2")
					assert.Equals(t, valueField.Anonymous().Size(), 0)
				}
			})
		}

		close(waitChan)
		wg.Wait()
	})
}
