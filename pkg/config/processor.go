package config

import (
	"fmt"

	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

const (
	// ProcessorTag specifies which configuration processor should populate the struct field.
	ProcessorTag = "config"

	// DefaultTag supplies a fallback value when a processor does not return a value for the field.
	DefaultTag = "config_default"
)

// Process sets struct field values using registered configuration sources. A field is processed only when it
// specifies the `config` tag with a source type. If the source returns no value and a default is not provided
// via the `config_default` tag, an error is returned.
func Process[T any]() (*T, error) {
	fieldsMetadata := structs.Metadata[T]()
	conf := new(T)

	for fieldName, fieldMetadata := range fieldsMetadata.All() {
		processorType, hasProcessorTag := fieldMetadata.Tags().Fetch(ProcessorTag)
		if !hasProcessorTag {
			continue
		}

		fetcher, ok := processors[processorType]
		if !ok {
			return nil, fmt.Errorf("processor %s not registered", processorType)
		}

		value, found, err := fetcher(fieldName, fieldMetadata)
		if err != nil {
			return nil, err
		}

		if !found {
			defaultValue, hasDefaultTag := fieldMetadata.Tags().Fetch(DefaultTag)
			if !hasDefaultTag {
				return nil, fmt.Errorf("no value found for field %s", fieldName)
			}
			value = defaultValue
			if err := structs.AssignToField(conf, fieldName, value); err != nil {
				return nil, fmt.Errorf("failed to assign default value %s to field %s (%w)", value, fieldName, err)
			}
			continue
		}

		if err := structs.AssignToField(conf, fieldName, value); err != nil {
			return nil, fmt.Errorf("failed to assign value %s to field %s (%w)", value, fieldName, err)
		}
	}

	return conf, nil
}

// ProcessAndValidate processes configuration and validates the resulting struct.
func ProcessAndValidate[T any]() (*T, error) {
	conf, err := Process[T]()
	if err != nil {
		return nil, err
	}
	if err := validation.Struct(conf); err != nil {
		return nil, fmt.Errorf("failed while validating the configuration (%w)", err)
	}
	return conf, nil
}
