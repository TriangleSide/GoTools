package config

import (
	"fmt"

	"github.com/TriangleSide/GoTools/pkg/structs"
)

var (
	// processors is a map of ProcessorTag value to how the values are fetched.
	processors = map[string]SourceFunc{}
)

// SourceFunc fetches a configuration value for a field. It should return the value and whether it was found.
type SourceFunc func(fieldName string, fieldMetadata *structs.FieldMetadata) (string, bool, error)

// MustRegisterProcessor registers a SourceFunc for a given processor name.
func MustRegisterProcessor(name string, fn SourceFunc) {
	if fn == nil {
		panic(fmt.Sprintf("Must register a non-nil SourceFunc for the %s configuration processor.", name))
	}
	if _, exists := processors[name]; exists {
		panic(fmt.Sprintf("Processor with name %q already registered.", name))
	}
	processors[name] = fn
}
