package config

// LoggerConf contains the values needed to configure the logger.
type LoggerConf struct {
	LogLevel string `split_words:"true" default:"INFO" validate:"required,oneof=ERROR WARN INFO DEBUG TRACE"`
}
