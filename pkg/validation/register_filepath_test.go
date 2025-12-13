package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func createTempFile(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "tempfile")
	f, err := os.Create(tempFile) // nolint:gosec
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	return tempFile
}

func TestFilepathValidator_StructField(t *testing.T) {
	t.Parallel()

	type Config struct {
		Path string `validate:"filepath"`
	}

	t.Run("when struct field has valid file path it should succeed", func(t *testing.T) {
		t.Parallel()
		cfg := Config{Path: createTempFile(t)}
		err := validation.Struct(&cfg)
		assert.NoError(t, err)
	})

	t.Run("when struct field has invalid file path it should fail", func(t *testing.T) {
		t.Parallel()
		cfg := Config{Path: "/non/existing/path"}
		err := validation.Struct(&cfg)
		assert.ErrorPart(t, err, "file '/non/existing/path' is not accessible")
	})
}

func TestFilepathValidator_VariousInputs_ReturnsExpectedErrors(t *testing.T) {
	t.Parallel()

	type testCaseDefinition struct {
		Name          string
		Value         any
		Setup         func(t *testing.T) any
		ExpectedError string
	}

	testCases := []testCaseDefinition{
		{
			Name: "when value is a string with existing file it should succeed",
			Setup: func(t *testing.T) any {
				t.Helper()
				return createTempFile(t)
			},
			ExpectedError: "",
		},
		{
			Name:          "when value is a string with non-existing file it should return an error",
			Value:         "/non/existing/path/that/does/not/exist",
			ExpectedError: "file '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			Name:          "when value is a non-string value, it should return an error",
			Value:         123,
			ExpectedError: "value must be a string for the filepath validator",
		},
		{
			Name:          "when value is a nil pointer, it should fail",
			Value:         (*string)(nil),
			ExpectedError: "value is nil",
		},
		{
			Name: "when value is a pointer to string with existing file it should succeed",
			Setup: func(t *testing.T) any {
				t.Helper()
				return ptr.Of(createTempFile(t))
			},
			ExpectedError: "",
		},
		{
			Name:          "when value is a pointer to string with non-existing file, it should return an error",
			Value:         ptr.Of("/non/existing/path/that/does/not/exist"),
			ExpectedError: "file '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			Name: "when value is an interface with string value and existing file it should succeed",
			Setup: func(t *testing.T) any {
				t.Helper()
				return any(createTempFile(t))
			},
			ExpectedError: "",
		},
		{
			Name:          "when value is an interface with string value and non-existing file it should return an error",
			Value:         any("/non/existing/path/that/does/not/exist"),
			ExpectedError: "file '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			Name:          "when value is a nil interface it should fail",
			Value:         any(nil),
			ExpectedError: "the value is nil",
		},
		{
			Name:          "when value is an empty string it should return an error",
			Value:         "",
			ExpectedError: "file '' is not accessible",
		},
		{
			Name: "when value is a directory path it should succeed",
			Setup: func(t *testing.T) any {
				t.Helper()
				return t.TempDir()
			},
			ExpectedError: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			var value any
			if testCase.Setup != nil {
				value = testCase.Setup(t)
			} else {
				value = testCase.Value
			}

			err := validation.Var(value, "filepath")
			if testCase.ExpectedError != "" {
				assert.ErrorPart(t, err, testCase.ExpectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
