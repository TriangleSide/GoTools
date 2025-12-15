package responders_test

import (
	"net/http"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/http/responders"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type uniqueTestError struct{}

func (e *uniqueTestError) Error() string {
	return "unique test error"
}

func TestMustRegisterErrorResponse_ValidErrorType_DoesNotPanic(t *testing.T) {
	t.Parallel()
	responders.MustRegisterErrorResponse(http.StatusBadRequest, func(*uniqueTestError) *responders.StandardErrorResponse {
		return &responders.StandardErrorResponse{}
	})
}

func TestMustRegisterErrorResponse_RegisteredTwice_Panics(t *testing.T) {
	t.Parallel()
	assert.Panic(t, func() {
		responders.MustRegisterErrorResponse(http.StatusBadRequest, func(*testError) *responders.StandardErrorResponse {
			return &responders.StandardErrorResponse{}
		})
	})
}

func TestMustRegisterErrorResponse_PointerGeneric_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		responders.MustRegisterErrorResponse(http.StatusBadRequest, func(**testError) *responders.StandardErrorResponse {
			return &responders.StandardErrorResponse{}
		})
	}, "cannot be a pointer")
}

func TestMustRegisterErrorResponse_TriplePointerGeneric_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		responders.MustRegisterErrorResponse(http.StatusBadRequest, func(***testError) *responders.StandardErrorResponse {
			return &responders.StandardErrorResponse{}
		})
	}, "cannot be a pointer")
}

func TestMustRegisterErrorResponse_NonErrorStruct_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		responders.MustRegisterErrorResponse(http.StatusBadRequest, func(*struct{}) *responders.StandardErrorResponse {
			return &responders.StandardErrorResponse{}
		})
	}, "error interface")
}
