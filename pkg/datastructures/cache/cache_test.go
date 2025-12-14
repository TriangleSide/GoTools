package cache_test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/datastructures/cache"
	"github.com/TriangleSide/GoTools/pkg/ptr"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func cacheMustHaveKeyAndValue[Key comparable, Value any](
	t *testing.T, testCache *cache.Cache[Key, Value], key Key, value Value) {
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
	testCache.Set(key, value, ptr.Of(time.Minute))
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
	testCache.Set(key, value, ptr.Of(time.Minute))
	removedValue, removed := testCache.Remove(key)
	assert.True(t, removed)
	assert.Equals(t, removedValue, value)
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestRemove_ExpiredItem_ReturnsFalse(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, ptr.Of(time.Nanosecond))
	time.Sleep(time.Nanosecond * 2)
	removedValue, removed := testCache.Remove(key)
	assert.False(t, removed)
	assert.Equals(t, removedValue, "")
}

func TestGet_NotExpired_ReturnsValue(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, ptr.Of(time.Minute))
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
	returnVal, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
		fnCalled = true
		return value, nil, nil
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
	returnVal, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
		fnCalled = true
		return value, nil, errors.New("error")
	})
	assert.ErrorExact(t, err, "error")
	assert.True(t, fnCalled)
	assert.Equals(t, value, returnVal)
	_, gotten = testCache.Get(key)
	assert.False(t, gotten)
}

func TestSet_NoExpiry_ValueIsAvailable(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, nil)
	cacheMustHaveKeyAndValue(t, testCache, key, value)
}

func TestGetOrSet_NoExpiry_DoesNotCallFunction(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, nil)
	fnCalled := false
	returnVal, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
		fnCalled = true
		return "other", nil, nil
	})
	assert.False(t, fnCalled)
	assert.NoError(t, err)
	assert.Equals(t, value, returnVal)
	cacheMustHaveKeyAndValue(t, testCache, key, value)
}

func TestSet_NoExpiry_CanBeOverwritten(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, nil)
	const newValue = "newValue"
	cacheMustHaveKeyAndValue(t, testCache, key, value)
	testCache.Set(key, newValue, nil)
	cacheMustHaveKeyAndValue(t, testCache, key, newValue)
}

func TestGet_Expired_ReturnsFalse(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, ptr.Of(time.Nanosecond))
	time.Sleep(time.Nanosecond * 2)
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestGetOrSet_Expired_CallsFunctionAndUpdatesValue(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	const other = "other"
	testCache.Set(key, value, ptr.Of(time.Nanosecond))
	time.Sleep(time.Nanosecond * 2)
	fnCalled := false
	returnVal, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
		fnCalled = true
		return other, nil, nil
	})
	assert.True(t, fnCalled)
	assert.NoError(t, err)
	assert.Equals(t, returnVal, other)
	cacheMustHaveKeyAndValue(t, testCache, key, other)
}

