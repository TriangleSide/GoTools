package config

import (
	"sync"

	"github.com/TriangleSide/GoTools/pkg/structs"
)

var (
	// processors maps processor type names to their SourceFunc implementations.
	processors = sync.Map{}
)

// SourceFunc fetches a configuration value for a field. It should return the value and whether it was found.
type SourceFunc func(fieldName string, fieldMetadata *structs.FieldMetadata) (string, bool, error)

// MustRegisterProcessor registers a SourceFunc for a given processor name.
func MustRegisterProcessor(name string, fn SourceFunc) {
	if fn == nil {
		panic(&NilSourceFuncError{ProcessorName: name})
	}
	_, found := processors.LoadOrStore(name, fn)
	if found {
		panic(&ProcessorAlreadyRegisteredError{ProcessorName: name})
	}
}
