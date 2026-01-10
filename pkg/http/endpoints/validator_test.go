package endpoints_test

import (
	"testing"

	_ "github.com/TriangleSide/go-toolkit/pkg/http/endpoints"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
	"github.com/TriangleSide/go-toolkit/pkg/validation"
)

func TestPathValidation_Root_Succeeds(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/", "")
}

func TestPathValidation_MultiPart_Succeeds(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/b/c/1/2/3", "")
}

func TestPathValidation_WithParameterPart_Succeeds(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/{b}/c", "")
}

func TestPathValidation_Empty_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "", "path cannot be empty")
}

func TestPathValidation_InvalidCharactersPlus_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/+", "path contains invalid characters")
}

func TestPathValidation_InvalidCharactersLeadingSpace_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, " /a", "path contains invalid characters")
}

func TestPathValidation_InvalidCharactersTrailingSpace_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a ", "path contains invalid characters")
}

func TestPathValidation_TrailingSlash_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/", "path cannot end with '/'")
}

func TestPathValidation_MissingLeadingSlashWithParts_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "a/b", "path must start with '/'")
}

func TestPathValidation_MissingLeadingSlashSinglePart_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "a", "path must start with '/'")
}

func TestPathValidation_EmptyPartInMiddle_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a//b", "path parts cannot be empty")
}

func TestPathValidation_EmptyPartAtStart_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "//a", "path parts cannot be empty")
}

func TestPathValidation_ParameterMissingClosingBrace_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/{b", "path parameters must start with '{' and end with '}'")
}

func TestPathValidation_ParameterMissingOpeningBrace_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/b}", "path parameters must start with '{' and end with '}'")
}

func TestPathValidation_ParameterHasExtraOpeningBrace_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/{{b}", "path parameters must have only one '{' and '}'")
}

func TestPathValidation_ParameterHasExtraClosingBrace_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/{b}}", "path parameters must have only one '{' and '}'")
}

func TestPathValidation_EmptyParameter_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/{}", "path parameters cannot be empty")
}

func TestPathValidation_DuplicateParameterPart_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/{b}/{b}", "path parts must be unique")
}

func TestPathValidation_DuplicateLiteralPartTwo_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/a", "path parts must be unique")
}

func TestPathValidation_DuplicateLiteralPartThree_ReturnsError(t *testing.T) {
	t.Parallel()
	assertRoutePathValidation(t, "/a/b/a", "path parts must be unique")
}

func TestPathValidation_NonStringReferenceField_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path int `validate:"api_endpoint_path"`
	}
	test := testStruct{Path: 1}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "must be a string")
}

func TestPathValidation_NonStringPointerField_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path *int `validate:"api_endpoint_path"`
	}
	i := 0
	test := testStruct{Path: &i}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "must be a string")
}

func TestPathValidation_NilPointerString_ReturnsError(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Path *string `validate:"api_endpoint_path"`
	}
	test := testStruct{Path: nil}
	err := validation.Struct(&test)
	assert.ErrorPart(t, err, "value is nil")
}

func assertRoutePathValidation(t *testing.T, path string, expectedErrorMsg string) {
	t.Helper()

	type testStructRef struct {
		Path string `validate:"api_endpoint_path"`
	}
	err := validation.Struct(&testStructRef{Path: path})
	if expectedErrorMsg != "" {
		assert.ErrorPart(t, err, expectedErrorMsg)
	} else {
		assert.NoError(t, err)
	}

	type testStructPtr struct {
		Path *string `validate:"api_endpoint_path"`
	}
	pathCopy := path
	err = validation.Struct(&testStructPtr{Path: &pathCopy})
	if expectedErrorMsg != "" {
		assert.ErrorPart(t, err, expectedErrorMsg)
	} else {
		assert.NoError(t, err)
	}
}
