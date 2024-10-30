package validation

import (
	"reflect"
	"testing"

	"github.com/TriangleSide/GoBase/pkg/ptr"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func TestCommonFunctions(t *testing.T) {
	t.Parallel()

	t.Run("when nil is passed to DereferenceValue it should do nothing", func(t *testing.T) {
		t.Parallel()
		invalidValue := reflect.ValueOf(nil)
		assert.False(t, DereferenceValue(&invalidValue))
	})

	t.Run("when a pointer chain of int is created value DereferenceValue should return false", func(t *testing.T) {
		t.Parallel()
		ptrChainValue := reflect.ValueOf(ptr.Of(ptr.Of(ptr.Of(1))))
		assert.Equals(t, ptrChainValue.Kind(), reflect.Ptr)
		assert.True(t, DereferenceValue(&ptrChainValue))
		assert.Equals(t, ptrChainValue.Kind(), reflect.Int)
	})
}
