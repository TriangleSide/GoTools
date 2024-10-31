package responders_test

import (
	"net/http"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/http/responders"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestErrorRegistry(t *testing.T) {
	t.Parallel()

	t.Run("when the option to return an error is registered twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			responders.MustRegisterErrorResponse[testError](http.StatusBadRequest, func(err *testError) string {
				return "registered twice"
			})
		})
	})

	t.Run("when a pointer generic is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			responders.MustRegisterErrorResponse[*testError](http.StatusBadRequest, func(err **testError) string {
				return "pointer is registered"
			})
		}, "registered error responses must be a struct")
	})

	t.Run("when a struct that is not an error is registered it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			responders.MustRegisterErrorResponse[struct{}](http.StatusBadRequest, func(err *struct{}) string {
				return "error"
			})
		}, "must have an error interface")
	})
}
