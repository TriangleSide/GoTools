package config_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestCamelToSnake_StandardCamelCase_MapsToSnakeCase(t *testing.T) {
	type testStruct struct {
		MyCamelCase string `config:"ENV"`
	}
	t.Setenv("MY_CAMEL_CASE", "test_value")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.MyCamelCase, "test_value")
}

func TestCamelToSnake_ConsecutiveUppercase_SplitsCorrectly(t *testing.T) {
	type testStruct struct {
		CAMELCase string `config:"ENV"`
	}
	t.Setenv("CAMEL_CASE", "consecutive_upper")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.CAMELCase, "consecutive_upper")
}

func TestCamelToSnake_NumbersFollowedByLetters_MapsCorrectly(t *testing.T) {
	type testStruct struct {
		Field1aSplit string `config:"ENV"`
	}
	t.Setenv("FIELD1A_SPLIT", "number_value")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.Field1aSplit, "number_value")
}

func TestCamelToSnake_MultipleConsecutiveNumbers_HandlesCorrectly(t *testing.T) {
	type testStruct struct {
		Field1a1Split string `config:"ENV"`
	}
	t.Setenv("FIELD1A1_SPLIT", "multi_number")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.Field1a1Split, "multi_number")
}
