package envprocessor_test

import (
	"os"
	"strings"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/config/envprocessor"
)

func TestEnvProcessor(t *testing.T) {
	setEnv := func(t *testing.T, envName string, value string) {
		err := os.Setenv(envName, value)
		if err != nil {
			t.Fatalf("failed to set environment variable %s to %s", envName, value)
		}
	}

	unsetEnv := func(t *testing.T, envName string) {
		err := os.Unsetenv(envName)
		if err != nil {
			t.Fatalf("failed to unset environment variable %s", envName)
		}
	}

	t.Run("when config_format is an invalid value", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic but got none")
			}
		}()
		type testStruct struct {
			Value int `config_format:"not_valid"`
		}
		_, _ = envprocessor.ProcessAndValidate[testStruct]()
	})

	t.Run("when the default value cannot be assigned to the struct field", func(t *testing.T) {
		type testStruct struct {
			Value *int `config_format:"snake" config_default:"NOT_AN_INT"`
		}
		conf, err := envprocessor.ProcessAndValidate[testStruct]()
		if err == nil {
			t.Fatalf("should not be able to process an invalid value")
		}
		if conf != nil {
			t.Fatalf("expected nil config but got %v", conf)
		}
		if !strings.Contains(err.Error(), "failed to assign default value NOT_AN_INT to field Value") {
			t.Fatalf("error message is not correct (%s)", err.Error())
		}
	})

	t.Run("when a struct has an int field called Value with a default of 1, a validation rule of gte=0, and is required", func(t *testing.T) {
		const (
			EnvName      = "VALUE"
			DefaultValue = 1
		)

		type testStruct struct {
			Value int `config_format:"snake" config_default:"1" validate:"required,gte=0"`
		}

		t.Run("when the environment variable VALUE is set to NOT_AN_INT", func(t *testing.T) {
			t.Cleanup(func() {
				unsetEnv(t, EnvName)
			})
			setEnv(t, EnvName, "NOT_AN_INT")
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			if err == nil {
				t.Fatalf("should not be able to process an invalid value")
			}
			if conf != nil {
				t.Fatalf("expected nil config but got %v", conf)
			}
			if !strings.Contains(err.Error(), "failed to assign env var NOT_AN_INT to field Value") {
				t.Fatalf("error message is not correct (%s)", err.Error())
			}
		})

		t.Run("when the environment variable VALUE is set to 2", func(t *testing.T) {
			t.Cleanup(func() {
				unsetEnv(t, EnvName)
			})
			setEnv(t, EnvName, "2")

			t.Run("it should be set in the Value field of the struct", func(t *testing.T) {
				conf, err := envprocessor.ProcessAndValidate[testStruct]()
				if err != nil {
					t.Fatalf("should be able to process the environment values")
				}
				if conf == nil {
					t.Fatalf("expected a configuration but got nil")
				}
				if conf.Value != 2 {
					t.Fatalf("expected value to be 2 but got %d", conf.Value)
				}
			})

			t.Run("it should use the default if a prefix is used", func(t *testing.T) {
				conf, err := envprocessor.ProcessAndValidate[testStruct](envprocessor.WithPrefix("PREFIX"))
				if err != nil {
					t.Fatalf("should be able to process the environment values")
				}
				if conf == nil {
					t.Fatalf("expected a configuration but got nil")
				}
				if conf.Value != DefaultValue {
					t.Fatalf("expected value to be the default but got %d", conf.Value)
				}
			})

			t.Run("when an environment variable called TEST_VALUE is set with a value of 3", func(t *testing.T) {
				const (
					EnvNameWithPrefix = "TEST_VALUE"
				)
				t.Cleanup(func() {
					unsetEnv(t, EnvNameWithPrefix)
				})
				setEnv(t, EnvNameWithPrefix, "3")

				t.Run("it should be able to be set in the struct if parse with a TEST prefix", func(t *testing.T) {
					conf, err := envprocessor.ProcessAndValidate[testStruct](envprocessor.WithPrefix("TEST"))
					if err != nil {
						t.Fatalf("should be able to process the environment values")
					}
					if conf == nil {
						t.Fatalf("expected a configuration but got nil")
					}
					if conf.Value != 3 {
						t.Fatalf("expected value to be 3 but got %d", conf.Value)
					}
				})
			})
		})

		t.Run("when the validation rule fails", func(t *testing.T) {
			t.Cleanup(func() {
				unsetEnv(t, EnvName)
			})
			setEnv(t, EnvName, "-1")

			t.Run("it should fail to process", func(t *testing.T) {
				conf, err := envprocessor.ProcessAndValidate[testStruct]()
				if err == nil {
					t.Fatalf("should not be able to process the environment values")
				}
				if conf != nil {
					t.Fatalf("expected nil config but got %v", conf)
				}
				if !strings.Contains(err.Error(), "validation failed") {
					t.Fatalf("error message is not correct (%s)", err.Error())
				}
			})
		})

		t.Run("when no environment variable is set it should use the default value", func(t *testing.T) {
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			if err != nil {
				t.Fatalf("should be able to process the environment values")
			}
			if conf == nil {
				t.Fatalf("expected a configuration but got nil")
			}
			if conf.Value != DefaultValue {
				t.Fatalf("expected value to be the default but got %d", conf.Value)
			}
		})
	})

	t.Run("when a struct has a field called Value with no default, validation, or required tag", func(t *testing.T) {
		type testStruct struct {
			Value *int
		}

		t.Run("it should return a struct with unmodified fields", func(t *testing.T) {
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			if err != nil {
				t.Fatalf("should be able to process the environment values")
			}
			if conf == nil {
				t.Fatalf("expected a configuration but got nil")
			}
			if conf.Value != nil {
				t.Fatalf("expected value to be nil but got %d", *conf.Value)
			}
		})
	})

	t.Run("when a struct has a field called Value with no config tags, but it has a required validation", func(t *testing.T) {
		type testStruct struct {
			Value *int `validate:"required"`
		}

		t.Run("it should return a validation error when processing the configuration", func(t *testing.T) {
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			if err == nil {
				t.Fatalf("should not be able to process the environment values")
			}
			if conf != nil {
				t.Fatalf("expected nil config but got %v", conf)
			}
			if !strings.Contains(err.Error(), "validation failed") {
				t.Fatalf("validation failed on field 'Value' with validator 'required'")
			}
		})
	})

	t.Run("when a struct has a field and has an embedded anonymous struct with a field", func(t *testing.T) {
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

		t.Cleanup(func() {
			unsetEnv(t, EmbeddedEnvName)
			unsetEnv(t, FieldEnvName)
		})
		setEnv(t, EmbeddedEnvName, EmbeddedValue)
		setEnv(t, FieldEnvName, FieldValue)

		t.Run("it should be able to set both fields", func(t *testing.T) {
			conf, err := envprocessor.ProcessAndValidate[testStruct]()
			if err != nil {
				t.Fatalf("should be able to process the environment values")
			}
			if conf == nil {
				t.Fatalf("expected a configuration but got nil")
			}
			if conf.EmbeddedField != EmbeddedValue {
				t.Fatalf("expected value to be %s field but got %s", EmbeddedValue, conf.EmbeddedField)
			}
			if conf.Field != FieldValue {
				t.Fatalf("expected value to be %s field but got %s", FieldValue, conf.Field)
			}
		})
	})
}
