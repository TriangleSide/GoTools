package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/TriangleSide/GoBase/pkg/config"
	"github.com/TriangleSide/GoBase/pkg/config/envprocessor"
)

// loggerConfig is configured by the ConfigOption functions.
type loggerConfig struct {
	configProvider func() (*config.Logger, error)
	outputProvider func() (io.Writer, error)
}

// ConfigOption sets values on the loggerConfig.
type ConfigOption func(*loggerConfig)

// WithConfigProvider sets the provider for the config.Logger.
func WithConfigProvider(provider func() (*config.Logger, error)) ConfigOption {
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

// MustConfigure parses the logger conf and configures the application logger.
func MustConfigure(opts ...ConfigOption) {
	cfg := &loggerConfig{
		configProvider: func() (*config.Logger, error) {
			return envprocessor.ProcessAndValidate[config.Logger]()
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
