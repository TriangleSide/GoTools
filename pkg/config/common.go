package config

import (
	"github.com/kelseyhightower/envconfig"

	"intelligence/pkg/validation"
)

// processorParameters holds parameters for the configuration processor.
type processorParameters struct {
	prefix string
}

// Option is used to set parameters for the configuration processor.
type Option func(*processorParameters)

// WithPrefix sets the prefix to look for in the environment variables.
// Given a struct field named Value and the prefix TEST, the processor will look for TEST_VALUE.
func WithPrefix(prefix string) Option {
	return func(p *processorParameters) {
		p.prefix = prefix
	}
}

// ProcessAndValidate fills out the member fields of a struct from the environment variables.
func ProcessAndValidate[T any](options ...Option) (*T, error) {
	config := &processorParameters{
		prefix: "",
	}
	for _, option := range options {
		option(config)
	}

	conf := new(T)
	if err := envconfig.Process(config.prefix, conf); err != nil {
		return nil, err
	}
	if err := validation.Validate(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
