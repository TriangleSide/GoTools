package cache_test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/datastructures/cache"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

const otherValue = "other"

func cacheMustHaveKeyAndValue[Key comparable, Value any](
	t *testing.T, testCache *cache.Cache[Key, Value], key Key, value Value,
) {
	t.Helper()
	gottenValue, gotten := testCache.Get(key)
	assert.True(t, gotten)
	assert.Equals(t, value, gottenValue)
}

func getRandomInt(t *testing.T, maxValue int) int {
	t.Helper()
	randomValueBig, err := rand.Int(rand.Reader, big.NewInt(int64(maxValue)))
	assert.Nil(t, err)
	return int(randomValueBig.Int64())
}

func TestClear_CalledRepeatedly_Succeeds(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	for range 3 {
		testCache.Clear()
	}
}

func TestRemove_KeyNotPresent_ReturnsFalse(t *testing.T) {
	t.Parallel()
	const key = "key"
	testCache := cache.New[string, string]()
	for range 3 {
		_, found := testCache.Remove(key)
		assert.False(t, found)
	}
}

func TestGet_EmptyCache_ReturnsFalse(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestRemove_KeyNotInCache_ReturnsFalseAndValueRemains(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value)
	removedValue, removed := testCache.Remove("otherKey")
	assert.False(t, removed)
	assert.Equals(t, removedValue, "")
	cacheMustHaveKeyAndValue(t, testCache, key, value)
}

func TestRemove_KeyInCache_RemovesAndReturnsTrue(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value)
	removedValue, removed := testCache.Remove(key)
	assert.True(t, removed)
	assert.Equals(t, removedValue, value)
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestGet_KeyExists_ReturnsValue(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value)
	gottenValue, gotten := testCache.Get(key)
	assert.True(t, gotten)
	assert.Equals(t, gottenValue, value)
}

func TestGetOrSet_EmptyCache_CallsFunctionAndSetsValue(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	fnCalled := false
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
	returnVal, err := testCache.GetOrSet(key, func(string) (string, error) {
		fnCalled = true
		return value, nil
	})
	assert.NoError(t, err)
	assert.True(t, fnCalled)
	assert.Equals(t, value, returnVal)
	_, gotten = testCache.Get(key)
	assert.True(t, gotten)
}

func TestGetOrSet_FunctionReturnsError_ReturnsError(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	fnCalled := false
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
	returnVal, err := testCache.GetOrSet(key, func(string) (string, error) {
		fnCalled = true
		return value, errors.New("error")
	})
	assert.ErrorExact(t, err, "error")
	assert.True(t, fnCalled)
	assert.Equals(t, value, returnVal)
	_, gotten = testCache.Get(key)
	assert.False(t, gotten)
}

func TestSet_ValueIsAvailable(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value)
	cacheMustHaveKeyAndValue(t, testCache, key, value)
}

func TestGetOrSet_KeyExists_DoesNotCallFunction(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value)
	fnCalled := false
	returnVal, err := testCache.GetOrSet(key, func(string) (string, error) {
		fnCalled = true
		return otherValue, nil
	})
	assert.False(t, fnCalled)
	assert.NoError(t, err)
	assert.Equals(t, value, returnVal)
	cacheMustHaveKeyAndValue(t, testCache, key, value)
}

func TestSet_CanBeOverwritten(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value)
	const newValue = "newValue"
	cacheMustHaveKeyAndValue(t, testCache, key, value)
	testCache.Set(key, newValue)
	cacheMustHaveKeyAndValue(t, testCache, key, newValue)
}

func TestCache_Concurrency_UniqueSequentialOperations(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const threadCount = 4
	const loopCount = 10000
	var waitGroup sync.WaitGroup
	startChan := make(chan struct{})
	for threadIdx := range threadCount {
		waitGroup.Go(func() {
			<-startChan
			for iteration := range loopCount {
				key := fmt.Sprintf("key-%d-%d", threadIdx, iteration)
				value := fmt.Sprintf("value-%d-%d", threadIdx, iteration)
				testCache.Set(key, value)
				gottenValue, gotten := testCache.Get(key)
				assert.True(t, gotten, assert.Continue())
				assert.Equals(t, value, gottenValue, assert.Continue())
				testCache.Remove(key)
				testCache.Set(key, value)
				other := fmt.Sprintf("other-%d-%d", threadIdx, iteration)
				gottenValue, err := testCache.GetOrSet(key, func(string) (string, error) {
					return other, nil
				})
				assert.NoError(t, err, assert.Continue())
				assert.Equals(t, gottenValue, value, assert.Continue())
				testCache.Remove(key)
			}
		})
	}
	close(startChan)
	waitGroup.Wait()
}

func TestCache_Concurrency_GetSetRemove(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const threadCount = 4
	const loopCount = 10000
	var waitGroup sync.WaitGroup
	startChan := make(chan struct{})
	for range threadCount {
		waitGroup.Go(func() {
			<-startChan
			for range loopCount {
				const key = "key"
				const value = "value"
				testCache.Set(key, value)
				gottenValue, gotten := testCache.Get(key)
				if gotten {
					assert.Equals(t, gottenValue, value, assert.Continue())
				}
				testCache.Remove(key)
			}
		})
	}
	close(startChan)
	waitGroup.Wait()
}

