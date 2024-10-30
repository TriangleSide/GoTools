package config

import (
	"fmt"
	"os"

	"github.com/TriangleSide/GoBase/pkg/structs"
	"github.com/TriangleSide/GoBase/pkg/utils/stringcase"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

const (
	// FormatTag is the field name pre-processor.
	FormatTag = "config_format"

	// DefaultTag is the default to use in case there is no environment variable that matches the formatted field name.
	DefaultTag = "config_default"

	// FormatTypeSnake tells the processor to transform the field name into snake-case. StructField becomes STRUCT_FIELD.
	FormatTypeSnake = "snake"
)

// config is the configuration for the ProcessAndValidate function.
type config struct {
	prefix string
}

// Option is used to set parameters for the environment variable processor.
type Option func(*config)

// WithPrefix sets the prefix to look for in the environment variables.
// Given a struct field named Value and the prefix TEST, the processor will look for TEST_VALUE.
func WithPrefix(prefix string) Option {
	return func(p *config) {
		p.prefix = prefix
	}
}

// Process sets the value of the struct fields from the associated environment variables.
func Process[T any](opts ...Option) (*T, error) {
	cfg := &config{
		prefix: "",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	fieldsMetadata := structs.Metadata[T]()
	conf := new(T)

	for fieldName, fieldMetadata := range fieldsMetadata.All() {
		formatValue, hasFormatTag := fieldMetadata.Tags().Fetch(FormatTag)
		if !hasFormatTag {
			continue
		}

		var formattedEnvName string
		switch formatValue {
		case FormatTypeSnake:
			formattedEnvName = stringcase.CamelToSnake(fieldName)
			if cfg.prefix != "" {
				formattedEnvName = fmt.Sprintf("%s_%s", cfg.prefix, formattedEnvName)
			}
		default:
			panic(fmt.Sprintf("invalid config format (%s)", formatValue))
		}

		envValue, hasEnvValue := os.LookupEnv(formattedEnvName)
		if hasEnvValue {
			if err := structs.AssignToField(conf, fieldName, envValue); err != nil {
				return nil, fmt.Errorf("failed to assign env var %s to field %s (%w)", envValue, fieldName, err)
			}
		} else {
			defaultValue, hasDefaultTag := fieldMetadata.Tags().Fetch(DefaultTag)
			if hasDefaultTag {
				if err := structs.AssignToField(conf, fieldName, defaultValue); err != nil {
					return nil, fmt.Errorf("failed to assign default value %s to field %s (%w)", defaultValue, fieldName, err)
				}
			}
		}
	}

	return conf, nil
}

// ProcessAndValidate sets the value of the struct fields from the associated environment variables.
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
