package config_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestRegistry(t *testing.T) {
	t.Run("when a custom processor is re-registered it should panic", func(t *testing.T) {
		config.MustRegisterProcessor("TWICE", func(fieldName string, _ *structs.FieldMetadata, _ config.Options) (string, bool, error) {
			return "", true, nil
		})
		assert.PanicExact(t, func() {
			config.MustRegisterProcessor("TWICE", func(fieldName string, _ *structs.FieldMetadata, _ config.Options) (string, bool, error) {
				return "", true, nil
			})
		}, "Processor with name \"TWICE\" already registered.")
	})

	t.Run("when a custom processor is registered with a nil SourceFunc it should panic", func(t *testing.T) {
		assert.PanicExact(t, func() {
			config.MustRegisterProcessor("TWICE", nil)
		}, "Must register a non-nil SourceFunc for the TWICE configuration processor.")
	})
}
