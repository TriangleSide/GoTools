package config_test

import (
	"errors"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestMustRegisterProcessor_ReRegistered_Panics(t *testing.T) {
	t.Parallel()
	config.MustRegisterProcessor("TWICE", func(string, *structs.FieldMetadata) (string, bool, error) {
		return "", true, nil
	})
	assert.PanicExact(t, func() {
		config.MustRegisterProcessor("TWICE", func(string, *structs.FieldMetadata) (string, bool, error) {
			return "", true, nil
		})
	}, "processor with name \"TWICE\" already registered")
}

func TestMustRegisterProcessor_ReRegistered_PanicsWithProcessorAlreadyRegisteredError(t *testing.T) {
	t.Parallel()
	config.MustRegisterProcessor("TWICE_STRUCT", func(string, *structs.FieldMetadata) (string, bool, error) {
		return "", true, nil
	})

	var panicValue any
	func() {
		defer func() {
			panicValue = recover()
		}()
		config.MustRegisterProcessor("TWICE_STRUCT", func(string, *structs.FieldMetadata) (string, bool, error) {
			return "", true, nil
		})
	}()

	err, ok := panicValue.(error)
	assert.True(t, ok)

	var regErr *config.ProcessorAlreadyRegisteredError
	assert.True(t, errors.As(err, &regErr))
	assert.Equals(t, regErr.ProcessorName, "TWICE_STRUCT")
}

func TestMustRegisterProcessor_NilSourceFunc_Panics(t *testing.T) {
	t.Parallel()
	assert.PanicExact(t, func() {
		config.MustRegisterProcessor("NIL_SOURCE", nil)
	}, "processor \"NIL_SOURCE\" requires a non-nil sourcing function")
}

func TestMustRegisterProcessor_NilSourceFunc_PanicsWithNilSourceFuncError(t *testing.T) {
	t.Parallel()

	var panicValue any
	func() {
		defer func() {
			panicValue = recover()
		}()
		config.MustRegisterProcessor("NIL_SOURCE_STRUCT", nil)
	}()

	err, ok := panicValue.(error)
	assert.True(t, ok)

	var nilErr *config.NilSourceFuncError
	assert.True(t, errors.As(err, &nilErr))
	assert.Equals(t, nilErr.ProcessorName, "NIL_SOURCE_STRUCT")
}
