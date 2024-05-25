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
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/utils/reflect"
)

type unmarshallTestStruct struct {
	Value string
}

func (t *unmarshallTestStruct) UnmarshalText(text []byte) error {
	t.Value = string(text)
	return nil
}

var _ = Describe("assign a struct field with a string value", func() {
	type testDeepEmbeddedStruct struct {
		DeepEmbeddedValue *string
	}

	type testEmbeddedStruct struct {
		testDeepEmbeddedStruct
		EmbeddedValue string
	}

	type testInternalStruct struct {
		Value string `json:"value"`
	}

	type testStruct struct {
		testEmbeddedStruct

		StringValue     string
		IntValue        int
		UintValue       uint
		FloatValue      float64
		BoolValue       bool
		StructValue     testInternalStruct
		MapValue        map[string]testInternalStruct
		UnmarshallValue unmarshallTestStruct
		TimeValue       time.Time

		StringPtrValue     *string
		IntPtrValue        *int
		UintPtrValue       *uint
		FloatPtrValue      *float64
		BoolPtrValue       *bool
		StructPtrValue     *testInternalStruct
		MapPtrValue        *map[string]testInternalStruct
		UnmarshallPtrValue *unmarshallTestStruct
		TimePtrValue       *time.Time

		ListStringValue []string
		ListIntValue    []int
		ListFloatValue  []float64
		ListBoolValue   []bool
		ListStructValue []testInternalStruct

		ListStringPtrValue []*string
		ListIntPtrValue    []*int
		ListFloatPtrValue  []*float64
		ListBoolPtrValue   []*bool
		ListStructPtrValue []*testInternalStruct

		UnhandledValue uintptr
	}

	It("should panic when setting the value on an object that is not a struct", func() {
		Expect(func() {
			_ = reflect.AssignToField(new(int), "StringValue", "test")
		}).To(PanicWith(ContainSubstring("obj must be a pointer to a struct")))
	})

	When("a test struct is initialized with no assigned values", func() {
		var (
			values *testStruct
		)

		BeforeEach(func() {
			values = &testStruct{}
		})

		Context("embedded value assignments", func() {
			It("should set the EmbeddedValue field", func() {
				const setValue = "test"
				Expect(reflect.AssignToField(values, "EmbeddedValue", setValue)).To(Succeed())
				Expect(values.EmbeddedValue).To(Equal(setValue))
			})

			It("should set the DeepEmbeddedValue field", func() {
				const setValue = "test"
				Expect(reflect.AssignToField(values, "DeepEmbeddedValue", setValue)).To(Succeed())
				Expect(*values.DeepEmbeddedValue).To(Equal(setValue))
			})
		})

		Context("normal value assignments", func() {
			It("should set the StringValue field", func() {
				const setValue = "test"
				Expect(reflect.AssignToField(values, "StringValue", setValue)).To(Succeed())
				Expect(values.StringValue).To(Equal(setValue))
			})

			It("should set the IntValue field", func() {
				const setValue = "-123"
				Expect(reflect.AssignToField(values, "IntValue", setValue)).To(Succeed())
				Expect(values.IntValue).To(BeNumerically("==", -123))
			})

			It("should set the UintValue field", func() {
				const setValue = "123"
				Expect(reflect.AssignToField(values, "UintValue", setValue)).To(Succeed())
				Expect(values.UintValue).To(BeNumerically("==", 123))
			})

			It("should set the FloatValue field", func() {
				const setValue = "123.456"
				Expect(reflect.AssignToField(values, "FloatValue", setValue)).To(Succeed())
				Expect(values.FloatValue).To(BeNumerically("~", 123.456, 0.001))
			})

			It("should set the BoolValue field", func() {
				const setValue = "true"
				Expect(reflect.AssignToField(values, "BoolValue", setValue)).To(Succeed())
				Expect(values.BoolValue).To(BeTrue())
			})

			It("should set the BoolValue field with integer", func() {
				const setValue = "1"
				Expect(reflect.AssignToField(values, "BoolValue", setValue)).To(Succeed())
				Expect(values.BoolValue).To(BeTrue())
			})

			It("should set the StructValue field", func() {
				const setValue = `{"value":"nested"}`
				Expect(reflect.AssignToField(values, "StructValue", setValue)).To(Succeed())
				Expect(values.StructValue.Value).To(Equal("nested"))
			})

			It("should set the MapValue field", func() {
				const setValue = `{"key1":{"value":"value1"}}`
				Expect(reflect.AssignToField(values, "MapValue", setValue)).To(Succeed())
				Expect(values.MapValue["key1"].Value).To(Equal("value1"))
			})

			It("should set the UnmarshallValue field", func() {
				const setValue = "custom text"
				Expect(reflect.AssignToField(values, "UnmarshallValue", setValue)).To(Succeed())
				Expect(values.UnmarshallValue.Value).To(Equal("custom text"))
			})

			It("should set the TimeValue field", func() {
				const setValue = "2024-01-01T12:34:56Z"
				Expect(reflect.AssignToField(values, "TimeValue", setValue)).To(Succeed())
				expectedTime, _ := time.Parse(time.RFC3339, setValue)
				Expect(values.TimeValue).To(Equal(expectedTime))
			})
		})

		Context("pointer value assignments", func() {
			It("should set the StringPtrValue field", func() {
				const setValue = "test ptr"
				Expect(reflect.AssignToField(values, "StringPtrValue", setValue)).To(Succeed())
				Expect(values.StringPtrValue).NotTo(BeNil())
				Expect(*values.StringPtrValue).To(Equal(setValue))
			})

			It("should set the IntPtrValue field", func() {
				const setValue = "-321"
				Expect(reflect.AssignToField(values, "IntPtrValue", setValue)).To(Succeed())
				Expect(values.IntPtrValue).NotTo(BeNil())
				Expect(*values.IntPtrValue).To(BeNumerically("==", -321))
			})

			It("should set the UintPtrValue field", func() {
				const setValue = "321"
				Expect(reflect.AssignToField(values, "UintPtrValue", setValue)).To(Succeed())
				Expect(values.UintPtrValue).NotTo(BeNil())
				Expect(*values.UintPtrValue).To(BeNumerically("==", 321))
			})

			It("should set the FloatPtrValue field", func() {
				const setValue = "123.456"
				Expect(reflect.AssignToField(values, "FloatPtrValue", setValue)).To(Succeed())
				Expect(values.FloatPtrValue).NotTo(BeNil())
				Expect(*values.FloatPtrValue).To(BeNumerically("~", 123.456, 0.001))
			})

			It("should set the BoolPtrValue field", func() {
				const setValue = "true"
				Expect(reflect.AssignToField(values, "BoolPtrValue", setValue)).To(Succeed())
				Expect(values.BoolPtrValue).NotTo(BeNil())
				Expect(*values.BoolPtrValue).To(BeTrue())
			})

			It("should set the BoolPtrValue field with integer", func() {
				const setValue = "true"
				Expect(reflect.AssignToField(values, "BoolPtrValue", setValue)).To(Succeed())
				Expect(values.BoolPtrValue).NotTo(BeNil())
				Expect(*values.BoolPtrValue).To(BeTrue())
			})

			It("should set the StructPtrValue field", func() {
				const setValue = `{"value":"nested ptr"}`
				Expect(reflect.AssignToField(values, "StructPtrValue", setValue)).To(Succeed())
				Expect(values.StructPtrValue).NotTo(BeNil())
				Expect(values.StructPtrValue.Value).To(Equal("nested ptr"))
			})

			It("should set the MapPtrValue field", func() {
				const setValue = `{"key1":{"value":"value1 ptr"}}`
				Expect(reflect.AssignToField(values, "MapPtrValue", setValue)).To(Succeed())
				Expect(values.MapPtrValue).NotTo(BeNil())
				Expect((*values.MapPtrValue)["key1"].Value).To(Equal("value1 ptr"))
			})

			It("should set the UnmarshallPtrValue field", func() {
				const setValue = "custom text ptr"
				Expect(reflect.AssignToField(values, "UnmarshallPtrValue", setValue)).To(Succeed())
				Expect(values.UnmarshallPtrValue).NotTo(BeNil())
				Expect(values.UnmarshallPtrValue.Value).To(Equal("custom text ptr"))
			})

			It("should set the TimePtrValue field", func() {
				const setValue = "2024-01-02T12:34:56Z"
				Expect(reflect.AssignToField(values, "TimePtrValue", setValue)).To(Succeed())
				expectedTime, _ := time.Parse(time.RFC3339, setValue)
				Expect(*values.TimePtrValue).To(Equal(expectedTime))
			})
		})

		Context("list value assignments", func() {
			It("should set the ListStringValue field", func() {
				const setValue = `["one", "two", "three"]`
				Expect(reflect.AssignToField(values, "ListStringValue", setValue)).To(Succeed())
				Expect(values.ListStringValue).To(Equal([]string{"one", "two", "three"}))
			})

			It("should set the ListIntValue field", func() {
				const setValue = "[1, 2, 3]"
				Expect(reflect.AssignToField(values, "ListIntValue", setValue)).To(Succeed())
				Expect(values.ListIntValue).To(Equal([]int{1, 2, 3}))
			})

			It("should set the ListFloatValue field", func() {
				const setValue = "[1.1, 2.2, 3.3]"
				Expect(reflect.AssignToField(values, "ListFloatValue", setValue)).To(Succeed())
				Expect(values.ListFloatValue).To(Equal([]float64{1.1, 2.2, 3.3}))
			})

			It("should set the ListBoolValue field", func() {
				const setValue = "[true, false, true]"
				Expect(reflect.AssignToField(values, "ListBoolValue", setValue)).To(Succeed())
				Expect(values.ListBoolValue).To(Equal([]bool{true, false, true}))
			})

			It("should set the ListStructValue field", func() {
				const setValue = `[{"value":"nested1"}, {"value":"nested2"}, {"value":"nested3"}]`
				Expect(reflect.AssignToField(values, "ListStructValue", setValue)).To(Succeed())
				Expect(values.ListStructValue).To(HaveLen(3))
				Expect(values.ListStructValue[0].Value).To(Equal("nested1"))
				Expect(values.ListStructValue[1].Value).To(Equal("nested2"))
				Expect(values.ListStructValue[2].Value).To(Equal("nested3"))
			})
		})

		Context("list pointer value assignments", func() {
			It("should set the ListStringPtrValue field", func() {
				const setValue = `["one", "two", "three"]`
				Expect(reflect.AssignToField(values, "ListStringPtrValue", setValue)).To(Succeed())
				Expect(values.ListStringPtrValue).To(HaveLen(3))
				Expect(*values.ListStringPtrValue[0]).To(Equal("one"))
				Expect(*values.ListStringPtrValue[1]).To(Equal("two"))
				Expect(*values.ListStringPtrValue[2]).To(Equal("three"))
			})

			It("should set the ListIntPtrValue field", func() {
				const setValue = "[1, 2, 3]"
				Expect(reflect.AssignToField(values, "ListIntPtrValue", setValue)).To(Succeed())
				Expect(values.ListIntPtrValue).To(HaveLen(3))
				Expect(*values.ListIntPtrValue[0]).To(Equal(1))
				Expect(*values.ListIntPtrValue[1]).To(Equal(2))
				Expect(*values.ListIntPtrValue[2]).To(Equal(3))
			})

			It("should set the ListFloatPtrValue field", func() {
				const setValue = "[1.1, 2.2, 3.3]"
				Expect(reflect.AssignToField(values, "ListFloatPtrValue", setValue)).To(Succeed())
				Expect(values.ListFloatPtrValue).To(HaveLen(3))
				Expect(*values.ListFloatPtrValue[0]).To(Equal(1.1))
				Expect(*values.ListFloatPtrValue[1]).To(Equal(2.2))
				Expect(*values.ListFloatPtrValue[2]).To(Equal(3.3))
			})

			It("should set the ListBoolPtrValue field", func() {
				const setValue = "[true, false, true]"
				Expect(reflect.AssignToField(values, "ListBoolPtrValue", setValue)).To(Succeed())
				Expect(values.ListBoolPtrValue).To(HaveLen(3))
				Expect(*values.ListBoolPtrValue[0]).To(BeTrue())
				Expect(*values.ListBoolPtrValue[1]).To(BeFalse())
				Expect(*values.ListBoolPtrValue[2]).To(BeTrue())
			})

			It("should set the ListStructPtrValue field", func() {
				const setValue = `[{"value":"nested1"}, {"value":"nested2"}, {"value":"nested3"}]`
				Expect(reflect.AssignToField(values, "ListStructPtrValue", setValue)).To(Succeed())
				Expect(values.ListStructPtrValue).To(HaveLen(3))
				Expect(values.ListStructPtrValue[0]).NotTo(BeNil())
				Expect(values.ListStructPtrValue[1]).NotTo(BeNil())
				Expect(values.ListStructPtrValue[2]).NotTo(BeNil())
				Expect(values.ListStructPtrValue[0].Value).To(Equal("nested1"))
				Expect(values.ListStructPtrValue[1].Value).To(Equal("nested2"))
				Expect(values.ListStructPtrValue[2].Value).To(Equal("nested3"))
			})
		})

		Context("failure cases", func() {
			It("should fail to set a string value to an int field", func() {
				const setValue = "not an integer"
				err := reflect.AssignToField(values, "IntValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("strconv.ParseInt"))
			})

			It("should fail to set an int that is bigger then 32 bits into an int field", func() {
				setValue := strings.Repeat("1", 100)
				err := reflect.AssignToField(values, "IntValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("strconv.ParseInt"))
			})

			It("should fail to set a negative int value to a uint", func() {
				const setValue = "-123"
				err := reflect.AssignToField(values, "UintValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("strconv.ParseUint"))
			})

			It("should fail to set a float value to an int field", func() {
				const setValue = "123.456"
				err := reflect.AssignToField(values, "IntValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("strconv.ParseInt"))
			})

			It("should fail to set a string value to a float field", func() {
				const setValue = "not a float"
				err := reflect.AssignToField(values, "FloatValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("strconv.ParseFloat"))
			})

			It("should fail to set an int value to a bool field", func() {
				const setValue = "2"
				err := reflect.AssignToField(values, "BoolValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("strconv.ParseBool"))
			})

			It("should fail to set a malformed JSON string to a struct field", func() {
				const setValue = "not a json object"
				err := reflect.AssignToField(values, "StructValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("json unmarshal error"))
			})

			It("should fail to set a string value to a map field", func() {
				const setValue = "not a json object"
				err := reflect.AssignToField(values, "MapValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("json unmarshal error"))
			})

			It("should panic when trying to set a field that does not exist", func() {
				const setValue = "some value"
				Expect(func() {
					_ = reflect.AssignToField(values, "NonExistentField", setValue)
				}).To(PanicWith(ContainSubstring("no field 'NonExistentField' in struct")))
			})

			It("should fail to set non-integer values in an integer list", func() {
				const setValue = `["one", "two", "three"]`
				err := reflect.AssignToField(values, "ListIntValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("json unmarshal error"))
			})

			It("should fail to set non-integer values in an pointer integer list", func() {
				const setValue = `["one", "two", "three"]`
				err := reflect.AssignToField(values, "ListIntPtrValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("json unmarshal error"))
			})

			It("should fail to set non-boolean values in a boolean list", func() {
				const setValue = `["true", "false", "maybe"]`
				err := reflect.AssignToField(values, "ListBoolValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("json unmarshal error"))
			})

			It("should fail to set incorrectly formatted JSON in a list of structs", func() {
				const setValue = `[{"value":"nested1"}, {"value":}]` // Malformed JSON
				err := reflect.AssignToField(values, "ListStructValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("json unmarshal error"))
			})

			It("should fail to set a list of strings in a list of floats", func() {
				const setValue = `["1.1", "two", "3.3"]`
				err := reflect.AssignToField(values, "ListFloatValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("json unmarshal error"))
			})

			It("should fail to set an incorrectly formatted time string to a time field", func() {
				const setValue = "not a time string"
				err := reflect.AssignToField(values, "TimeValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parsing time"))
			})

			It("should fail to set a non-time string to a pointer time field", func() {
				const setValue = "this is not a time"
				err := reflect.AssignToField(values, "TimePtrValue", setValue)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parsing time"))
			})

			It("should fail when setting an unhandled value", func() {
				err := reflect.AssignToField(values, "UnhandledValue", "test")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported field type"))
			})

		})
	})
})
