package config_test

import (
	"errors"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestConfigProcessor(t *testing.T) {
	t.Run("when the default value cannot be assigned to the struct field", func(t *testing.T) {
		type testStruct struct {
			Value *int `config:"ENV" config_default:"NOT_AN_INT"`
		}
		conf, err := config.ProcessAndValidate[testStruct]()
		assert.ErrorPart(t, err, "failed to assign default value NOT_AN_INT to field Value")
		assert.Nil(t, conf)
	})

	t.Run("when a struct has an int field called Value with a default of 1 and a validation rule of gte=0 and is required", func(t *testing.T) {
		const (
			EnvName      = "VALUE"
			DefaultValue = 1
		)

		type testStruct struct {
			Value int `config:"ENV" config_default:"1" validate:"required,gte=0"`
		}

		t.Run("when the environment variable VALUE is set to NOT_AN_INT", func(t *testing.T) {
			t.Setenv(EnvName, "NOT_AN_INT")
			conf, err := config.ProcessAndValidate[testStruct]()
			assert.ErrorPart(t, err, "failed to assign value NOT_AN_INT to field Value")
			assert.Nil(t, conf)
		})

		t.Run("when the environment variable VALUE is set to 2", func(t *testing.T) {
			t.Setenv(EnvName, "2")

			t.Run("it should be set in the Value field of the struct", func(t *testing.T) {
				conf, err := config.ProcessAndValidate[testStruct]()
				assert.NoError(t, err)
				assert.NotNil(t, conf)
				assert.Equals(t, conf.Value, 2)
			})
		})

		t.Run("when the validation rule fails it should fail to process", func(t *testing.T) {
			t.Setenv(EnvName, "-1")
			conf, err := config.ProcessAndValidate[testStruct]()
			assert.ErrorPart(t, err, "validation failed")
			assert.Nil(t, conf)
		})

		t.Run("when no environment variable is set it should use the default value", func(t *testing.T) {
			conf, err := config.ProcessAndValidate[testStruct]()
			assert.NoError(t, err)
			assert.NotNil(t, conf)
			assert.Equals(t, conf.Value, DefaultValue)
		})

		t.Run("when there is no environment variable and no default it should fail", func(t *testing.T) {
			type noDefault struct {
				Value string `config:"ENV"`
			}
			conf, err := config.ProcessAndValidate[noDefault]()
			assert.ErrorPart(t, err, "no value found for field Value")
			assert.Nil(t, conf)
		})
	})

	t.Run("when a struct has a field called Value with no default or validation or required tag it should return a struct with unmodified fields", func(t *testing.T) {
		type testStruct struct {
			Value *int
		}
		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.Nil(t, conf.Value)
	})

	t.Run("when a struct has a field called Value with no config tags but it has a required validation it should return and error", func(t *testing.T) {
		type testStruct struct {
			Value *int `validate:"required"`
		}
		conf, err := config.ProcessAndValidate[testStruct]()
		assert.ErrorPart(t, err, "validation failed")
		assert.Nil(t, conf)
	})

	t.Run("when a struct has a field and has an embedded anonymous struct with a field it should be able to set both fields", func(t *testing.T) {
		type embeddedStruct struct {
			EmbeddedField string `config:"ENV" validate:"required"`
		}

		type testStruct struct {
			embeddedStruct
			Field string `config:"ENV" validate:"required"`
		}

		const (
			EmbeddedEnvName = "EMBEDDED_FIELD"
			EmbeddedValue   = "embeddedField"
			FieldEnvName    = "FIELD"
			FieldValue      = "field"
		)

		t.Setenv(EmbeddedEnvName, EmbeddedValue)
		t.Setenv(FieldEnvName, FieldValue)

		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.Equals(t, conf.EmbeddedField, EmbeddedValue)
		assert.Equals(t, conf.Field, FieldValue)
	})

	t.Run("when a custom processor is registered it should be used", func(t *testing.T) {
		type testStruct struct {
			Value string `config:"CUSTOM"`
		}

		var called bool
		config.MustRegisterProcessor("CUSTOM", func(fieldName string, _ *structs.FieldMetadata) (string, bool, error) {
			called = true
			return "custom", true, nil
		})

		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.True(t, called)
		assert.Equals(t, conf.Value, "custom")
	})

	t.Run("multiple processors should be able to be used", func(t *testing.T) {
		type testStruct struct {
			EnvValue   string `config:"ENV"`
			OtherValue string `config:"OTHER"`
		}

		var called bool
		config.MustRegisterProcessor("OTHER", func(fieldName string, _ *structs.FieldMetadata) (string, bool, error) {
			called = true
			return "OtherValue", true, nil
		})

		t.Setenv("ENV_VALUE", "EnvValue")
		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.True(t, called)
		assert.Equals(t, conf.OtherValue, "OtherValue")
		assert.Equals(t, conf.EnvValue, "EnvValue")
	})

	t.Run("when a custom processor returns not found and a default value is provided", func(t *testing.T) {
		const procName = "NOT_FOUND_DEFAULT"

		config.MustRegisterProcessor(procName, func(_ string, _ *structs.FieldMetadata) (string, bool, error) {
			return "", false, nil
		})

		type testStruct struct {
			Value string `config:"NOT_FOUND_DEFAULT" config_default:"default"`
		}

		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.Equals(t, conf.Value, "default")
	})

	t.Run("when a custom processor returns not found and no default value is provided", func(t *testing.T) {
		const procName = "NOT_FOUND_NO_DEFAULT"

		config.MustRegisterProcessor(procName, func(_ string, _ *structs.FieldMetadata) (string, bool, error) {
			return "", false, nil
		})

		type testStruct struct {
			Value string `config:"NOT_FOUND_NO_DEFAULT"`
		}

		conf, err := config.ProcessAndValidate[testStruct]()
		assert.ErrorPart(t, err, "no value found for field Value")
		assert.Nil(t, conf)
	})

	t.Run("when a SourceFunc returns an error it should fail", func(t *testing.T) {
		const procName = "ERROR_PROCESSOR"

		customErr := errors.New("custom processor failed")

		config.MustRegisterProcessor(procName, func(_ string, _ *structs.FieldMetadata) (string, bool, error) {
			return "", false, customErr
		})

		type testStruct struct {
			Value string `config:"ERROR_PROCESSOR"`
		}

		conf, err := config.ProcessAndValidate[testStruct]()
		assert.ErrorPart(t, err, customErr.Error())
		assert.Nil(t, conf)
	})

	t.Run("when the processor does not exist it should return an error", func(t *testing.T) {
		type testStruct struct {
			Value string `config:"DOES_NOT_EXIST"`
		}

		conf, err := config.ProcessAndValidate[testStruct]()
		assert.ErrorPart(t, err, "processor DOES_NOT_EXIST not registered")
		assert.Nil(t, conf)
	})
}
