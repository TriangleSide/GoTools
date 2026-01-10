package config_test

import (
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/config"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
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

func TestCamelToSnake_SingleUppercaseCharacter_NoUnderscore(t *testing.T) {
	type testStruct struct {
		A string `config:"ENV"`
	}
	t.Setenv("A", "single_upper")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.A, "single_upper")
}

func TestCamelToSnake_TrailingUppercaseAfterLowercase_AddsUnderscore(t *testing.T) {
	type testStruct struct {
		TestAB string `config:"ENV"`
	}
	t.Setenv("TEST_AB", "trailing_upper")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.TestAB, "trailing_upper")
}

func TestCamelToSnake_AllUppercase_NoUnderscores(t *testing.T) {
	type testStruct struct {
		ABC string `config:"ENV"`
	}
	t.Setenv("ABC", "all_upper")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.ABC, "all_upper")
}

func TestCamelToSnake_LowercaseStart_ConvertsToUppercase(t *testing.T) {
	type testStruct struct {
		CamelCase string `config:"ENV"`
	}
	t.Setenv("CAMEL_CASE", "lowercase_start")
	conf, err := config.ProcessAndValidate[testStruct]()
	assert.NoError(t, err)
	assert.NotNil(t, conf)
	assert.Equals(t, conf.CamelCase, "lowercase_start")
}
