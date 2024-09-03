package envprocessor_test

import (
	"errors"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/config/envprocessor"
)

var _ = Describe("config", func() {
	When("a struct has a field that has an invalid format", func() {
		It("should panic", func() {
			type testStruct struct {
				Value int `config_format:"not_valid"`
			}
			Expect(func() {
				_, _ = envprocessor.ProcessAndValidate[testStruct]()
			}).To(PanicWith(ContainSubstring("invalid config format (not_valid)")))
		})
	})

	When("a struct has a field called Value with an invalid default", func() {
		It("should return a struct with unmodified fields", func() {
			type testStruct struct {
				Value *int `config_format:"snake" config_default:"NOT_AN_INT"`
			}
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			Expect(err).To(HaveOccurred())
			Expect(conf).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to assign default value NOT_AN_INT to field Value"))
		})
	})

	When("a struct has an int field called Value with a default of 1, a validation rule of gte=0, and is required", func() {
		const (
			EnvName      = "VALUE"
			DefaultValue = 1
		)

		type testStruct struct {
			Value int `config_format:"snake" config_default:"1" validate:"required,gte=0"`
		}

		AfterEach(func() {
			Expect(os.Unsetenv(EnvName)).To(Succeed())
		})

		When("an environment variable called VALUE is set with a value of 'NOT_AN_INT'", func() {
			It("should return an error", func() {
				Expect(os.Setenv(EnvName, "NOT_AN_INT")).To(Succeed())
				conf, err := envprocessor.ProcessAndValidate[testStruct]()
				Expect(err).To(HaveOccurred())
				Expect(conf).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to assign env var NOT_AN_INT to field Value"))
			})
		})

		When("an environment variable called VALUE is set with a value of 2", func() {
			const (
				EnvValueStr = "2"
			)

			BeforeEach(func() {
				Expect(os.Setenv(EnvName, EnvValueStr)).To(Succeed())
			})

			It("should be set in the Value field of the struct", func() {
				conf, err := envprocessor.ProcessAndValidate[testStruct]()
				Expect(err).To(Not(HaveOccurred()))
				Expect(conf).To(Not(BeNil()))
				intValue, err := strconv.Atoi(EnvValueStr)
				Expect(err).To(Not(HaveOccurred()))
				Expect(conf.Value).To(Equal(intValue))
			})

			It("should use the default value if a prefix of NOT_EXIST is used", func() {
				conf, err := envprocessor.ProcessAndValidate[testStruct](envprocessor.WithPrefix("NOT_EXIST"))
				Expect(err).To(Not(HaveOccurred()))
				Expect(conf).To(Not(BeNil()))
				Expect(conf.Value).To(Equal(DefaultValue))
			})

			When("an environment variable called TEST_VALUE is set with a value of 2", func() {
				const (
					Prefix            = "TEST"
					PrefixEnvName     = Prefix + "_" + EnvName
					PrefixEnvValueStr = "2"
				)

				BeforeEach(func() {
					Expect(os.Setenv(PrefixEnvName, PrefixEnvValueStr)).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.Unsetenv(PrefixEnvName)).To(Succeed())
				})

				It("should be set in the Value field of the struct if the TEST prefix is used with the processor", func() {
					conf, err := envprocessor.ProcessAndValidate[testStruct](envprocessor.WithPrefix(Prefix))
					Expect(err).To(Not(HaveOccurred()))
					Expect(conf).To(Not(BeNil()))
					intValue, err := strconv.Atoi(PrefixEnvValueStr)
					Expect(err).To(Not(HaveOccurred()))
					Expect(conf.Value).To(Equal(intValue))
				})
			})
		})

		When("an environment variable called VALUE is set with a value of -1", func() {
			const (
				EnvValueStr = "-1"
			)

			BeforeEach(func() {
				Expect(os.Setenv(EnvName, EnvValueStr)).To(Succeed())
			})

			It("should return a validation error when processing the configuration", func() {
				conf, err := envprocessor.ProcessAndValidate[testStruct]()
				Expect(err).To(HaveOccurred())
				Expect(conf).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed"))
			})
		})

		When("no environment variable is set", func() {
			It("should set the value to the default", func() {
				conf, err := envprocessor.ProcessAndValidate[testStruct]()
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

		It("should return a struct with unmodified fields", func() {
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			Expect(err).To(Not(HaveOccurred()))
			Expect(conf).To(Not(BeNil()))
			Expect(conf.Value).To(BeNil())
		})
	})

	When("a struct has a field called Value with no config tags, but it has a required validation", func() {
		type testStruct struct {
			Value *int `validate:"required"`
		}

		It("should return a validation error when processing the configuration", func() {
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			Expect(err).To(HaveOccurred())
			Expect(conf).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("validation failed on field 'Value' with validator 'required'"))
		})
	})

	When("a struct a field and has an embedded anonymous struct with a field", func() {
		type embeddedStruct struct {
			EmbeddedField string `config_format:"snake" validate:"required"`
		}

		type testStruct struct {
			embeddedStruct
			Field string `config_format:"snake" validate:"required"`
		}

		const (
			EmbeddedEnvName = "EMBEDDED_FIELD"
			EmbeddedValue   = "embeddedField"
			FieldEnvName    = "FIELD"
			FieldValue      = "field"
		)

		BeforeEach(func() {
			Expect(os.Setenv(EmbeddedEnvName, EmbeddedValue)).To(Succeed())
			Expect(os.Setenv(FieldEnvName, FieldValue)).To(Succeed())
		})

		AfterEach(func() {
			Expect(os.Unsetenv(EmbeddedEnvName)).To(Succeed())
			Expect(os.Unsetenv(FieldEnvName)).To(Succeed())
		})

		It("should be able to set both fields", func() {
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			Expect(err).ToNot(HaveOccurred())
			Expect(conf.Field).To(Equal(FieldValue))
			Expect(conf.EmbeddedField).To(Equal(EmbeddedValue))
		})
	})

	When("the options fails", func() {
		It("should return a validation error when processing the configuration", func() {
			conf, err := envprocessor.ProcessAndValidate[struct {
				Field string
			}](func(cfg *envprocessor.Config) error {
				return errors.New("error")
			})
			Expect(err).To(HaveOccurred())
			Expect(conf).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to set the options for the configuration processor (error)"))
		})
	})
})