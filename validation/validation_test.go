package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/validation"
)

var _ = Describe("validation", func() {
	When("a struct has a field called Value with a validation rule of gte=0", func() {
		type testStruct struct {
			Value int `validate:"gte=0"`
		}

		var (
			test testStruct
		)

		BeforeEach(func() {
			test = testStruct{}
		})

		When("when the struct field has a valid value", func() {
			BeforeEach(func() {
				test.Value = 1
			})

			It("should succeed when using the Validate method when passed by value", func() {
				Expect(validation.Validate(test)).To(Succeed())
			})

			It("should succeed when using the Validate method when passed by reference", func() {
				Expect(validation.Validate(&test)).To(Succeed())
			})
		})

		When("when the struct field has an invalid value", func() {
			BeforeEach(func() {
				test.Value = -1
			})

			It("should return an error when using the Validate method when passed by value", func() {
				Expect(validation.Validate(test)).To(HaveOccurred())
			})

			It("should return an error when using the Validate method when passed by reference", func() {
				Expect(validation.Validate(&test)).To(HaveOccurred())
			})
		})
	})

	When("a struct has a field called Value with no validation rules", func() {
		type testStruct struct {
			Value int
		}

		var (
			test testStruct
		)

		BeforeEach(func() {
			test = testStruct{
				Value: 0,
			}
		})

		It("should succeed when using the Validate method when passed by value", func() {
			Expect(validation.Validate(test)).To(Succeed())
		})

		It("should succeed when using the Validate method when passed by reference", func() {
			Expect(validation.Validate(&test)).To(Succeed())
		})
	})

	When("an argument is passed to the Validate function that is not a struct", func() {
		var (
			test int
		)

		BeforeEach(func() {
			test = 0
		})

		It("should return an error when using the Validate method when passed by value", func() {
			Expect(validation.Validate(test)).To(HaveOccurred())
		})

		It("should return an error when using the Validate method when passed by reference", func() {
			Expect(validation.Validate(&test)).To(HaveOccurred())
		})
	})

	When("nil passed to the Validate function", func() {
		It("should return an error", func() {
			Expect(validation.Validate(nil)).To(HaveOccurred())
		})
	})
})
