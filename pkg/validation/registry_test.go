package validation

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestRegistry(t *testing.T) {
	t.Parallel()

	t.Run("when a validation is registered twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			MustRegisterValidator(RequiredValidatorName, func(parameters *CallbackParameters) error { return nil })
		}, "named required already exists")
	})
}
