package once

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

var (
	registeredSingleInvocations sync.Map
)

// Do executes the callback exactly once per test or subtest.
// Using `go test -count=N` will cause the callback to be executed only once across all N runs.
func Do(t *testing.T, callback func()) {
	t.Helper()

	_, file, lineNumber, _ := runtime.Caller(1)
	uniqueID := fmt.Sprintf("%s-%d-%s", file, lineNumber, t.Name())

	once, _ := registeredSingleInvocations.LoadOrStore(uniqueID, &sync.Once{})
	once.(*sync.Once).Do(callback)
}
