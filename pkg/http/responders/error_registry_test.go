package responders_test

import (
	"net/http"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/responders"
	"github.com/TriangleSide/GoBase/pkg/ptr"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestErrorRegistry(t *testing.T) {
	t.Parallel()

	t.Run("when the option to return an error is registered twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			responders.MustRegisterErrorResponse(http.StatusBadRequest, func(err *testError) *struct{} {
				return &struct{}{}
			})
		})
	})

	t.Run("when a pointer generic is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			responders.MustRegisterErrorResponse(http.StatusBadRequest, func(err **testError) *struct{} {
				return &struct{}{}
			})
		}, "registered error responses must be a struct")
	})

	t.Run("when a struct that is not an error is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			responders.MustRegisterErrorResponse(http.StatusBadRequest, func(err *struct{}) *struct{} {
				return &struct{}{}
			})
		}, "must have an error interface")
	})

	t.Run("when the response type is not a struct it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			responders.MustRegisterErrorResponse(http.StatusBadRequest, func(err *testError) *int {
				return ptr.Of(0)
			})
		}, "response type must be a struct")
	})
}
