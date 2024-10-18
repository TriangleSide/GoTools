package validation

import (
	"reflect"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	t.Run("when Error is called on a struct Violation it should a formatted message", func(t *testing.T) {
		violation := NewViolation("test", &CallbackParameters{
			StructValidation: true,
			StructValue: reflect.ValueOf(struct {
				Value int
			}{}),
			StructFieldName: "Value",
			Value:           reflect.ValueOf(1),
			Parameters:      "parameters",
		}, "test message")
		assert.Equals(t, violation.Error(), "validation failed on field 'Value' with validator 'test' and parameters 'parameters' because test message")
	})
}
