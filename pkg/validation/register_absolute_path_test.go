package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func createAbsoluteTempFile(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "tempfile")
	f, err := os.Create(tempFile) // nolint:gosec
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	return tempFile
}

func TestAbsolutePathValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		value         any
		setup         func(t *testing.T) any
		expectedError string
	}{
		{
			name:          "when value contains invalid path segments it should fail",
			value:         "/tmp/../etc/passwd",
			expectedError: "path '/tmp/../etc/passwd' is not valid",
		},
		{
			name: "when value is a string with existing absolute file it should succeed",
			setup: func(t *testing.T) any {
				t.Helper()
				return createAbsoluteTempFile(t)
			},
			expectedError: "",
		},
		{
			name:          "when value is a string with non-existing absolute path it should return an error",
			value:         "/non/existing/path/that/does/not/exist",
			expectedError: "path '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			name:          "when value is a relative path even if it exists it should return an error",
			value:         "relative/path",
			expectedError: "path 'relative/path' is not absolute",
		},
		{
			name:          "when value is a non-string value, it should return an error",
			value:         123,
			expectedError: "value must be a string for the absolute_path validator",
		},
		{
			name:          "when value is a nil pointer, it should fail",
			value:         (*string)(nil),
			expectedError: "value is nil",
		},
		{
			name: "when value is a pointer to string with existing absolute file it should succeed",
			setup: func(t *testing.T) any {
				t.Helper()
				return ptr.Of(createAbsoluteTempFile(t))
			},
			expectedError: "",
		},
		{
			name:          "when value is a pointer to string with non-existing absolute path it should return an error",
			value:         ptr.Of("/non/existing/path/that/does/not/exist"),
			expectedError: "path '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			name: "when value is an interface with string value and existing absolute file it should succeed",
			setup: func(t *testing.T) any {
				t.Helper()
				return any(createAbsoluteTempFile(t))
			},
			expectedError: "",
		},
		{
			name:          "when value is an interface with string value and relative path it should return an error",
			value:         any("relative/path"),
			expectedError: "path 'relative/path' is not absolute",
		},
		{
			name:          "when value is a nil interface it should fail",
			value:         any(nil),
			expectedError: "the value is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var value any
			if tc.setup != nil {
				value = tc.setup(t)
			} else {
				value = tc.value
			}

			err := validation.Var(value, "absolute_path")
			if tc.expectedError != "" {
				assert.ErrorPart(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
