package span

// spanOptions is configured by the Option type.
type spanOptions struct {
	endCallback func(*Span)
}

// Option configures a spanOptions instance.
type Option func(opts *spanOptions)

// configure applies options to the default spanOptions values.
func configure(opts ...Option) *spanOptions {
	spanOpts := &spanOptions{
		endCallback: nil,
	}
	for _, opt := range opts {
		opt(spanOpts)
	}
	return spanOpts
}

// WithEndCallback provides an Option to set a callback that is invoked when the span ends.
func WithEndCallback(callback func(*Span)) Option {
	return func(opts *spanOptions) {
		opts.endCallback = callback
	}
}
