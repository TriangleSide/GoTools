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

package parameters_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/http/headers"
	"intelligence/pkg/http/parameters"
)

var _ = Describe("parameter tags", func() {
	When("a lookup key is verified against the naming convention", func() {
		DescribeTable("naming convention test cases",
			func(lookupKey string, expect bool) {
				Expect(parameters.TagLookupKeyFollowsNamingConvention(lookupKey)).To(Equal(expect),
					fmt.Sprintf("expect the validation of '%s' to be %v", lookupKey, expect))
			},
			// True cases.
			Entry("", headers.ContentType, true),
			Entry("", "a", true),
			Entry("", "A", true),
			Entry("", "aa", true),
			Entry("", "aA", true),
			Entry("", "a0", true),
			Entry("", "a-", true),
			Entry("", "a_", true),
			Entry("", "aaA-_", true),
			// False cases.
			Entry("", "", false),
			Entry("", "0", false),
			Entry("", "0a", false),
			Entry("", " name", false),
			Entry("", "name ", false),
			Entry("", "na me", false),
		)
	})

	When("a structs tags are extracted and validated", func() {
		It("should succeed when the tags are properly formatted", func() {
			type testStruct struct {
				QueryField1  string `urlQuery:"Query1" json:"-" otherTag1:"value"`
				QueryField2  string `urlQuery:"Query2" json:"-" otherTag1:"value"`
				HeaderField1 string `httpHeader:"Header1" json:"-" otherTag2:"value1"`
				HeaderField2 string `httpHeader:"Header2" json:"-" otherTag2:"value2"`
				PathField1   string `urlPath:"Path1" json:"-" otherTag3:""`
				PathField2   string `urlPath:"Path2" json:"-" otherTag4:"!@#$%^&*()"`
				JSONField1   string `json:"JSON1,omitempty"`
				JSONField2   string `json:"JSON2,omitempty"`
			}

			var err error
			var tagToLookupKeyToFieldName map[parameters.Tag]map[string]string

			for i := 0; i < 3; i++ {
				tagToLookupKeyToFieldName, err = parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
				Expect(err).To(Not(HaveOccurred()))
			}

			Expect(tagToLookupKeyToFieldName).To(HaveLen(3))

			Expect(tagToLookupKeyToFieldName).To(HaveKey(parameters.QueryTag))
			Expect(tagToLookupKeyToFieldName).To(HaveKey(parameters.HeaderTag))
			Expect(tagToLookupKeyToFieldName).To(HaveKey(parameters.PathTag))

			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]).To(HaveKey("query1"))
			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]["query1"]).To(Equal("QueryField1"))
			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]).To(HaveKey("query2"))
			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]["query2"]).To(Equal("QueryField2"))

			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]).To(HaveKey("header1"))
			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]["header1"]).To(Equal("HeaderField1"))
			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]).To(HaveKey("header2"))
			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]["header2"]).To(Equal("HeaderField2"))

			Expect(tagToLookupKeyToFieldName[parameters.PathTag]).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName[parameters.PathTag]).To(HaveKey("Path1"))
			Expect(tagToLookupKeyToFieldName[parameters.PathTag]["Path1"]).To(Equal("PathField1"))
			Expect(tagToLookupKeyToFieldName[parameters.PathTag]).To(HaveKey("Path2"))
			Expect(tagToLookupKeyToFieldName[parameters.PathTag]["Path2"]).To(Equal("PathField2"))
		})

		It("should fail when validating a struct that has two fields with the same tag", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field1 string `urlQuery:"QueryField" json:"-"`
				Field2 int    `urlQuery:"QueryField" json:"-"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not unique"))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when validating a struct that has two fields with the same tag in different cases", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field1 string `urlQuery:"QueryField" json:"-"`
				Field2 int    `urlQuery:"qUeRyfIeLd" json:"-"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not unique"))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when validating a struct that has overlapping tags", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field string `urlQuery:"QueryField" httpHeader:"HeaderField" json:"-"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("there can only be one encoding tag on the field 'Field'"))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when validating a struct that has no accompanying json tag", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field string `urlQuery:"QueryField"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("struct field 'Field' with tag 'urlQuery' must have accompanying tag json:\"-\""))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when validating a struct that has an accompanying json tag with the wrong format", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field string `httpHeader:"QueryField" json:"notRight"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("struct field 'Field' with tag 'httpHeader' must have accompanying tag json:\"-\""))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when validating a struct that has a tag that is empty", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field string `urlPath:"" json:"-"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("tag 'urlPath' with lookup key '' must adhere to the naming convention"))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should panic when the generic isn't a struct", func() {
			Expect(func() {
				_, _ = parameters.ExtractAndValidateFieldTagLookupKeys[string]()
			}).Should(Panic())
		})

		It("should fail when the generic is a struct pointer", func() {
			type parameterStruct struct {
				Field string `urlQuery:"" json:"-"`
			}
			Expect(func() {
				_, _ = parameters.ExtractAndValidateFieldTagLookupKeys[*parameterStruct]()
			}).Should(Panic())
		})
	})
})