func TestCache_Concurrency_UniqueSequentialOperations(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const threadCount = 4
	const loopCount = 10000
	waitGroup := sync.WaitGroup{}
	startChan := make(chan struct{})
	for threadIdx := range threadCount {
		waitGroup.Go(func() {
			<-startChan
			for iteration := range loopCount {
				key := fmt.Sprintf("key-%d-%d", threadIdx, iteration)
				value := fmt.Sprintf("value-%d-%d", threadIdx, iteration)
				testCache.Set(key, value, ptr.Of(time.Minute))
				gottenValue, gotten := testCache.Get(key)
				assert.True(t, gotten, assert.Continue())
				assert.Equals(t, value, gottenValue, assert.Continue())
				testCache.Remove(key)
				testCache.Set(key, value, ptr.Of(time.Nanosecond))
				time.Sleep(time.Nanosecond * 2)
				_, gotten = testCache.Get(key)
				assert.False(t, gotten, assert.Continue())
				other := fmt.Sprintf("other-%d-%d", threadIdx, iteration)
				gottenValue, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
					return other, ptr.Of(time.Minute), nil
				})
				assert.NoError(t, err, assert.Continue())
				assert.Equals(t, gottenValue, other, assert.Continue())
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
	waitGroup := sync.WaitGroup{}
	startChan := make(chan struct{})
	for range threadCount {
		waitGroup.Go(func() {
			<-startChan
			for range loopCount {
				const key = "key"
				const value = "value"
				testCache.Set(key, value, ptr.Of(time.Millisecond))
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
	waitGroup := sync.WaitGroup{}
	startChan := make(chan struct{})
	for range threadCount {
		waitGroup.Go(func() {
			<-startChan
			for range loopCount {
				key := "key" + strconv.Itoa(getRandomInt(t, threadCount))
				value := "value" + strconv.Itoa(getRandomInt(t, threadCount))
				_, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
					return value, ptr.Of(time.Millisecond), nil
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
	waitGroup := sync.WaitGroup{}
	const key = "key"
	const threadCount = 4

	firstWaitChan := make(chan struct{})

	for range threadCount {
		waitGroup.Go(func() {
			<-firstWaitChan
			returnedValue, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
				return "other", ptr.Of(time.Hour), nil
			})
			assert.NoError(t, err, assert.Continue())
			assert.Equals(t, returnedValue, "first", assert.Continue())
		})
	}

	waitGroup.Go(func() {
		returnedValue, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
			close(firstWaitChan)
			time.Sleep(time.Second)
			return "first", ptr.Of(time.Nanosecond), nil
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

func TestSet_OverwriteExpiryWithNoExpiry_ValuePersists(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	testCache.Set(key, "value1", ptr.Of(time.Nanosecond))
	testCache.Set(key, "value2", nil)
	time.Sleep(time.Nanosecond * 2)
	gottenValue, gotten := testCache.Get(key)
	assert.True(t, gotten)
	assert.Equals(t, gottenValue, "value2")
}

func TestSet_OverwriteNoExpiryWithExpiry_ValueExpires(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	testCache.Set(key, "value1", nil)
	testCache.Set(key, "value2", ptr.Of(time.Nanosecond))
	time.Sleep(time.Nanosecond * 2)
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestGet_AfterClear_ReturnsFalse(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, nil)
	cacheMustHaveKeyAndValue(t, testCache, key, value)
	testCache.Clear()
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestClear_WithMultipleItems_ClearsAll(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	testCache.Set("key1", "value1", nil)
	testCache.Set("key2", "value2", ptr.Of(time.Minute))
	testCache.Set("key3", "value3", nil)
	testCache.Clear()
	_, gotten1 := testCache.Get("key1")
	_, gotten2 := testCache.Get("key2")
	_, gotten3 := testCache.Get("key3")
	assert.False(t, gotten1)
	assert.False(t, gotten2)
	assert.False(t, gotten3)
}

func TestGetOrSet_WithTTL_ValueExpiresAfterTTL(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	returnVal, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
		return value, ptr.Of(time.Nanosecond), nil
	})
	assert.NoError(t, err)
	assert.Equals(t, returnVal, value)
	time.Sleep(time.Nanosecond * 2)
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestSet_ZeroTTL_ExpiresImmediately(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, ptr.Of(time.Duration(0)))
	_, gotten := testCache.Get(key)
	assert.False(t, gotten)
}

func TestCache_MultipleKeys_IndependentExpiry(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	testCache.Set("key1", "value1", ptr.Of(time.Nanosecond))
	testCache.Set("key2", "value2", ptr.Of(time.Hour))
	time.Sleep(time.Nanosecond * 2)
	_, gotten1 := testCache.Get("key1")
	assert.False(t, gotten1)
	gottenValue2, gotten2 := testCache.Get("key2")
	assert.True(t, gotten2)
	assert.Equals(t, gottenValue2, "value2")
}

func TestRemove_AfterClear_ReturnsFalse(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	const value = "value"
	testCache.Set(key, value, nil)
	testCache.Clear()
	_, removed := testCache.Remove(key)
	assert.False(t, removed)
}

func TestGetOrSet_AfterRemove_CallsFunction(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, string]()
	const key = "key"
	testCache.Set(key, "value1", nil)
	testCache.Remove(key)
	fnCalled := false
	returnVal, err := testCache.GetOrSet(key, func(string) (string, *time.Duration, error) {
		fnCalled = true
		return "value2", nil, nil
	})
	assert.True(t, fnCalled)
	assert.NoError(t, err)
	assert.Equals(t, returnVal, "value2")
}

func TestCache_IntegerKeys_Works(t *testing.T) {
	t.Parallel()
	testCache := cache.New[int, string]()
	testCache.Set(1, "one", nil)
	testCache.Set(2, "two", ptr.Of(time.Minute))
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
	testCache.Set("key", expected, nil)
	gottenValue, gotten := testCache.Get("key")
	assert.True(t, gotten)
	assert.Equals(t, gottenValue, expected)
}

func TestCache_PointerValues_Works(t *testing.T) {
	t.Parallel()
	testCache := cache.New[string, *string]()
	value := "test"
	testCache.Set("key", &value, nil)
	gottenValue, gotten := testCache.Get("key")
	assert.True(t, gotten)
	assert.Equals(t, *gottenValue, value)
}
