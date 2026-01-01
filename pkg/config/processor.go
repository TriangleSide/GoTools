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

	for fieldName, fieldMetadata := range fieldsMetadata {
		processorType, hasProcessorTag := fieldMetadata.Tags()[ProcessorTag]
		if !hasProcessorTag {
			continue
		}

		fetcherNotCast, ok := processors.Load(processorType)
		if !ok {
			return nil, &ProcessorNotRegisteredError{ProcessorName: processorType}
		}
		fetcher := fetcherNotCast.(SourceFunc)

		value, found, err := fetcher(fieldName, fieldMetadata)
		if err != nil {
			return nil, &SourceFetchError{FieldName: fieldName, ProcessorName: processorType, Err: err}
		}

		if !found {
			defaultValue, hasDefaultTag := fieldMetadata.Tags()[DefaultTag]
			if !hasDefaultTag {
				return nil, &NoValueFoundError{FieldName: fieldName}
			}

			value = defaultValue
			if err := structs.AssignToField(conf, fieldName, value); err != nil {
				return nil, &FieldAssignmentError{FieldName: fieldName, Value: value, Err: err}
			}

			continue
		}

		if err := structs.AssignToField(conf, fieldName, value); err != nil {
			return nil, &FieldAssignmentError{FieldName: fieldName, Value: value, Err: err}
		}
	}

	return conf, nil
}

// ProcessAndValidate processes configuration and validates the resulting struct.
func ProcessAndValidate[T any]() (*T, error) {
	conf, err := Process[T]()
	if err != nil {
		return nil, fmt.Errorf("failed to process configuration: %w", err)
	}

	if err := validation.Struct(conf); err != nil {
		return nil, fmt.Errorf("failed while validating the configuration: %w", err)
	}

	return conf, nil
}
