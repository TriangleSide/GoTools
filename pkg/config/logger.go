package config

// Logger contains the values needed to configure the logger.
type Logger struct {
	LogLevel string `config_format:"snake" config_default:"INFO" validate:"required,oneof=ERROR WARN INFO DEBUG TRACE"`
}
