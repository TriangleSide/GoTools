package validation_test

import (
	"os"
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/ptr"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
	"github.com/TriangleSide/go-toolkit/pkg/validation"
)

func createTempFile(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()
	f, err := os.CreateTemp(tempDir, "tempfile-*")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	return f.Name()
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

func TestFilepathValidator_StringWithExistingFile_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var(createTempFile(t), "filepath")
	assert.NoError(t, err)
}

func TestFilepathValidator_StringWithNonExistingFile_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("/non/existing/path/that/does/not/exist", "filepath")
	assert.ErrorPart(t, err, "file '/non/existing/path/that/does/not/exist' is not accessible")
}

func TestFilepathValidator_NonStringValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(123, "filepath")
	assert.ErrorPart(t, err, "value must be a string for the filepath validator")
}

func TestFilepathValidator_NilPointer_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*string)(nil), "filepath")
	assert.ErrorPart(t, err, "value is nil")
}

func TestFilepathValidator_PointerToStringWithExistingFile_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of(createTempFile(t)), "filepath")
	assert.NoError(t, err)
}

func TestFilepathValidator_PointerToStringWithNonExistingFile_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("/non/existing/path/that/does/not/exist"), "filepath")
	assert.ErrorPart(t, err, "file '/non/existing/path/that/does/not/exist' is not accessible")
}

func TestFilepathValidator_InterfaceWithStringAndExistingFile_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var(any(createTempFile(t)), "filepath")
	assert.NoError(t, err)
}

func TestFilepathValidator_InterfaceWithStringAndNonExistingFile_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any("/non/existing/path/that/does/not/exist"), "filepath")
	assert.ErrorPart(t, err, "file '/non/existing/path/that/does/not/exist' is not accessible")
}

func TestFilepathValidator_NilInterface_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any(nil), "filepath")
	assert.ErrorPart(t, err, "the value is nil")
}

func TestFilepathValidator_EmptyString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "filepath")
	assert.ErrorPart(t, err, "file '' is not accessible")
}

func TestFilepathValidator_DirectoryPath_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var(t.TempDir(), "filepath")
	assert.NoError(t, err)
}
