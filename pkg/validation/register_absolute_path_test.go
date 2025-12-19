package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestAbsolutePathValidator_VariousInputs_ReturnsExpectedErrors(t *testing.T) {
	t.Parallel()

	type testCaseDefinition struct {
		Name             string
		Setup            func(t *testing.T) any
		Value            any
		ExpectedErrorMsg string
	}

	testCases := []testCaseDefinition{
		{
			Name:             "invalid path segments with double dot returns error",
			Value:            "/tmp/../etc/passwd",
			ExpectedErrorMsg: "path '/tmp/../etc/passwd' is not valid",
		},
		{
			Name:             "invalid path with multiple double dot segments returns error",
			Value:            "/a/b/../../c/../d",
			ExpectedErrorMsg: "path '/a/b/../../c/../d' is not valid",
		},
		{
			Name: "path with single dot segment returns error",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				return tempDir + "/."
			},
			ExpectedErrorMsg: "is not valid",
		},
		{
			Name: "string with existing absolute file succeeds",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				f, err := os.CreateTemp(tempDir, "tempfile-*")
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				return f.Name()
			},
			ExpectedErrorMsg: "",
		},
		{
			Name: "string with existing absolute directory succeeds",
			Setup: func(t *testing.T) any {
				t.Helper()
				return t.TempDir()
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:             "string with non-existing absolute path returns error",
			Value:            "/non/existing/path/that/does/not/exist",
			ExpectedErrorMsg: "path '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			Name:             "relative path returns error",
			Value:            "relative/path",
			ExpectedErrorMsg: "path 'relative/path' is not absolute",
		},
		{
			Name:             "empty string returns error",
			Value:            "",
			ExpectedErrorMsg: "path '' is not absolute",
		},
		{
			Name:             "non-string value returns error",
			Value:            123,
			ExpectedErrorMsg: "value must be a string for the absolute_path validator",
		},
		{
			Name:             "nil pointer returns error",
			Value:            (*string)(nil),
			ExpectedErrorMsg: "value is nil",
		},
		{
			Name: "pointer to string with existing absolute file succeeds",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				f, err := os.CreateTemp(tempDir, "tempfile-*")
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				return ptr.Of(f.Name())
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:             "pointer to string with non-existing absolute path returns error",
			Value:            ptr.Of("/non/existing/path/that/does/not/exist"),
			ExpectedErrorMsg: "path '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			Name:             "pointer to string with relative path returns error",
			Value:            ptr.Of("relative/path"),
			ExpectedErrorMsg: "path 'relative/path' is not absolute",
		},
		{
			Name: "interface with string value and existing absolute file succeeds",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				f, err := os.CreateTemp(tempDir, "tempfile-*")
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				return any(f.Name())
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:             "interface with string value and relative path returns error",
			Value:            any("relative/path"),
			ExpectedErrorMsg: "path 'relative/path' is not absolute",
		},
		{
			Name:             "nil interface returns error",
			Value:            any(nil),
			ExpectedErrorMsg: "the value is nil",
		},
		{
			Name:             "root path succeeds",
			Value:            "/",
			ExpectedErrorMsg: "",
		},
		{
			Name: "path with spaces succeeds if exists",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				pathWithSpaces := filepath.Join(tempDir, "path with spaces")
				err := os.Mkdir(pathWithSpaces, 0750)
				assert.NoError(t, err)
				return pathWithSpaces
			},
			ExpectedErrorMsg: "",
		},
		{
			Name: "symlink to valid file succeeds",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				f, err := os.CreateTemp(tempDir, "target-*")
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				targetFile := f.Name()
				symlinkPath := filepath.Join(tempDir, "symlink")
				err = os.Symlink(targetFile, symlinkPath)
				assert.NoError(t, err)
				return symlinkPath
			},
			ExpectedErrorMsg: "",
		},
		{
			Name: "broken symlink returns error",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				targetFile := filepath.Join(tempDir, "nonexistent")
				symlinkPath := filepath.Join(tempDir, "broken_symlink")
				err := os.Symlink(targetFile, symlinkPath)
				assert.NoError(t, err)
				return symlinkPath
			},
			ExpectedErrorMsg: "is not accessible",
		},
		{
			Name: "double pointer to string with existing file succeeds",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				f, err := os.CreateTemp(tempDir, "tempfile-*")
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				p := ptr.Of(f.Name())
				return &p
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:             "double pointer with nil inner pointer returns error",
			Value:            ptr.Of((*string)(nil)),
			ExpectedErrorMsg: "value is nil",
		},
		{
			Name: "path with special characters succeeds if exists",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				specialPath := filepath.Join(tempDir, "special!@#$%^&()_+-=")
				err := os.Mkdir(specialPath, 0750)
				assert.NoError(t, err)
				return specialPath
			},
			ExpectedErrorMsg: "",
		},
		{
			Name: "file in subdirectory succeeds",
			Setup: func(t *testing.T) any {
				t.Helper()
				tempDir := t.TempDir()
				subDir := filepath.Join(tempDir, "sub", "dir")
				err := os.MkdirAll(subDir, 0750)
				assert.NoError(t, err)
				f, err := os.CreateTemp(subDir, "file-*.txt")
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				return f.Name()
			},
			ExpectedErrorMsg: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			value := testCase.Value
			if testCase.Setup != nil {
				value = testCase.Setup(t)
			}
			err := validation.Var(value, "absolute_path")
			if testCase.ExpectedErrorMsg != "" {
				assert.ErrorPart(t, err, testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
