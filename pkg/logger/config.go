package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/TriangleSide/GoTools/pkg/config"
)

// Config contains the values needed to configure the logger.
type Config struct {
	LogLevel string `config_format:"snake" config_default:"INFO" validate:"required,oneof=ERROR WARN INFO DEBUG TRACE"`
}

// loggerConfig is configured by the ConfigOption functions.
type loggerConfig struct {
	configProvider func() (*Config, error)
	outputProvider func() (io.Writer, error)
}

// ConfigOption sets values on the loggerConfig.
type ConfigOption func(*loggerConfig)

// WithConfigProvider sets the provider for the Config.
func WithConfigProvider(provider func() (*Config, error)) ConfigOption {
	return func(c *loggerConfig) {
		c.configProvider = provider
	}
}

// WithOutputProvider sets the logger output.
func WithOutputProvider(provider func() (io.Writer, error)) ConfigOption {
	return func(c *loggerConfig) {
		c.outputProvider = provider
	}
}

// MustConfigure parses the Config and sets values for the application logger.
func MustConfigure(opts ...ConfigOption) {
	cfg := &loggerConfig{
		configProvider: func() (*Config, error) {
			return config.ProcessAndValidate[Config]()
		},
		outputProvider: func() (io.Writer, error) {
			return os.Stdout, nil
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	envConf, err := cfg.configProvider()
	if err != nil {
		panic(fmt.Sprintf("Failed to get logger config (%s).", err.Error()))
	}

	level, err := ParseLevel(envConf.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse the log level (%s).", err.Error()))
	}
	SetLevel(level)

	output, err := cfg.outputProvider()
	if err != nil {
		panic(fmt.Sprintf("Failed to get logger output (%s).", err.Error()))
	}
	SetOutput(output)
}