func TestCache_Concurrency_GetOrSet(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const threadCount = 4
	const loopCount = 10000
	var waitGroup sync.WaitGroup
	startChan := make(chan struct{})
	for range threadCount {
		waitGroup.Go(func() {
			<-startChan
			for range loopCount {
				key := "key" + strconv.Itoa(getRandomInt(t, threadCount))
				value := "value" + strconv.Itoa(getRandomInt(t, threadCount))
				_, err := testCache.GetOrSet(key, func(string) (string, error) {
					return value, nil
				})
				assert.NoError(t, err, assert.Continue())
			}
		})
	}
	close(startChan)
	waitGroup.Wait()
}

func TestGetOrSet_ConcurrentCalls_ReturnsFirstCallersValue(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	var waitGroup sync.WaitGroup
	const key = "key"
	const threadCount = 4

	firstWaitChan := make(chan struct{})

	for range threadCount {
		waitGroup.Go(func() {
			<-firstWaitChan
			returnedValue, err := testCache.GetOrSet(key, func(string) (string, error) {
				return otherValue, nil
			})
			assert.NoError(t, err, assert.Continue())
			assert.Equals(t, returnedValue, "first", assert.Continue())
		})
	}

	waitGroup.Go(func() {
		returnedValue, err := testCache.GetOrSet(key, func(string) (string, error) {
			close(firstWaitChan)
			return "first", nil
		})
		assert.NoError(t, err, assert.Continue())
		assert.Equals(t, returnedValue, "first", assert.Continue())
	})

	waitGroup.Wait()
}

func TestNew_CreatesEmptyCache(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	_, gotten := testCache.Get("anyKey")
	assert.False(t, gotten)
}

func TestGet_AfterClear_ReturnsFalse(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value)
	cacheMustHaveKeyAndValue(t, testCache, key, value)
	testCache.Clear()
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestClear_WithMultipleItems_ClearsAll(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	testCache.Set("key1", "value1")
	testCache.Set("key2", "value2")
	testCache.Set("key3", "value3")
	testCache.Clear()
	_, gotten1 := testCache.Get("key1")
	_, gotten2 := testCache.Get("key2")
	_, gotten3 := testCache.Get("key3")
	assert.False(t, gotten1)
	assert.False(t, gotten2)
	assert.False(t, gotten3)
}

func TestRemove_AfterClear_ReturnsFalse(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value)
	testCache.Clear()
	_, removed := testCache.Remove(key)
	assert.False(t, removed)
}

func TestGetOrSet_AfterRemove_CallsFunction(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	testCache.Set(key, "value1")
	testCache.Remove(key)
	fnCalled := false
	returnVal, err := testCache.GetOrSet(key, func(string) (string, error) {
		fnCalled = true
		return "value2", nil
	})
	assert.True(t, fnCalled)
	assert.NoError(t, err)
	assert.Equals(t, returnVal, "value2")
}

func TestCache_IntegerKeys_Works(t *testing.T) {
	t.Parallel()
	testCache := cache.New[int, string]()
	testCache.Set(1, "one")
	testCache.Set(2, "two")
	gottenValue1, gotten1 := testCache.Get(1)
	gottenValue2, gotten2 := testCache.Get(2)
	assert.True(t, gotten1)
	assert.Equals(t, gottenValue1, "one")
	assert.True(t, gotten2)
	assert.Equals(t, gottenValue2, "two")
}

func TestCache_StructValues_Works(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field1 string
		Field2 int
	}
	testCache := cache.New[string, testStruct]()
	expected := testStruct{Field1: "test", Field2: 42}
	testCache.Set("key", expected)
	gottenValue, gotten := testCache.Get("key")
	assert.True(t, gotten)
	assert.Equals(t, gottenValue, expected)
}

func TestCache_PointerValues_Works(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, *string]()
	value := "test"
	testCache.Set("key", &value)
	gottenValue, gotten := testCache.Get("key")
	assert.True(t, gotten)
	assert.Equals(t, *gottenValue, value)
}

func TestGetOrSet_FunctionPanics_Panics(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	panicValue := errors.New("panic error")
	assert.PanicExact(t, func() {
		_, _ = testCache.GetOrSet(key, func(string) (string, error) {
			panic(panicValue)
		})
	}, panicValue.Error())
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestGetOrSet_ConcurrentCallsWithPanic_AllCallersPanic(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const threadCount = 2
	panicValue := errors.New("panic error")
	var waitGroup sync.WaitGroup
	var panicCount atomic.Int64
	waitersReadyChan := make(chan struct{})
	proceedToPanicChan := make(chan struct{})
	var waitersStarted sync.WaitGroup
	waitersStarted.Add(threadCount)

	for range threadCount {
		waitGroup.Go(func() {
			defer func() {
				recovered := recover()
				if recovered != nil {
					panicCount.Add(1)
					assert.Equals(t, recovered, panicValue, assert.Continue())
				}
			}()
			<-waitersReadyChan
			waitersStarted.Done()
			_, _ = testCache.GetOrSet(key, func(string) (string, error) {
				return otherValue, nil
			})
		})
	}

	waitGroup.Go(func() {
		defer func() {
			recovered := recover()
			if recovered != nil {
				panicCount.Add(1)
				assert.Equals(t, recovered, panicValue, assert.Continue())
			}
		}()
		_, _ = testCache.GetOrSet(key, func(string) (string, error) {
			close(waitersReadyChan)
			waitersStarted.Wait()
			<-proceedToPanicChan
			time.Sleep(time.Millisecond * 250)
			panic(panicValue)
		})
	})

	close(proceedToPanicChan)
	waitGroup.Wait()

	assert.Equals(t, panicCount.Load(), int64(threadCount+1))
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}
