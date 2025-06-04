package config

import (
	"fmt"
	"os"

	"github.com/TriangleSide/GoTools/pkg/stringcase"
	"github.com/TriangleSide/GoTools/pkg/structs"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

const (
	// ProcessorTag specifies which configuration processor should populate the field.
	ProcessorTag = "config"

	// DefaultTag supplies a fallback value when a processor does not return
	// a value for the field.
	DefaultTag = "config_default"

	// ProcessorTypeEnv identifies the environment variable processor.
	ProcessorTypeEnv = "ENV"
)

// config is the configuration for the ProcessAndValidate function.
type config struct {
	prefix string
}

// Option configures how Process operates.  The available options are mostly
// used by the environment processor but may be honoured by any registered
// processor.
type Option func(*config)

// SourceFunc fetches a configuration value for a field.
// It should return the value and whether it was found.
type SourceFunc func(fieldName string, fieldMetadata *structs.FieldMetadata, prefix string) (string, bool, error)

var processors = map[string]SourceFunc{}

// RegisterProcessor registers a SourceFunc for a given name.
func RegisterProcessor(name string, fn SourceFunc) {
	processors[name] = fn
}

// WithPrefix sets the prefix to look for in the environment variables.
// Given a struct field named Value and the prefix TEST, the processor will look for TEST_VALUE.
func WithPrefix(prefix string) Option {
	return func(p *config) {
		p.prefix = prefix
	}
}

// envSource fetches configuration values from environment variables. The
// variable name is derived from the struct field name converted to
// SNAKE_CASE. If a prefix is provided, it is prepended followed by an
// underscore.
func envSource(fieldName string, _ *structs.FieldMetadata, prefix string) (string, bool, error) {
	formattedEnvName := stringcase.CamelToSnake(fieldName)
	if prefix != "" {
		formattedEnvName = fmt.Sprintf("%s_%s", prefix, formattedEnvName)
	}

	envValue, hasEnvValue := os.LookupEnv(formattedEnvName)
	return envValue, hasEnvValue, nil
}

func init() {
	RegisterProcessor(ProcessorTypeEnv, envSource)
}

// Process sets struct field values using registered configuration sources.
// A field is processed only when it specifies the `config` tag with a source
// type. The environment source derives variable names from the struct field name
// converted to SNAKE_CASE. If WithPrefix is used, the prefix is prepended to the
// environment variable name separated by an underscore. If the source returns no
// value and a default is not provided via the `config_default` tag, an error is
// returned.
func Process[T any](opts ...Option) (*T, error) {
	cfg := &config{prefix: ""}

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

		value, found, err := fetcher(fieldName, fieldMetadata, cfg.prefix)
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

// ProcessAndValidate processes configuration values using registered processors
// and validates the resulting struct.
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
