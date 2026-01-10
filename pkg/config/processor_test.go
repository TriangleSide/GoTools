package config_test

import (
	"errors"
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/config"
	"github.com/TriangleSide/go-toolkit/pkg/structs"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
	"github.com/TriangleSide/go-toolkit/pkg/test/once"
)

func TestProcessAndValidate_DefaultValueCannotBeAssigned_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value *int `config:"ENV" config_default:"NOT_AN_INT"`
	}
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.ErrorPart(t, err, "failed to assign value NOT_AN_INT to field Value")
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

	once.Do(t, func() {
		config.MustRegisterProcessor("RETURNS_VALUE_CUSTOM_PROCESSOR",
			func(string, *structs.FieldMetadata) (string, bool, error) {
				return "value", true, nil
			})
	})

	type testStruct struct {
		Value string `config:"RETURNS_VALUE_CUSTOM_PROCESSOR"`
	}

	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.Value, "value")
}

func TestProcessAndValidate_MultipleProcessors_UsesAllProcessors(t *testing.T) {
	once.Do(t, func() {
		config.MustRegisterProcessor("RETURNS_VALUE_MULTIPLE_PROCESSORS",
			func(string, *structs.FieldMetadata) (string, bool, error) {
				return "value", true, nil
			})
	})

	type testStruct struct {
		EnvValue   string `config:"ENV"`
		OtherValue string `config:"RETURNS_VALUE_MULTIPLE_PROCESSORS"`
	}

	t.Setenv("ENV_VALUE", "EnvValue")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.OtherValue, "value")
	assert.Equals(t, conf.EnvValue, "EnvValue")
}

func TestProcessAndValidate_CustomProcessorNotFound_WithDefault_UsesDefault(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		config.MustRegisterProcessor("NOT_FOUND_WITH_DEFAULT",
			func(_ string, _ *structs.FieldMetadata) (string, bool, error) {
				return "", false, nil
			})
	})

	type testStruct struct {
		Value string `config:"NOT_FOUND_WITH_DEFAULT" config_default:"default"`
	}

	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.Value, "default")
}

func TestProcessAndValidate_CustomProcessorNotFound_NoDefault_ReturnsError(t *testing.T) {
	t.Parallel()

	once.Do(t, func() {
		config.MustRegisterProcessor("NOT_FOUND_NO_DEFAULT",
			func(_ string, _ *structs.FieldMetadata) (string, bool, error) {
				return "", false, nil
			})
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

	errCustomProcessorLocal := errors.New("custom processor failed")
	once.Do(t, func() {
		config.MustRegisterProcessor("ERROR_PROCESSOR_SOURCE_FUNC",
			func(_ string, _ *structs.FieldMetadata) (string, bool, error) {
				return "", false, errCustomProcessorLocal
			})
	})

	type testStruct struct {
		Value string `config:"ERROR_PROCESSOR_SOURCE_FUNC"`
	}

	conf, err := config.ProcessAndValidate[testStruct]()
	assert.ErrorPart(t, err, "failed to process configuration")
	assert.ErrorPart(t, err, "error while fetching the value for field Value using processor ERROR_PROCESSOR_SOURCE_FUNC")
	assert.ErrorPart(t, err, errCustomProcessorLocal.Error())
	assert.Nil(t, conf)
}

func TestProcessAndValidate_ProcessorNotRegistered_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value string `config:"DOES_NOT_EXIST"`
	}

	conf, err := config.ProcessAndValidate[testStruct]()
	assert.ErrorPart(t, err, "failed to process configuration")
	assert.ErrorPart(t, err, "processor \"DOES_NOT_EXIST\" not registered")
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
