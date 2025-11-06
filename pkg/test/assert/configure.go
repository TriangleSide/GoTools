package assert

import "sync"

var (
	// testLocks holds a lock per test case.
	testLocks = sync.Map{}
)

// Testing matches the functions on the testing.T struct.
type Testing interface {
	Name() string
	Helper()
	Error(...any)
	Fatal(...any)
}

// testContext implements the Testing interface. It can be configured using options.
type testContext struct {
	Testing
	mu       *sync.Mutex
	failFunc func(args ...any)
}

// Option modifies the configuration of testing.
type Option func(t *testContext)

// Continue marks the test as having failed but continues execution.
func Continue() Option {
	return func(t *testContext) {
		t.failFunc = t.Error
	}
}

// newTestContext creates a new testing struct with optional settings.
func newTestContext(t Testing, options ...Option) *testContext {
	muNotCast, _ := testLocks.LoadOrStore(t.Name(), &sync.Mutex{})
	mu, _ := muNotCast.(*sync.Mutex)

	newT := &testContext{
		Testing:  t,
		mu:       mu,
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
	t.mu.Lock()
	defer t.mu.Unlock()
	t.failFunc(args...)
}
