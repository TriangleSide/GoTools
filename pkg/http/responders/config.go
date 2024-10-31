package responders

// config holds all the configurations for the responders.
type config struct {
	errorCallback func(error)
}

// Option configures the responders.
type Option func(*config)

// WithErrorCallback configures the responder to invoke this callback when there's a write error.
func WithErrorCallback(callback func(error)) Option {
	return func(cfg *config) {
		cfg.errorCallback = callback
	}
}

// configure creates a config out of the provided options.
func configure(opts ...Option) *config {
	cfg := &config{
		errorCallback: func(error) {},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
