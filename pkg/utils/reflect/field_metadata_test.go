// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

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
			_ = reflectutils.FieldsToMetadata[int]()
		}).Should(Panic())
	})

	It("should panic if the type is a pointer to a struct", func() {
		Expect(func() {
			_ = reflectutils.FieldsToMetadata[*struct{}]()
		}).Should(Panic())
	})

	It("should return an empty map for an empty struct", func() {
		metadata := reflectutils.FieldsToMetadata[struct{}]()
		Expect(metadata).To(BeEmpty())
	})

	It("should succeed if done many times on the same struct type", func() {
		type testStruct struct{}
		for i := 0; i < 2; i++ {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata).To(BeEmpty())
		}
	})

	When("when a struct has a string field called Value and no tag", func() {
		type testStruct struct {
			Value string
		}

		It("should return the field name and its type with no metadata", func() {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata).To(HaveLen(1))
			Expect(metadata).To(HaveKey("Value"))
			Expect(metadata["Value"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["Value"].Tags).To(BeEmpty())
			Expect(metadata["Value"].Anonymous).To(BeEmpty())
		})
	})

	When("when a struct has a string field called Value and a tag", func() {
		type testStruct struct {
			Value int `key:"Value"`
		}

		It("should return the field name and its type with the tag metadata", func() {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata).To(HaveLen(1))
			Expect(metadata).To(HaveKey("Value"))
			Expect(metadata["Value"].Type.Kind()).To(Equal(reflect.Int))
			Expect(metadata["Value"].Tags).To(HaveLen(1))
			Expect(metadata["Value"].Tags).To(HaveKeyWithValue("key", "Value"))
			Expect(metadata["Value"].Anonymous).To(BeEmpty())
		})
	})

	When("when a struct has a string field called Value and a tag with multiple fields", func() {
		type testStruct struct {
			Value float32 `key1:"Value1" key2:"Value2"`
		}

		It("should return the field name and its type with the tags metadata", func() {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata).To(HaveLen(1))
			Expect(metadata).To(HaveKey("Value"))
			Expect(metadata["Value"].Type.Kind()).To(Equal(reflect.Float32))
			Expect(metadata["Value"].Tags).To(HaveLen(2))
			Expect(metadata["Value"].Tags).To(HaveKeyWithValue("key1", "Value1"))
			Expect(metadata["Value"].Tags).To(HaveKeyWithValue("key2", "Value2"))
			Expect(metadata["Value"].Anonymous).To(BeEmpty())
		})
	})

	When("when a struct has multiple fields with tags with multiple fields", func() {
		type testStruct struct {
			Value1 string `key2:"Value3" key4:"Value5"`
			Value6 string `key7:"Value8" key9:"Value10"`
		}

		It("should return the field names and their type with their tags metadata", func() {
			metadata := reflectutils.FieldsToMetadata[testStruct]()
			Expect(metadata).To(HaveLen(2))
			Expect(metadata).To(HaveKey("Value1"))
			Expect(metadata["Value1"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["Value1"].Tags).To(HaveLen(2))
			Expect(metadata["Value1"].Tags).To(HaveKeyWithValue("key2", "Value3"))
			Expect(metadata["Value1"].Tags).To(HaveKeyWithValue("key4", "Value5"))
			Expect(metadata["Value1"].Anonymous).To(BeEmpty())
			Expect(metadata).To(HaveKey("Value6"))
			Expect(metadata["Value6"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["Value6"].Tags).To(HaveLen(2))
			Expect(metadata["Value6"].Tags).To(HaveKeyWithValue("key7", "Value8"))
			Expect(metadata["Value6"].Tags).To(HaveKeyWithValue("key9", "Value10"))
			Expect(metadata["Value6"].Anonymous).To(BeEmpty())
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
			Expect(metadata).To(HaveLen(4))
			Expect(metadata).To(HaveKey("DeepField"))
			Expect(metadata["DeepField"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["DeepField"].Tags).To(HaveLen(1))
			Expect(metadata["DeepField"].Tags).Should(HaveKeyWithValue("key", "Deep"))
			Expect(metadata["DeepField"].Anonymous).To(HaveLen(2))
			Expect(metadata["DeepField"].Anonymous[0]).To(Equal("embeddedStruct1"))
			Expect(metadata["DeepField"].Anonymous[1]).To(Equal("deepStruct"))
			Expect(metadata).To(HaveKey("EmbeddedField1"))
			Expect(metadata["EmbeddedField1"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["EmbeddedField1"].Tags).To(HaveLen(1))
			Expect(metadata["EmbeddedField1"].Tags).To(HaveKeyWithValue("key", "Embedded1"))
			Expect(metadata["EmbeddedField1"].Anonymous).To(HaveLen(1))
			Expect(metadata["EmbeddedField1"].Anonymous[0]).To(Equal("embeddedStruct1"))
			Expect(metadata).To(HaveKey("EmbeddedField2"))
			Expect(metadata["EmbeddedField2"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["EmbeddedField2"].Tags).To(HaveLen(1))
			Expect(metadata["EmbeddedField2"].Tags).To(HaveKeyWithValue("key", "Embedded2"))
			Expect(metadata["EmbeddedField2"].Anonymous).To(HaveLen(1))
			Expect(metadata["EmbeddedField2"].Anonymous[0]).To(Equal("embeddedStruct2"))
			Expect(metadata).To(HaveKey("OuterField"))
			Expect(metadata["OuterField"].Type.Kind()).To(Equal(reflect.String))
			Expect(metadata["OuterField"].Tags).To(HaveLen(1))
			Expect(metadata["OuterField"].Tags).To(HaveKeyWithValue("key", "Outer"))
			Expect(metadata["OuterField"].Anonymous).To(BeEmpty())
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
			}).Should(Panic())
		})
	})
})
