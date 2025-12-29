package status_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/trace/status"
)

func TestCode_Unset_IsZeroValue(t *testing.T) {
	t.Parallel()
	var code status.Code
	assert.Equals(t, status.Unset, code)
}

func TestCode_Values_AreDistinct(t *testing.T) {
	t.Parallel()
	assert.True(t, status.Unset != status.Error)
	assert.True(t, status.Unset != status.Success)
	assert.True(t, status.Error != status.Success)
}

func TestCode_String_Unset_ReturnsUnset(t *testing.T) {
	t.Parallel()
	assert.Equals(t, "Unset", status.Unset.String())
}

func TestCode_String_Error_ReturnsError(t *testing.T) {
	t.Parallel()
	assert.Equals(t, "Error", status.Error.String())
}

func TestCode_String_Success_ReturnsSuccess(t *testing.T) {
	t.Parallel()
	assert.Equals(t, "Success", status.Success.String())
}

func TestCode_String_UnknownValue_ReturnsUnknown(t *testing.T) {
	t.Parallel()
	unknownCode := status.Code(999)
	assert.Equals(t, "Unknown", unknownCode.String())
}
