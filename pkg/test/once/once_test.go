package once_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/TriangleSide/go-toolkit/pkg/test/assert"
	"github.com/TriangleSide/go-toolkit/pkg/test/once"
)

var (
	testDoSingleCallOnce        sync.Once
	testDoMultipleCallsOnce     sync.Once
	testDoConcurrentCallsOnce   sync.Once
	testDoMultipleCallSitesOnce sync.Once
	testDoDifferentSubtestsOnce sync.Once
	testDoCallbackPanicsOnce    sync.Once
)

func TestDo_SingleCall_InvokesCallback(t *testing.T) {
	t.Parallel()
	testDoSingleCallOnce.Do(func() {
		var callCount atomic.Int32
		once.Do(t, func() {
			callCount.Add(1)
		})
		assert.Equals(t, int32(1), callCount.Load())
	})
}

func TestDo_MultipleCalls_InvokesCallbackOnlyOnce(t *testing.T) {
	t.Parallel()
	testDoMultipleCallsOnce.Do(func() {
		var callCount atomic.Int32
		for range 10 {
			once.Do(t, func() { callCount.Add(1) })
		}
		assert.Equals(t, int32(1), callCount.Load())
	})
}

func TestDo_ConcurrentCalls_InvokesCallbackOnlyOnce(t *testing.T) {
	t.Parallel()
	testDoConcurrentCallsOnce.Do(func() {
		const goroutines = 100
		var callCount atomic.Int32
		var waitGroup sync.WaitGroup
		for range goroutines {
			waitGroup.Go(func() {
				once.Do(t, func() { callCount.Add(1) })
			})
		}
		waitGroup.Wait()
		assert.Equals(t, int32(1), callCount.Load())
	})
}

func TestDo_MultipleCallSites_InvokesEachCallbackOnce(t *testing.T) {
	t.Parallel()
	testDoMultipleCallSitesOnce.Do(func() {
		var callCount1 atomic.Int32
		var callCount2 atomic.Int32
		var callCount3 atomic.Int32
		once.Do(t, func() { callCount1.Add(1) })
		once.Do(t, func() { callCount2.Add(1) })
		once.Do(t, func() { callCount3.Add(1) })
		assert.Equals(t, int32(1), callCount1.Load())
		assert.Equals(t, int32(1), callCount2.Load())
		assert.Equals(t, int32(1), callCount3.Load())
	})
}

func TestDo_DifferentSubtests_InvokesCallbackForEach(t *testing.T) {
	t.Parallel()
	testDoDifferentSubtestsOnce.Do(func() {
		subtests := []struct {
			name      string
			callCount *atomic.Int32
		}{
			{name: "subtest1", callCount: &atomic.Int32{}},
			{name: "subtest2", callCount: &atomic.Int32{}},
			{name: "subtest3", callCount: &atomic.Int32{}},
		}
		for _, tc := range subtests {
			t.Run(tc.name, func(t *testing.T) {
				once.Do(t, func() { tc.callCount.Add(1) })
				assert.Equals(t, int32(1), tc.callCount.Load())
			})
		}
	})
}

func TestDo_CallbackPanics_PropagatesPanic(t *testing.T) {
	t.Parallel()
	testDoCallbackPanicsOnce.Do(func() {
		assert.PanicExact(t, func() {
			once.Do(t, func() { panic(errors.New("callback panic")) })
		}, "callback panic")
	})
}
