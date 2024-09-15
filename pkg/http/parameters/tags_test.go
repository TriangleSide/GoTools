package parameters_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/datastructures"
	"github.com/TriangleSide/GoBase/pkg/http/headers"
	"github.com/TriangleSide/GoBase/pkg/http/parameters"
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
			var tagToLookupKeyToFieldName datastructures.ReadOnlyMap[parameters.Tag, parameters.LookupKeyToFieldName]

			for i := 0; i < 3; i++ {
				tagToLookupKeyToFieldName, err = parameters.ExtractAndValidateFieldTagLookupKeys[testStruct]()
				Expect(err).To(Not(HaveOccurred()))
			}

			Expect(tagToLookupKeyToFieldName.Size()).To(Equal(3))

			Expect(tagToLookupKeyToFieldName.Has(parameters.QueryTag))
			Expect(tagToLookupKeyToFieldName.Has(parameters.HeaderTag))
			Expect(tagToLookupKeyToFieldName.Has(parameters.PathTag))

			Expect(tagToLookupKeyToFieldName.Get(parameters.QueryTag)).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName.Get(parameters.QueryTag)).To(HaveKey("query1"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.QueryTag)["query1"]).To(Equal("QueryField1"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.QueryTag)).To(HaveKey("query2"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.QueryTag)["query2"]).To(Equal("QueryField2"))

			Expect(tagToLookupKeyToFieldName.Get(parameters.HeaderTag)).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName.Get(parameters.HeaderTag)).To(HaveKey("header1"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.HeaderTag)["header1"]).To(Equal("HeaderField1"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.HeaderTag)).To(HaveKey("header2"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.HeaderTag)["header2"]).To(Equal("HeaderField2"))

			Expect(tagToLookupKeyToFieldName.Get(parameters.PathTag)).To(HaveLen(2))
			Expect(tagToLookupKeyToFieldName.Get(parameters.PathTag)).To(HaveKey("Path1"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.PathTag)["Path1"]).To(Equal("PathField1"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.PathTag)).To(HaveKey("Path2"))
			Expect(tagToLookupKeyToFieldName.Get(parameters.PathTag)["Path2"]).To(Equal("PathField2"))
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
			}).Should(PanicWith(ContainSubstring("type must be a struct")))
		})

		It("should fail when the generic is a struct pointer", func() {
			type parameterStruct struct {
				Field string `urlQuery:"" json:"-"`
			}
			Expect(func() {
				_, _ = parameters.ExtractAndValidateFieldTagLookupKeys[*parameterStruct]()
			}).Should(PanicWith(ContainSubstring("type must be a struct")))
		})
	})
})
