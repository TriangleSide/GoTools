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

// Do executes the callback once per test or subtest.
func Do(t *testing.T, callback func()) {
	t.Helper()

	_, file, lineNumber, _ := runtime.Caller(1)
	uniqueID := fmt.Sprintf("%s-%d-%s", file, lineNumber, t.Name())

	once, _ := registeredSingleInvocations.LoadOrStore(uniqueID, &sync.Once{})
	once.(*sync.Once).Do(callback)
}
