package config

import (
	"github.com/kelseyhightower/envconfig"
	"intelligence/pkg/validation"
)

// ProcessConfiguration fills out the member fields of a struct from environment variables.
func ProcessConfiguration[T any]() (*T, error) {
	conf := new(T)
	if err := envconfig.Process("", conf); err != nil {
		return nil, err
	}
	if err := validation.Validate(conf); err != nil {
		return nil, err
	}
	return conf, nil
}
