package reflect_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	reflectutils "github.com/TriangleSide/GoBase/pkg/utils/reflect"
)

var _ = Describe("struct field metadata", func() {
	It("should panic if the type is not a struct", func() {
		Expect(func() {
			_ = reflectutils.FieldsToMetadata[int]()
		}).Should(PanicWith(ContainSubstring("type must be a struct")))
	})

	It("should panic if the type is a pointer to a struct", func() {
		Expect(func() {
			_ = reflectutils.FieldsToMetadata[*struct{}]()
		}).Should(PanicWith(ContainSubstring("type must be a struct")))
	})

	It("should return an empty map for an empty struct", func() {
		metadata := reflectutils.FieldsToMetadata[struct{}]()
		Expect(metadata.Size()).To(Equal(0))
	})

	It("should succeed if done many times on the same struct type", func() {
		type testStruct struct{}
		for i := 0; i < 2; i++ {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata.Size()).To(Equal(0))
		}
	})

	When("when a struct has a string field called Value and no tag", func() {
		type testStruct struct {
			Value string
		}

		It("should return the field name and its type with no metadata", func() {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata.Size()).To(Equal(1))
			Expect(metadata.Has("Value")).To(BeTrue())
			Expect(metadata.Get("Value").Type.Kind()).To(Equal(reflect.String))
			Expect(metadata.Get("Value").Tags).To(BeEmpty())
			Expect(metadata.Get("Value").Anonymous).To(BeEmpty())
		})
	})

	When("when a struct has a string field called Value and a tag", func() {
		type testStruct struct {
			Value int `key:"Value"`
		}

		It("should return the field name and its type with the tag metadata", func() {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata.Size()).To(Equal(1))
			Expect(metadata.Has("Value")).To(BeTrue())
			Expect(metadata.Get("Value").Type.Kind()).To(Equal(reflect.Int))
			Expect(metadata.Get("Value").Tags).To(HaveLen(1))
			Expect(metadata.Get("Value").Tags).To(HaveKeyWithValue("key", "Value"))
			Expect(metadata.Get("Value").Anonymous).To(BeEmpty())
		})
	})

	When("when a struct has a string field called Value and a tag with multiple fields", func() {
		type testStruct struct {
			Value float32 `key1:"Value1" key2:"Value2"`
		}

		It("should return the field name and its type with the tags metadata", func() {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata.Size()).To(Equal(1))
			Expect(metadata.Has("Value")).To(BeTrue())
			Expect(metadata.Get("Value").Type.Kind()).To(Equal(reflect.Float32))
			Expect(metadata.Get("Value").Tags).To(HaveLen(2))
			Expect(metadata.Get("Value").Tags).To(HaveKeyWithValue("key1", "Value1"))
			Expect(metadata.Get("Value").Tags).To(HaveKeyWithValue("key2", "Value2"))
			Expect(metadata.Get("Value").Anonymous).To(BeEmpty())
		})
	})

	When("when a struct has multiple fields with tags with multiple fields", func() {
		type testStruct struct {
			Value1 string `key2:"Value3" key4:"Value5"`
			Value6 string `key7:"Value8" key9:"Value10"`
		}

		It("should return the field names and their type with their tags metadata", func() {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata.Size()).To(Equal(2))
			Expect(metadata.Has("Value1")).To(BeTrue())
			Expect(metadata.Get("Value1").Type.Kind()).To(Equal(reflect.String))
			Expect(metadata.Get("Value1").Tags).To(HaveLen(2))
			Expect(metadata.Get("Value1").Tags).To(HaveKeyWithValue("key2", "Value3"))
			Expect(metadata.Get("Value1").Tags).To(HaveKeyWithValue("key4", "Value5"))
			Expect(metadata.Get("Value1").Anonymous).To(BeEmpty())
			Expect(metadata.Has("Value6")).To(BeTrue())
			Expect(metadata.Get("Value6").Type.Kind()).To(Equal(reflect.String))
			Expect(metadata.Get("Value6").Tags).To(HaveLen(2))
			Expect(metadata.Get("Value6").Tags).To(HaveKeyWithValue("key7", "Value8"))
			Expect(metadata.Get("Value6").Tags).To(HaveKeyWithValue("key9", "Value10"))
			Expect(metadata.Get("Value6").Anonymous).To(BeEmpty())
		})
	})

	When("when a struct has nested structs, all with fields and a tags", func() {
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

		It("should return the anonymous structs fields", func() {
			metadata := reflectutils.FieldsToMetadata[outerStruct]()
			Expect(metadata.Size()).To(Equal(4))
			Expect(metadata.Has("DeepField")).To(BeTrue())
			Expect(metadata.Get("DeepField").Type.Kind()).To(Equal(reflect.String))
			Expect(metadata.Get("DeepField").Tags).To(HaveLen(1))
			Expect(metadata.Get("DeepField").Tags).Should(HaveKeyWithValue("key", "Deep"))
			Expect(metadata.Get("DeepField").Anonymous).To(HaveLen(2))
			Expect(metadata.Get("DeepField").Anonymous[0]).To(Equal("embeddedStruct1"))
			Expect(metadata.Get("DeepField").Anonymous[1]).To(Equal("deepStruct"))
			Expect(metadata.Has("EmbeddedField1")).To(BeTrue())
			Expect(metadata.Get("EmbeddedField1").Type.Kind()).To(Equal(reflect.String))
			Expect(metadata.Get("EmbeddedField1").Tags).To(HaveLen(1))
			Expect(metadata.Get("EmbeddedField1").Tags).To(HaveKeyWithValue("key", "Embedded1"))
			Expect(metadata.Get("EmbeddedField1").Anonymous).To(HaveLen(1))
			Expect(metadata.Get("EmbeddedField1").Anonymous[0]).To(Equal("embeddedStruct1"))
			Expect(metadata.Has("EmbeddedField2")).To(BeTrue())
			Expect(metadata.Get("EmbeddedField2").Type.Kind()).To(Equal(reflect.String))
			Expect(metadata.Get("EmbeddedField2").Tags).To(HaveLen(1))
			Expect(metadata.Get("EmbeddedField2").Tags).To(HaveKeyWithValue("key", "Embedded2"))
			Expect(metadata.Get("EmbeddedField2").Anonymous).To(HaveLen(1))
			Expect(metadata.Get("EmbeddedField2").Anonymous[0]).To(Equal("embeddedStruct2"))
			Expect(metadata.Has("OuterField")).To(BeTrue())
			Expect(metadata.Get("OuterField").Type.Kind()).To(Equal(reflect.String))
			Expect(metadata.Get("OuterField").Tags).To(HaveLen(1))
			Expect(metadata.Get("OuterField").Tags).To(HaveKeyWithValue("key", "Outer"))
			Expect(metadata.Get("OuterField").Anonymous).To(BeEmpty())
		})
	})

	When("when a struct has a nested struct, both with the same field name", func() {
		type embeddedStruct struct {
			Field string
		}

		type outerStruct struct {
			embeddedStruct
			Field string
		}

		It("should panic", func() {
			Expect(func() {
				_ = reflectutils.FieldsToMetadata[outerStruct]()
			}).Should(PanicWith(ContainSubstring("field Field is ambiguous")))
		})
	})
})
