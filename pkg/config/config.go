// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

package config

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	reflectutils "intelligence/pkg/utils/reflect"
	"intelligence/pkg/validation"
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

// envVarsProcessorParameters holds information needed by the environment variable processor.
type envVarsProcessorParameters struct {
	prefix string
}

// Option is used to set parameters for the environment variable processor.
type Option func(*envVarsProcessorParameters)

// WithPrefix sets the prefix to look for in the environment variables.
// Given a struct field named Value and the prefix TEST, the processor will look for TEST_VALUE.
func WithPrefix(prefix string) Option {
	return func(p *envVarsProcessorParameters) {
		p.prefix = prefix
	}
}

// ProcessAndValidate fills out the fields of a struct from the environment variables.
func ProcessAndValidate[T any](options ...Option) (*T, error) {
	config := &envVarsProcessorParameters{
		prefix: "",
	}
	for _, option := range options {
		option(config)
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
			formattedEnvName = camelToSnake(fieldName)
			if config.prefix != "" {
				formattedEnvName = fmt.Sprintf("%s_%s", config.prefix, formattedEnvName)
			}
		default:
			panic(fmt.Sprintf("invalid format tag value (%s)", formatValue))
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

// camelToSnake converts a camel-case string to a upper-case snake-case format.
//
//	MyCamelCase becomes MY_CAMEL_CASE.
//	myCamelCase becomes MY_CAMEL_CASE.
//	CAMELCase becomes CAMEL_CASE.
func camelToSnake(str string) string {
	var snake strings.Builder
	for i, r := range str {
		if i > 0 && unicode.IsUpper(r) && (i+1 < len(str) && unicode.IsLower(rune(str[i+1])) || unicode.IsLower(rune(str[i-1]))) {
			snake.WriteRune('_')
		}
		snake.WriteRune(unicode.ToUpper(r))
	}
	return snake.String()
}
