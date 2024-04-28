package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/validation"
)

var _ = Describe("path", func() {
	DescribeTable("path validation",
		func(path string, errorMsg string) {
			errCheck := func(err error) {
				if len(errorMsg) > 0 {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(errorMsg))
				} else {
					Expect(err).To(Succeed())
				}
			}

			type testStructRef struct {
				Path string `validate:"api_path"`
			}
			errCheck(validation.Struct(&testStructRef{Path: path}))

			type testStructPtr struct {
				Path *string `validate:"api_path"`
			}
			errCheck(validation.Struct(&testStructPtr{Path: &path}))
		},
		Entry("root path", "/", ""),
		Entry("sub paths", "/a/b/c/1/2/3", ""),
		Entry("sub paths with a param", "/a/{b}/c", ""),
		Entry("empty string", "", "path cannot be empty"),
		Entry("invalid characters", "/+", "path contains invalid characters"),
		Entry("invalid characters", " /a", "path contains invalid characters"),
		Entry("invalid characters", "/a ", "path contains invalid characters"),
		Entry("end with /", "/a/", "path cannot end with '/'"),
		Entry("start with /", "a/b", "path must start with '/'"),
		Entry("start with /", "a", "path must start with '/'"),
		Entry("empty path part", "/a//b", "path parts cannot be empty"),
		Entry("empty path part", "//a", "path parts cannot be empty"),
		Entry("no matching }", "/a/{b", "path parameters must start with '{' and end with '}'"),
		Entry("no matching {", "/a/b}", "path parameters must start with '{' and end with '}'"),
		Entry("many {", "/a/{{b}", "path parameters have only one '{' and '}'"),
		Entry("many }", "/a/{b}}", "path parameters have only one '{' and '}'"),
		Entry("empty param", "/a/{}", "path parameters cannot be empty"),
		Entry("reused param", "/a/{b}/{b}", "path part must be unique"),
		Entry("reused part", "/a/a", "path part must be unique"),
		Entry("reused part", "/a/b/a", "path part must be unique"),
	)

	When("path validation is done on a reference field that it not a string", func() {
		type testStruct struct {
			Path int `validate:"api_path"`
		}

		var (
			test testStruct
		)

		BeforeEach(func() {
			test = testStruct{
				Path: 1,
			}
		})

		It("should fail the validation", func() {
			err := validation.Struct(&test)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path must be a string"))
		})
	})

	When("path validation is done on a pointer field that it not a string", func() {
		type testStruct struct {
			Path *int `validate:"api_path"`
		}

		var (
			test testStruct
		)

		BeforeEach(func() {
			i := 0
			test = testStruct{
				Path: &i,
			}
		})

		It("should fail the validation", func() {
			err := validation.Struct(&test)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("path must be a string"))
		})
	})
})
