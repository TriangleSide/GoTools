package migration

import (
	"github.com/TriangleSide/GoTools/pkg/config"
)

// migrateConfig is configured by the Option type.
type migrateConfig struct {
	configProvider func() (*Config, error)
	registry       *Registry
}

// Option configures a migrateConfig instance.
type Option func(cfg *migrateConfig)

// configure applies options to the default migrateConfig values.
func configure(opts ...Option) *migrateConfig {
	migrateCfg := &migrateConfig{
		configProvider: func() (*Config, error) {
			return config.ProcessAndValidate[Config]()
		},
		registry: nil,
	}
	for _, opt := range opts {
		opt(migrateCfg)
	}
	return migrateCfg
}

// WithConfigProvider provides an Option to overwrite the configuration provider.
func WithConfigProvider(callback func() (*Config, error)) Option {
	return func(cfg *migrateConfig) {
		cfg.configProvider = callback
	}
}

// WithRegistry overwrites the registry used during migration orchestration.
func WithRegistry(reg *Registry) Option {
	return func(cfg *migrateConfig) {
		cfg.registry = reg
	}
}
