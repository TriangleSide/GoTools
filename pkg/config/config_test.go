package config_test

import (
	"intelligence/pkg/config"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	When("a struct has a field called Value with a default of 0, a validation rule of gte 0, and is required", func() {
		const (
			EnvName      = "VALUE"
			DefaultValue = 0
		)

		type testStruct struct {
			Value int `default:"0" validate:"gte=0" required:"true"`
		}

		When("an environment variable called VALUE is set with a value of 1", func() {
			const (
				EnvValueStr = "1"
			)

			BeforeEach(func() {
				Expect(os.Setenv(EnvName, EnvValueStr)).To(Succeed())
			})

			AfterEach(func() {
				Expect(os.Unsetenv(EnvName)).To(Succeed())
			})

			It("should be set in the Value field of the struct", func() {
				conf, err := config.ProcessConfiguration[testStruct]()
				Expect(err).To(Not(HaveOccurred()))
				Expect(conf).To(Not(BeNil()))
				intValue, err := strconv.Atoi(EnvValueStr)
				Expect(err).To(Not(HaveOccurred()))
				Expect(conf.Value).To(Equal(intValue))
			})
		})

		When("an environment variable called VALUE is set with a value of -1", func() {
			const (
				EnvValueStr = "-1"
			)

			BeforeEach(func() {
				Expect(os.Setenv(EnvName, EnvValueStr)).To(Succeed())
			})

			AfterEach(func() {
				Expect(os.Unsetenv(EnvName)).To(Succeed())
			})

			It("should return an error when processing the configuration", func() {
				conf, err := config.ProcessConfiguration[testStruct]()
				Expect(err).To(HaveOccurred())
				Expect(conf).To(BeNil())
			})
		})

		When("no environment variable is set for the struct field", func() {
			It("should set the value to the default", func() {
				conf, err := config.ProcessConfiguration[testStruct]()
				Expect(err).To(Not(HaveOccurred()))
				Expect(conf).To(Not(BeNil()))
				Expect(conf.Value).To(Equal(DefaultValue))
			})
		})
	})

	When("a struct has a field called Value with no default, validation, or required tag", func() {
		type testStruct struct {
			Value *int
		}

		It("should set the field to nil when processing the configuration", func() {
			conf, err := config.ProcessConfiguration[testStruct]()
			Expect(err).To(Not(HaveOccurred()))
			Expect(conf).To(Not(BeNil()))
			Expect(conf.Value).To(BeNil())
		})
	})

	When("a struct has a field called Value with no default or validation, but it has a required tag", func() {
		type testStruct struct {
			Value *int `required:"true"`
		}

		It("should return an error when processing the configuration", func() {
			conf, err := config.ProcessConfiguration[testStruct]()
			Expect(err).To(HaveOccurred())
			Expect(conf).To(BeNil())
		})
	})
})
