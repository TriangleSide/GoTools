package validation_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestRegistry(t *testing.T) {
	t.Parallel()

	t.Run("when a validation is registered twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.PanicPart(t, func() {
			for range 2 {
				validation.MustRegisterValidator("test_validator_name", func(parameters *validation.CallbackParameters) *validation.CallbackResult { return nil })
			}
		}, "named test_validator_name already exists")
	})
}
