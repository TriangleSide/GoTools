package responders

import (
	"encoding/json"
)

// config holds all the configurations for the responders.
type config struct {
	errorCallback func(error)
	jsonMarshal   func(any) ([]byte, error)
}

// Option configures the responders.
type Option func(*config)

// WithJSONMarshal configures a JSON marshal function for testing.
func WithJSONMarshal(marshal func(any) ([]byte, error)) Option {
	return func(cfg *config) {
		cfg.jsonMarshal = marshal
	}
}

// WithErrorCallback configures the responder to invoke this callback when a responder processing error occurs.
// Do not retry the responder when this is invoked.
func WithErrorCallback(callback func(error)) Option {
	return func(cfg *config) {
		cfg.errorCallback = callback
	}
}

// configure creates a config out of the provided options.
func configure(opts ...Option) *config {
	cfg := &config{
		errorCallback: func(error) {},
		jsonMarshal:   json.Marshal,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
