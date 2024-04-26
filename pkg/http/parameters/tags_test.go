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
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				QueryField1  string `urlQuery:"Query1" json:"-"`
				QueryField2  string `urlQuery:"Query2" json:"-"`
				HeaderField1 string `httpHeader:"Header1" json:"-"`
				HeaderField2 string `httpHeader:"Header2" json:"-"`
				PathField1   string `urlPath:"Path1" json:"-"`
				PathField2   string `urlPath:"Path2" json:"-"`
				JSONField1   string `json:"JSON1,omitempty"`
				JSONField2   string `json:"JSON2,omitempty"`
			}]()
			Expect(err).To(Not(HaveOccurred()))

			Expect(tagToLookupKeyToFieldName).To(HaveLen(3))

			Expect(tagToLookupKeyToFieldName).To(HaveKey(parameters.QueryTag))
			Expect(tagToLookupKeyToFieldName).To(HaveKey(parameters.HeaderTag))
			Expect(tagToLookupKeyToFieldName).To(HaveKey(parameters.PathTag))

			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]).To(HaveKey("query1"))
			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]["query1"].Name).To(Equal("QueryField1"))
			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]).To(HaveKey("query2"))
			Expect(tagToLookupKeyToFieldName[parameters.QueryTag]["query2"].Name).To(Equal("QueryField2"))

			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]).To(HaveKey("header1"))
			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]["header1"].Name).To(Equal("HeaderField1"))
			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]).To(HaveKey("header2"))
			Expect(tagToLookupKeyToFieldName[parameters.HeaderTag]["header2"].Name).To(Equal("HeaderField2"))

			Expect(tagToLookupKeyToFieldName[parameters.PathTag]).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName[parameters.PathTag]).To(HaveKey("Path1"))
			Expect(tagToLookupKeyToFieldName[parameters.PathTag]["Path1"].Name).To(Equal("PathField1"))
			Expect(tagToLookupKeyToFieldName[parameters.PathTag]).To(HaveKey("Path2"))
			Expect(tagToLookupKeyToFieldName[parameters.PathTag]["Path2"].Name).To(Equal("PathField2"))
		})

		It("should fail when validating a struct that has two fields with the same tag", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field1 string `urlQuery:"QueryField" json:"-"`
				Field2 int    `urlQuery:"QueryField" json:"-"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("tag 'urlQuery' with lookup key 'QueryField' is not unique"))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when validating a struct that has two fields with the same tag in different cases", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field1 string `urlQuery:"QueryField" json:"-"`
				Field2 int    `urlQuery:"qUeRyfIeLd" json:"-"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("tag 'urlQuery' with lookup key 'qUeRyfIeLd' is not unique"))
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
				Field string `urlQuery:"QueryField" json:"notRight"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("struct field 'Field' with tag 'urlQuery' must have accompanying tag json:\"-\""))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when validating a struct that has a tag that is empty", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[struct {
				Field string `urlQuery:"" json:"-"`
			}]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("tag 'urlQuery' with lookup key '' must adhere to the naming convention"))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when the generic isn't a struct", func() {
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[string]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the generic must be a struct"))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})

		It("should fail when the generic is a struct pointer", func() {
			type parameterStruct struct {
				Field string `urlQuery:"" json:"-"`
			}
			tagToLookupKeyToFieldName, err := parameters.ExtractAndValidateFieldTagLookupKeys[*parameterStruct]()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the generic must be a struct"))
			Expect(tagToLookupKeyToFieldName).To(BeNil())
		})
	})
})
