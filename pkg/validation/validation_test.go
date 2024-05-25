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

package validation_test

import (
	"github.com/go-playground/validator/v10"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/validation"
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

			It("should succeed when using the Struct method when passed by value", func() {
				Expect(validation.Struct(test)).To(Succeed())
			})

			It("should succeed when using the Struct method when passed by reference", func() {
				Expect(validation.Struct(&test)).To(Succeed())
			})
		})

		When("when the struct field has an invalid value", func() {
			BeforeEach(func() {
				test.Value = -1
			})

			It("should return an error when using the Struct method when passed by value", func() {
				err := validation.Struct(test)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("validation failed on field 'Value' with validator 'gte' and parameter(s) '0'"))
			})

			It("should return an error when using the Struct method when passed by reference", func() {
				err := validation.Struct(&test)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("validation failed on field 'Value' with validator 'gte' and parameter(s) '0'"))
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

		It("should succeed when using the Struct method when passed by value", func() {
			Expect(validation.Struct(test)).To(Succeed())
		})

		It("should succeed when using the Struct method when passed by reference", func() {
			Expect(validation.Struct(&test)).To(Succeed())
		})
	})

	When("an argument is passed to the Struct function that is not a struct", func() {
		var (
			test int
		)

		BeforeEach(func() {
			test = 0
		})

		It("should return an error when using the Struct method when passed by value", func() {
			Expect(validation.Struct(test)).To(HaveOccurred())
		})

		It("should return an error when using the Struct method when passed by reference", func() {
			Expect(validation.Struct(&test)).To(HaveOccurred())
		})
	})

	When("nil passed to the Struct function", func() {
		It("should return an error", func() {
			Expect(validation.Struct(nil)).To(HaveOccurred())
		})
	})

	When("a custom validator for a tag called 'test' is registered and always fails", Ordered, func() {
		const (
			tag    = "test"
			errMsg = "test validation error msg"
		)

		BeforeAll(func() {
			validation.RegisterValidation(tag, func(field validator.FieldLevel) bool {
				return false
			}, func(err validator.FieldError) string {
				return errMsg
			})
		})

		When("a struct uses the test tag", func() {
			type testStruct struct {
				Name string `validate:"test"`
			}

			var (
				test testStruct
			)

			BeforeEach(func() {
				test = testStruct{
					Name: "testStructName",
				}
			})

			It("should return an error when using the Struct method", func() {
				err := validation.Struct(test)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(errMsg))
			})
		})

		When("a struct doesn't use the test tag", func() {
			type testStruct struct {
				Value int `validate:"gte=0"`
			}

			var (
				test testStruct
			)

			BeforeEach(func() {
				test = testStruct{
					Value: 0,
				}
			})

			It("should validate without errors", func() {
				Expect(validation.Struct(test)).To(Succeed())
			})
		})
	})

	When("a struct has multiple fields with validation rules", func() {
		type testStruct struct {
			IntValue int    `validate:"gte=0"`
			StrValue string `validate:"required"`
		}

		var (
			test testStruct
		)

		BeforeEach(func() {
			test = testStruct{}
		})

		When("all the struct values fail validation", func() {
			BeforeEach(func() {
				test.IntValue = -1
				test.StrValue = ""
			})

			It("should return an error that has a message for both fields when using the Struct method", func() {
				err := validation.Struct(test)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'IntValue' with validator 'gte' and parameter(s) '0'"))
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'StrValue' with validator 'required'"))
			})
		})
	})

	When("validation is done on a standalone variable", func() {
		It("should return an error when the variable is nil and has a required tag", func() {
			err := validation.Var(nil, "required")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("validation failed with validator 'required'"))
		})

		It("should succeed when the variable is nil and has no tag", func() {
			Expect(validation.Var(nil, "")).To(Succeed())
		})
	})

	When("validation is done on a standalone integer pointer variable", func() {
		var (
			testVar *int
		)

		BeforeEach(func() {
			testVar = new(int)
			*testVar = 1
		})

		It("should succeed if there are no validation violations", func() {
			Expect(validation.Var(testVar, "required,gt=0")).To(Succeed())
		})

		It("should fail if there is a validation violation", func() {
			err := validation.Var(testVar, "required,lt=0")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("validation failed with validator 'lt' and parameter(s) '0'"))
		})

		It("should succeed when there are no validators", func() {
			Expect(validation.Var(testVar, "")).To(Succeed())
		})
	})

	When("the same custom validation is registered twice", func() {
		It("should panic", func() {
			Expect(func() {
				for i := 0; i < 2; i++ {
					validation.RegisterValidation("custom_twice", func(fl validator.FieldLevel) bool { return true }, func(err validator.FieldError) string { return "" })
				}
			}).To(PanicWith(ContainSubstring("'custom_twice' already has a registered validation function")))
		})
	})

	When("registering a validation tag and the validation func is nil", func() {
		It("should panic", func() {
			Expect(func() {
				validation.RegisterValidation("nil_func", nil, func(err validator.FieldError) string { return "" })
			}).To(PanicWith(ContainSubstring("Failed to register the validation function for the tag 'nil_func'")))
		})
	})

	When("registering a validation tag and the error msg func is nil", func() {
		It("should panic", func() {
			Expect(func() {
				validation.RegisterValidation("nil_error_func", func(fl validator.FieldLevel) bool { return true }, nil)
			}).To(PanicWith(ContainSubstring("Tag 'nil_error_func' has a nil error message function")))
		})
	})
})
