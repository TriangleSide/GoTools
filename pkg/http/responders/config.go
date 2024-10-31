package responders

// config holds all the configurations for the responders.
type config struct {
	writeErrorCallback func(error)
}

// Option configures the responders.
type Option func(*config)

// WithWriteErrorCallback configures the responder to invoke this callback when there's a write error.
func WithWriteErrorCallback(callback func(error)) Option {
	return func(cfg *config) {
		cfg.writeErrorCallback = callback
	}
}

// configure creates a config out of the provided options.
func configure(opts ...Option) *config {
	cfg := &config{
		writeErrorCallback: func(error) {},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
