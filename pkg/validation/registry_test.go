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
			for i := 0; i < 2; i++ {
				MustRegisterValidator("test_validator_name", func(parameters *CallbackParameters) *CallbackResult { return nil })
			}
		}, "named test_validator_name already exists")
	})
}
