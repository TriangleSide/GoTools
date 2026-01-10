package validation_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/ptr"
	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
	"github.com/TriangleSide/go-toolkit/pkg/validation"
)

func TestAbsolutePathValidator_InvalidPathSegmentsWithDoubleDot_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("/tmp/../etc/passwd", "absolute_path")
	assert.ErrorPart(t, err, "path '/tmp/../etc/passwd' is not valid")
}

func TestAbsolutePathValidator_InvalidPathWithMultipleDoubleDotSegments_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var("/a/b/../../c/../d", "absolute_path")
	assert.ErrorPart(t, err, "path '/a/b/../../c/../d' is not valid")
}

func TestAbsolutePathValidator_PathWithSingleDotSegment_ReturnsError(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	err := validation.Var(tempDir+"/.", "absolute_path")
	assert.ErrorPart(t, err, "is not valid")
}

func TestAbsolutePathValidator_StringWithExistingAbsoluteFile_Succeeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	f, err := os.CreateTemp(tempDir, "tempfile-*")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	err = validation.Var(f.Name(), "absolute_path")
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
	tempDir := t.TempDir()
	f, err := os.CreateTemp(tempDir, "tempfile-*")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	err = validation.Var(ptr.Of(f.Name()), "absolute_path")
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
	tempDir := t.TempDir()
	f, err := os.CreateTemp(tempDir, "tempfile-*")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	err = validation.Var(any(f.Name()), "absolute_path")
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

func TestAbsolutePathValidator_RootPath_Succeeds(t *testing.T) {
	t.Parallel()
	err := validation.Var("/", "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_PathWithSpacesIfExists_Succeeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	pathWithSpaces := filepath.Join(tempDir, "path with spaces")
	err := os.Mkdir(pathWithSpaces, 0750)
	assert.NoError(t, err)
	err = validation.Var(pathWithSpaces, "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_SymlinkToValidFile_Succeeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	f, err := os.CreateTemp(tempDir, "target-*")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	targetFile := f.Name()
	symlinkPath := filepath.Join(tempDir, "symlink")
	err = os.Symlink(targetFile, symlinkPath)
	assert.NoError(t, err)
	err = validation.Var(symlinkPath, "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_BrokenSymlink_ReturnsError(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "nonexistent")
	symlinkPath := filepath.Join(tempDir, "broken_symlink")
	err := os.Symlink(targetFile, symlinkPath)
	assert.NoError(t, err)
	err = validation.Var(symlinkPath, "absolute_path")
	assert.ErrorPart(t, err, "is not accessible")
}

func TestAbsolutePathValidator_DoublePointerToStringWithExistingFile_Succeeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	f, err := os.CreateTemp(tempDir, "tempfile-*")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	p := ptr.Of(f.Name())
	err = validation.Var(&p, "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_DoublePointerWithNilInnerPointer_ReturnsError(t *testing.T) {
	t.Parallel()
	err := validation.Var(ptr.Of((*string)(nil)), "absolute_path")
	assert.ErrorPart(t, err, "value is nil")
}

func TestAbsolutePathValidator_PathWithSpecialCharactersIfExists_Succeeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	specialPath := filepath.Join(tempDir, "special!@#$%^&()_+-=")
	err := os.Mkdir(specialPath, 0750)
	assert.NoError(t, err)
	err = validation.Var(specialPath, "absolute_path")
	assert.NoError(t, err)
}

func TestAbsolutePathValidator_FileInSubdirectory_Succeeds(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "sub", "dir")
	err := os.MkdirAll(subDir, 0750)
	assert.NoError(t, err)
	f, err := os.CreateTemp(subDir, "file-*.txt")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())
	err = validation.Var(f.Name(), "absolute_path")
	assert.NoError(t, err)
}
