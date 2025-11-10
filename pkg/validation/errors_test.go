package validation_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	t.Run("when Error is called on a struct Violation it should a formatted message", func(t *testing.T) {
		t.Parallel()
		violation := validation.NewViolation(&validation.CallbackParameters{
			Validator:          "test",
			IsStructValidation: true,
			StructValue: reflect.ValueOf(struct {
				Value int
			}{}),
			StructFieldName: "Value",
			Value:           reflect.ValueOf(1),
			Parameters:      "parameters",
		}, errors.New("test message"))
		assert.Equals(t, violation.Error(), "validation failed on field 'Value' with validator 'test' and parameters 'parameters' because test message")
	})
}
