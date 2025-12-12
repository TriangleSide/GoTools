package api_test

import (
	"testing"

	_ "github.com/TriangleSide/GoTools/pkg/http/api"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestPathValidation_VariousPaths_ValidatesCorrectly(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name             string
		path             string
		expectedErrorMsg string
	}

	testCases := []testCase{
		{name: "Root_Succeeds", path: "/", expectedErrorMsg: ""},
		{name: "MultiPart_Succeeds", path: "/a/b/c/1/2/3", expectedErrorMsg: ""},
		{name: "WithParameterPart_Succeeds", path: "/a/{b}/c", expectedErrorMsg: ""},
		{name: "Empty_Fails", path: "", expectedErrorMsg: "path cannot be empty"},
		{name: "InvalidCharactersPlus_Fails", path: "/+", expectedErrorMsg: "path contains invalid characters"},
		{name: "InvalidCharactersLeadingSpace_Fails", path: " /a", expectedErrorMsg: "path contains invalid characters"},
		{name: "InvalidCharactersTrailingSpace_Fails", path: "/a ", expectedErrorMsg: "path contains invalid characters"},
		{name: "TrailingSlash_Fails", path: "/a/", expectedErrorMsg: "path cannot end with '/'"},
		{name: "MissingLeadingSlashWithParts_Fails", path: "a/b", expectedErrorMsg: "path must start with '/'"},
		{name: "MissingLeadingSlashSinglePart_Fails", path: "a", expectedErrorMsg: "path must start with '/'"},
		{name: "EmptyPartInMiddle_Fails", path: "/a//b", expectedErrorMsg: "path parts cannot be empty"},
		{name: "EmptyPartAtStart_Fails", path: "//a", expectedErrorMsg: "path parts cannot be empty"},
		{name: "ParameterMissingClosingBrace_Fails", path: "/a/{b", expectedErrorMsg: "path parameters must start with '{' and end with '}'"},
		{name: "ParameterMissingOpeningBrace_Fails", path: "/a/b}", expectedErrorMsg: "path parameters must start with '{' and end with '}'"},
		{name: "ParameterHasExtraOpeningBrace_Fails", path: "/a/{{b}", expectedErrorMsg: "path parameters must have only one '{' and '}'"},
		{name: "ParameterHasExtraClosingBrace_Fails", path: "/a/{b}}", expectedErrorMsg: "path parameters must have only one '{' and '}'"},
		{name: "EmptyParameter_Fails", path: "/a/{}", expectedErrorMsg: "path parameters cannot be empty"},
		{name: "DuplicateParameterPart_Fails", path: "/a/{b}/{b}", expectedErrorMsg: "path parts must be unique"},
		{name: "DuplicateLiteralPartTwo_Fails", path: "/a/a", expectedErrorMsg: "path parts must be unique"},
		{name: "DuplicateLiteralPartThree_Fails", path: "/a/b/a", expectedErrorMsg: "path parts must be unique"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertAPIPathValidation(t, tc.path, tc.expectedErrorMsg)
		})
	}
}

func TestPathValidation_NonStringReferenceField_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path int `validate:"api_path"`
	}
	test := testStruct{Path: 1}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "must be a string")
}

func TestPathValidation_NonStringPointerField_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path *int `validate:"api_path"`
	}
	i := 0
	test := testStruct{Path: &i}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "must be a string")
}

func TestPathValidation_NilPointerString_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path *string `validate:"api_path"`
	}
	test := testStruct{Path: nil}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "value is nil")
}

func assertAPIPathValidation(t *testing.T, path string, expectedErrorMsg string) {
	t.Helper()

	type testStructRef struct {
		Path string `validate:"api_path"`
	}
	err := validation.Struct(&testStructRef{Path: path})
	if expectedErrorMsg != "" {
		assert.ErrorPart(t, err, expectedErrorMsg)
	} else {
		assert.NoError(t, err)
	}

	type testStructPtr struct {
		Path *string `validate:"api_path"`
	}
	pathCopy := path
	err = validation.Struct(&testStructPtr{Path: &pathCopy})
	if expectedErrorMsg != "" {
		assert.ErrorPart(t, err, expectedErrorMsg)
	} else {
		assert.NoError(t, err)
	}
}
