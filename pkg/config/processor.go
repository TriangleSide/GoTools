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

// Options is the configuration copied to the SourceFunc.
type Options struct {
	Prefix string
}

// Option configures how Process operates.
type Option func(*Options)

// WithPrefix sets the prefix to look for in the source values. For the ENV processor, given a struct field named
// Value and the prefix TEST, the processor will look for TEST_VALUE.
func WithPrefix(prefix string) Option {
	return func(p *Options) {
		p.Prefix = prefix
	}
}

// Process sets struct field values using registered configuration sources. A field is processed only when it
// specifies the `config` tag with a source type. If the source returns no value and a default is not provided
// via the `config_default` tag, an error is returned.
func Process[T any](opts ...Option) (*T, error) {
	cfg := &Options{
		Prefix: "",
	}
	for _, opt := range opts {
		opt(cfg)
	}

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

		value, found, err := fetcher(fieldName, fieldMetadata, *cfg)
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
func ProcessAndValidate[T any](opts ...Option) (*T, error) {
	conf, err := Process[T](opts...)
	if err != nil {
		return nil, err
	}
	if err := validation.Struct(conf); err != nil {
		return nil, fmt.Errorf("failed while validating the configuration (%w)", err)
	}
	return conf, nil
}
