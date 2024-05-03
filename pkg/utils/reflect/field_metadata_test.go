package reflect_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	reflectutils "intelligence/pkg/utils/reflect"
)

var _ = Describe("struct field metadata", func() {
	It("should panic if the type is not a struct", func() {
		Expect(func() {
			_, _ = reflectutils.FieldsToMetadata[int]()
		}).Should(Panic())
	})

	It("should panic if the type is a pointer to a struct", func() {
		Expect(func() {
			_, _ = reflectutils.FieldsToMetadata[*struct{}]()
		}).Should(Panic())
	})

	It("should return an empty map for an empty struct", func() {
		metadata, err := reflectutils.FieldsToMetadata[struct{}]()
		Expect(err).ToNot(HaveOccurred())
		Expect(metadata).To(BeEmpty())
	})

	It("should succeed if done twice on the same struct type", func() {
		type testStruct struct{}
		metadata1, err1 := reflectutils.FieldsToMetadata[testStruct]()
		Expect(err1).ToNot(HaveOccurred())
		Expect(metadata1).To(BeEmpty())
		metadata2, err2 := reflectutils.FieldsToMetadata[testStruct]()
		Expect(err2).ToNot(HaveOccurred())
		Expect(metadata2).To(BeEmpty())
	})

	When("when a struct has a string field called Value and no tag", func() {
		type testStruct struct {
			Value string
		}

		It("return the field name and its type with no metadata", func() {
			metadata, err := reflectutils.FieldsToMetadata[testStruct]()
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata).To(HaveLen(1))
			Expect(metadata).To(HaveKey("Value"))
			Expect(metadata["Value"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["Value"].Tags).To(HaveLen(0))
		})
	})

	When("when a struct has a string field called Value and a tag", func() {
		type testStruct struct {
			Value int `key:"Value"`
		}

		It("return the field name and its type with no metadata", func() {
			metadata, err := reflectutils.FieldsToMetadata[testStruct]()
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata).To(HaveLen(1))
			Expect(metadata).To(HaveKey("Value"))
			Expect(metadata["Value"].Type.Kind()).To(Equal(reflect.Int))
			Expect(metadata["Value"].Tags).To(HaveLen(1))
			Expect(metadata["Value"].Tags).Should(HaveKeyWithValue("key", "Value"))
		})
	})

	When("when a struct has a string field called Value and a tag with multiple fields", func() {
		type testStruct struct {
			Value float32 `key1:"Value1" key2:"Value2"`
		}

		It("return the field name and its type with no metadata", func() {
			metadata, err := reflectutils.FieldsToMetadata[testStruct]()
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata).To(HaveLen(1))
			Expect(metadata).To(HaveKey("Value"))
			Expect(metadata["Value"].Type.Kind()).To(Equal(reflect.Float32))
			Expect(metadata["Value"].Tags).To(HaveLen(2))
			Expect(metadata["Value"].Tags).Should(HaveKeyWithValue("key1", "Value1"))
			Expect(metadata["Value"].Tags).Should(HaveKeyWithValue("key2", "Value2"))
		})
	})

	When("when a struct has multiple fields with tags with multiple fields", func() {
		type testStruct struct {
			Value1 string `key2:"Value3" key4:"Value5"`
			Value6 string `key7:"Value8" key9:"Value10"`
		}

		It("return the field name and its type with no metadata", func() {
			metadata, err := reflectutils.FieldsToMetadata[testStruct]()
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata).To(HaveLen(2))
			Expect(metadata).To(HaveKey("Value1"))
			Expect(metadata["Value1"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["Value1"].Tags).To(HaveLen(2))
			Expect(metadata["Value1"].Tags).Should(HaveKeyWithValue("key2", "Value3"))
			Expect(metadata["Value1"].Tags).Should(HaveKeyWithValue("key4", "Value5"))
			Expect(metadata).To(HaveKey("Value6"))
			Expect(metadata["Value6"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["Value6"].Tags).To(HaveLen(2))
			Expect(metadata["Value6"].Tags).Should(HaveKeyWithValue("key7", "Value8"))
			Expect(metadata["Value6"].Tags).Should(HaveKeyWithValue("key9", "Value10"))
		})
	})
})
