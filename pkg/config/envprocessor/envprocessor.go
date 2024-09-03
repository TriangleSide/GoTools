package envprocessor

import (
	"fmt"
	"os"

	reflectutils "github.com/TriangleSide/GoBase/pkg/utils/reflect"
	stringutils "github.com/TriangleSide/GoBase/pkg/utils/string"
	"github.com/TriangleSide/GoBase/pkg/validation"
)

// EnvName is used to indicate that the value of the variable is the name of an environment variable.
type EnvName string

const (
	// FormatTag is the field name pre-processor. Is a field is called StructField and has a snake-case formatter,
	// it is transformed into STRUCT_FIELD.
	FormatTag = "config_format"

	// DefaultTag is the default to use in case there is no environment variable that matches the formatted field name.
	DefaultTag = "config_default"

	// FormatTypeSnake tells the processor to transform the field name into snake-case. StructField becomes STRUCT_FIELD.
	FormatTypeSnake = "snake"
)

// Config is the configuration for the ProcessAndValidate function.
type Config struct {
	prefix string
}

// Option is used to set parameters for the environment variable processor.
type Option func(*Config) error

// WithPrefix sets the prefix to look for in the environment variables.
// Given a struct field named Value and the prefix TEST, the processor will look for TEST_VALUE.
func WithPrefix(prefix string) Option {
	return func(p *Config) error {
		p.prefix = prefix
		return nil
	}
}

// ProcessAndValidate fills out the fields of a struct from the environment variables.
func ProcessAndValidate[T any](opts ...Option) (*T, error) {
	config := &Config{
		prefix: "",
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, fmt.Errorf("failed to set the options for the configuration processor (%s)", err.Error())
		}
	}

	fieldsMetadata := reflectutils.FieldsToMetadata[T]()
	conf := new(T)

	for fieldName, fieldMetadata := range fieldsMetadata {
		formatValue, hasFormatTag := fieldMetadata.Tags[FormatTag]
		if !hasFormatTag {
			continue
		}

		var formattedEnvName string
		switch formatValue {
		case FormatTypeSnake:
			formattedEnvName = stringutils.CamelToUpperSnake(fieldName)
			if config.prefix != "" {
				formattedEnvName = fmt.Sprintf("%s_%s", config.prefix, formattedEnvName)
			}
		default:
			panic(fmt.Sprintf("invalid config format (%s)", formatValue))
		}

		envValue, hasEnvValue := os.LookupEnv(formattedEnvName)
		if hasEnvValue {
			if err := reflectutils.AssignToField(conf, fieldName, envValue); err != nil {
				return nil, fmt.Errorf("failed to assign env var %s to field %s", envValue, fieldName)
			}
		} else {
			defaultValue, hasDefaultTag := fieldMetadata.Tags[DefaultTag]
			if hasDefaultTag {
				if err := reflectutils.AssignToField(conf, fieldName, defaultValue); err != nil {
					return nil, fmt.Errorf("failed to assign default value %s to field %s", defaultValue, fieldName)
				}
			}
		}
	}

	if err := validation.Struct(conf); err != nil {
		return nil, fmt.Errorf("failed while validating the configuration (%s)", err.Error())
	}

	return conf, nil
}
