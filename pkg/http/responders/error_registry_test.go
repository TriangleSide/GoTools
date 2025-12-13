package responders_test

import (
	"net/http"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/http/responders"
	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type uniqueTestError struct{}

func (e *uniqueTestError) Error() string {
	return "unique test error"
}

func TestMustRegisterErrorResponse_ValidErrorType_DoesNotPanic(t *testing.T) {
	t.Parallel()
	responders.MustRegisterErrorResponse(http.StatusBadRequest, func(*uniqueTestError) *struct{} {
		return &struct{}{}
	})
}

func TestMustRegisterErrorResponse_RegisteredTwice_Panics(t *testing.T) {
	t.Parallel()
	assert.Panic(t, func() {
		responders.MustRegisterErrorResponse(http.StatusBadRequest, func(*testError) *struct{} {
			return &struct{}{}
		})
	})
}

func TestMustRegisterErrorResponse_PointerGeneric_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		responders.MustRegisterErrorResponse(http.StatusBadRequest, func(**testError) *struct{} {
			return &struct{}{}
		})
	}, "registered error responses must be a struct")
}

func TestMustRegisterErrorResponse_NonErrorStruct_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		responders.MustRegisterErrorResponse(http.StatusBadRequest, func(*struct{}) *struct{} {
			return &struct{}{}
		})
	}, "must have an error interface")
}

func TestMustRegisterErrorResponse_NonStructResponse_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicPart(t, func() {
		responders.MustRegisterErrorResponse(http.StatusBadRequest, func(*testError) *int {
			return ptr.Of(0)
		})
	}, "response type must be a struct")
}
