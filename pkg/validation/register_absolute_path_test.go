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

func TestAbsolutePathValidator_InvalidPathSegments_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("/tmp/../etc/passwd", "absolute_path")
	assert.ErrorPart(t, err, "path '/tmp/../etc/passwd' is not valid")
}

func TestAbsolutePathValidator_StringWithExistingAbsoluteFile_Succeeds(t *testing.T) {
	t.Parallel()
	value := createAbsoluteTempFile(t)
	err := validation.Var(value, "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_StringWithExistingAbsoluteDirectory_Succeeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	err := validation.Var(tempDir, "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_StringWithNonExistingAbsolutePath_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("/non/existing/path/that/does/not/exist", "absolute_path")
	assert.ErrorPart(t, err, "path '/non/existing/path/that/does/not/exist' is not accessible")
}

func TestAbsolutePathValidator_RelativePath_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("relative/path", "absolute_path")
	assert.ErrorPart(t, err, "path 'relative/path' is not absolute")
}

func TestAbsolutePathValidator_EmptyString_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("", "absolute_path")
	assert.ErrorPart(t, err, "path '' is not absolute")
}

func TestAbsolutePathValidator_NonStringValue_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(123, "absolute_path")
	assert.ErrorPart(t, err, "value must be a string for the absolute_path validator")
}

func TestAbsolutePathValidator_NilPointer_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var((*string)(nil), "absolute_path")
	assert.ErrorPart(t, err, "value is nil")
}

func TestAbsolutePathValidator_PointerToStringWithExistingAbsoluteFile_Succeeds(t *testing.T) {
	t.Parallel()
	value := ptr.Of(createAbsoluteTempFile(t))
	err := validation.Var(value, "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_PointerToStringWithNonExistingAbsolutePath_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("/non/existing/path/that/does/not/exist"), "absolute_path")
	assert.ErrorPart(t, err, "path '/non/existing/path/that/does/not/exist' is not accessible")
}

func TestAbsolutePathValidator_PointerToStringWithRelativePath_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of("relative/path"), "absolute_path")
	assert.ErrorPart(t, err, "path 'relative/path' is not absolute")
}

func TestAbsolutePathValidator_InterfaceWithStringValueAndExistingAbsoluteFile_Succeeds(t *testing.T) {
	t.Parallel()
	value := any(createAbsoluteTempFile(t))
	err := validation.Var(value, "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_InterfaceWithStringValueAndRelativePath_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any("relative/path"), "absolute_path")
	assert.ErrorPart(t, err, "path 'relative/path' is not absolute")
}

func TestAbsolutePathValidator_NilInterface_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(any(nil), "absolute_path")
	assert.ErrorPart(t, err, "the value is nil")
}
