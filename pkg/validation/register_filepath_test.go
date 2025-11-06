package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestFilepathValidator(t *testing.T) {
	testCases := []struct {
		name          string
		value         any
		setup         func(t *testing.T) interface{}
		expectedError string
	}{
		{
			name: "when value is a string with existing file it should succeed",
			setup: func(t *testing.T) interface{} {
				t.Helper()
				tempDir := t.TempDir()
				tempFile := filepath.Join(tempDir, "testfile")
				f, err := os.Create(tempFile)
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				return tempFile
			},
			expectedError: "",
		},
		{
			name:          "when value is a string with non-existing file it should return an error",
			value:         "/non/existing/path/that/does/not/exist",
			expectedError: "file '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			name:          "when value is a non-string value, it should return an error",
			value:         123,
			expectedError: "value must be a string for the filepath validator",
		},
		{
			name:          "when value is a nil pointer, it should fail",
			value:         (*string)(nil),
			expectedError: "found nil while dereferencing",
		},
		{
			name: "when value is a pointer to string with existing file it should succeed",
			setup: func(t *testing.T) interface{} {
				tempDir := t.TempDir()
				tempFile := filepath.Join(tempDir, "testfile")
				f, err := os.Create(tempFile)
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				return ptr.Of(tempFile)
			},
			expectedError: "",
		},
		{
			name:          "when value is a pointer to string with non-existing file, it should return an error",
			value:         ptr.Of("/non/existing/path/that/does/not/exist"),
			expectedError: "file '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			name: "when value is an interface with string value and existing file it should succeed",
			setup: func(t *testing.T) interface{} {
				tempDir := t.TempDir()
				tempFile := filepath.Join(tempDir, "testfile")
				f, err := os.Create(tempFile)
				assert.NoError(t, err)
				assert.NoError(t, f.Close())
				return interface{}(tempFile)
			},
			expectedError: "",
		},
		{
			name:          "when value is an interface with string value and non-existing file it should return an error",
			value:         interface{}("/non/existing/path/that/does/not/exist"),
			expectedError: "file '/non/existing/path/that/does/not/exist' is not accessible",
		},
		{
			name:          "when value is a nil interface it should fail",
			value:         interface{}(nil),
			expectedError: "the value is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var value interface{}
			if tc.setup != nil {
				value = tc.setup(t)
			} else {
				value = tc.value
			}

			err := validation.Var(value, "filepath")
			if tc.expectedError != "" {
				assert.ErrorPart(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
