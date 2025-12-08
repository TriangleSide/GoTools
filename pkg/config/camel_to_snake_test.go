package config_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestCamelToSnake(t *testing.T) {
	t.Run("when field is MyCamelCase it should map to MY_CAMEL_CASE", func(t *testing.T) {
		type testStruct struct {
			MyCamelCase string `config:"ENV"`
		}
		t.Setenv("MY_CAMEL_CASE", "test_value")
		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.Equals(t, conf.MyCamelCase, "test_value")
	})

	t.Run("when field has consecutive uppercase letters it should split correctly", func(t *testing.T) {
		type testStruct struct {
			CAMELCase string `config:"ENV"`
		}
		t.Setenv("CAMEL_CASE", "consecutive_upper")
		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.Equals(t, conf.CAMELCase, "consecutive_upper")
	})

	t.Run("when field has numbers followed by letters it should map correctly", func(t *testing.T) {
		type testStruct struct {
			Field1aSplit string `config:"ENV"`
		}
		t.Setenv("FIELD1A_SPLIT", "number_value")
		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.Equals(t, conf.Field1aSplit, "number_value")
	})

	t.Run("when field has multiple consecutive numbers it should handle them correctly", func(t *testing.T) {
		type testStruct struct {
			Field1a1Split string `config:"ENV"`
		}
		t.Setenv("FIELD1A1_SPLIT", "multi_number")
		conf, err := config.ProcessAndValidate[testStruct]()
		assert.NoError(t, err)
		assert.NotNil(t, conf)
		assert.Equals(t, conf.Field1a1Split, "multi_number")
	})
}
