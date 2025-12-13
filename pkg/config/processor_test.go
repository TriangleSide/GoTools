package config_test

import (
	"errors"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestProcessAndValidate_DefaultValueCannotBeAssigned_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value *int `config:"ENV" config_default:"NOT_AN_INT"`
	}
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.ErrorPart(t, err, "failed to assign default value NOT_AN_INT to field Value")
	assert.Nil(t, conf)
}

func TestProcessAndValidate_EnvSetToInvalidType_ReturnsError(t *testing.T) {
	const EnvName = "VALUE"

	type testStruct struct {
		Value int `config:"ENV" config_default:"1" validate:"required,gte=0"`
	}

	t.Setenv(EnvName, "NOT_AN_INT")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.ErrorPart(t, err, "failed to assign value NOT_AN_INT to field Value")
	assert.Nil(t, conf)
}

func TestProcessAndValidate_EnvSetToValidValue_SetsField(t *testing.T) {
	const EnvName = "VALUE"

	type testStruct struct {
		Value int `config:"ENV" config_default:"1" validate:"required,gte=0"`
	}

	t.Setenv(EnvName, "2")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.Value, 2)
}

func TestProcessAndValidate_ValidationRuleFails_ReturnsError(t *testing.T) {
	const EnvName = "VALUE"

	type testStruct struct {
		Value int `config:"ENV" config_default:"1" validate:"required,gte=0"`
	}

	t.Setenv(EnvName, "-1")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.ErrorPart(t, err, "validation failed")
	assert.Nil(t, conf)
}

func TestProcessAndValidate_NoEnvSet_UsesDefaultValue(t *testing.T) {
	t.Parallel()
	const DefaultValue = 1

	type testStruct struct {
		Value int `config:"ENV" config_default:"1" validate:"required,gte=0"`
	}

	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.Value, DefaultValue)
}

func TestProcessAndValidate_NoEnvAndNoDefault_ReturnsError(t *testing.T) {
	t.Parallel()
	type noDefault struct {
		Value string `config:"ENV"`
	}
	conf, err := config.ProcessAndValidate[noDefault]()
	assert.ErrorPart(t, err, "no value found for field Value")
	assert.Nil(t, conf)
}

func TestProcessAndValidate_NoConfigTags_ReturnsUnmodifiedFields(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value *int
	}
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Nil(t, conf.Value)
}

func TestProcessAndValidate_NoConfigTagsWithRequiredValidation_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value *int `validate:"required"`
	}
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.ErrorPart(t, err, "validation failed")
	assert.Nil(t, conf)
}

func TestProcessAndValidate_EmbeddedAnonymousStruct_SetsBothFields(t *testing.T) {
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
}

func TestProcessAndValidate_CustomProcessor_UsesCustomProcessor(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string `config:"CUSTOM"`
	}

	var called bool
	config.MustRegisterProcessor("CUSTOM", func(string, *structs.FieldMetadata) (string, bool, error) {
		called = true
		return "custom", true, nil
	})

	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.True(t, called)
	assert.Equals(t, conf.Value, "custom")
}

func TestProcessAndValidate_MultipleProcessors_UsesAllProcessors(t *testing.T) {
	type testStruct struct {
		EnvValue   string `config:"ENV"`
		OtherValue string `config:"OTHER"`
	}

	var called bool
	config.MustRegisterProcessor("OTHER", func(string, *structs.FieldMetadata) (string, bool, error) {
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
}

func TestProcessAndValidate_CustomProcessorNotFound_WithDefault_UsesDefault(t *testing.T) {
	t.Parallel()
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
}

func TestProcessAndValidate_CustomProcessorNotFound_NoDefault_ReturnsError(t *testing.T) {
	t.Parallel()
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
}

func TestProcessAndValidate_SourceFuncReturnsError_ReturnsError(t *testing.T) {
	t.Parallel()
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
}

func TestProcessAndValidate_ProcessorNotRegistered_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string `config:"DOES_NOT_EXIST"`
	}

	conf, err := config.ProcessAndValidate[testStruct]()
	assert.ErrorPart(t, err, "processor DOES_NOT_EXIST not registered")
	assert.Nil(t, conf)
}

func TestProcess_EnvSetToValidValue_SetsFieldWithoutValidation(t *testing.T) {
	type testStruct struct {
		Value int `config:"ENV" validate:"gte=100"`
	}

	t.Setenv("VALUE", "5")
	conf, err := config.Process[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.Value, 5)
}
