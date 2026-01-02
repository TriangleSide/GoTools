package config_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func init() {
	config.MustRegisterProcessor("TWICE", func(string, *structs.FieldMetadata) (string, bool, error) {
		return "", true, nil
	})
}

func TestMustRegisterProcessor_ReRegistered_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicExact(t, func() {
		config.MustRegisterProcessor("TWICE", func(string, *structs.FieldMetadata) (string, bool, error) {
			return "", true, nil
		})
	}, "processor with name \"TWICE\" already registered")
}

func TestMustRegisterProcessor_NilSourceFunc_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicExact(t, func() {
		config.MustRegisterProcessor("NIL_SOURCE", nil)
	}, "processor \"NIL_SOURCE\" requires a non-nil sourcing function")
}
