package assert

// Testing matches the functions on the testing.T struct.
type Testing interface {
	Helper()
	Error(...any)
	Fatal(...any)
}

// testContext implements the Testing interface. It can be configured using options.
type testContext struct {
	Testing
	failFunc func(args ...any)
}

// Option modifies the configuration of testing.
type Option func(t *testContext)

// Continue marks the test as having failed but continues execution.
func Continue() Option {
	return func(t *testContext) {
		t.failFunc = t.Testing.Error
	}
}

// newTestContext creates a new testing struct with optional settings.
func newTestContext(t Testing, options ...Option) *testContext {
	newT := &testContext{
		Testing:  t,
		failFunc: t.Fatal,
	}
	for _, opt := range options {
		opt(newT)
	}
	return newT
}

// fail calls either testing.T's Error or Fatal function.
func (t *testContext) fail(args ...any) {
	t.Helper()
	t.failFunc(args...)
}
