package cache

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoBase/pkg/test/assert"
	"github.com/TriangleSide/GoBase/pkg/utils/ptr"
)

func cacheMustHaveKeyAndValue[Key comparable, Value any](t *testing.T, testCache *Cache[Key, Value], key Key, value Value) {
	t.Helper()
	gottenValue, gotten := testCache.Get(key)
	assert.True(t, gotten)
	assert.Equals(t, value, gottenValue)
}

func TestCache(t *testing.T) {
	t.Parallel()

	t.Run("should be able to reset the cache repeatedly", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		for i := 0; i < 3; i++ {
			testCache.Reset()
		}
		assert.Equals(t, len(testCache.keyToItem), 0)
	})

	t.Run("should be able to remove a key repeatedly", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		testCache := New[string, string]()
		for i := 0; i < 3; i++ {
			testCache.Remove(key)
		}
		assert.Equals(t, len(testCache.keyToItem), 0)
	})

	t.Run("when there is no values in the cache it should should return false when getting a key", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const key = "key"
		_, gotten := testCache.Get(key)
		assert.False(t, gotten)
		assert.Equals(t, len(testCache.keyToItem), 0)
	})

	t.Run("when there is no values in the cache it should call the fn with get or set", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const key = "key"
		const value = "value"
		fnCalled := false
		_, gotten := testCache.Get(key)
		assert.False(t, gotten)
		returnVal, err := testCache.GetOrSet(key, func(key string) (string, *time.Duration, error) {
			fnCalled = true
			return value, nil, nil
		})
		assert.NoError(t, err)
		assert.True(t, fnCalled)
		assert.Equals(t, value, returnVal)
		_, gotten = testCache.Get(key)
		assert.True(t, gotten)
		assert.Equals(t, len(testCache.keyToItem), 1)
	})

	t.Run("when there is no values in the cache it should return an error if it occurs in get or set", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const key = "key"
		const value = "value"
		fnCalled := false
		_, gotten := testCache.Get(key)
		assert.False(t, gotten)
		returnVal, err := testCache.GetOrSet(key, func(key string) (string, *time.Duration, error) {
			fnCalled = true
			return value, nil, errors.New("error")
		})
		assert.ErrorExact(t, err, "error")
		assert.True(t, fnCalled)
		assert.Equals(t, value, returnVal)
		_, gotten = testCache.Get(key)
		assert.False(t, gotten)
		assert.Equals(t, len(testCache.keyToItem), 0)
	})

	t.Run("when an item is cached without an expiry time it should be available to get", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const key = "key"
		const value = "value"
		testCache.Set(key, value, nil)
		cacheMustHaveKeyAndValue(t, testCache, key, value)
		assert.Equals(t, len(testCache.keyToItem), 1)
	})

	t.Run("when an item is cached without an expiry time it should not call the function in get or set since it's not expired", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const key = "key"
		const value = "value"
		testCache.Set(key, value, nil)
		fnCalled := false
		returnVal, err := testCache.GetOrSet(key, func(key string) (string, *time.Duration, error) {
			fnCalled = true
			return "other", nil, nil
		})
		assert.False(t, fnCalled)
		assert.NoError(t, err)
		assert.Equals(t, value, returnVal)
		cacheMustHaveKeyAndValue(t, testCache, key, value)
		assert.Equals(t, len(testCache.keyToItem), 1)
	})

	t.Run("when an item is cached without an expiry time it should be able to be overwritten by set", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const key = "key"
		const value = "value"
		testCache.Set(key, value, nil)
		const newValue = "newValue"
		cacheMustHaveKeyAndValue(t, testCache, key, value)
		testCache.Set(key, newValue, nil)
		cacheMustHaveKeyAndValue(t, testCache, key, newValue)
		assert.Equals(t, len(testCache.keyToItem), 1)
	})

	t.Run("when a cache item expires it should not be available to get", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const key = "key"
		const value = "value"
		testCache.Set(key, value, ptr.Of(time.Nanosecond))
		time.Sleep(time.Nanosecond * 2)
		_, gotten := testCache.Get(key)
		assert.False(t, gotten)
		assert.Equals(t, len(testCache.keyToItem), 0)
	})

	t.Run("when a cache item expires it should call the function in get or set since it's expired", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const key = "key"
		const value = "value"
		const other = "other"
		testCache.Set(key, value, ptr.Of(time.Nanosecond))
		time.Sleep(time.Nanosecond * 2)
		fnCalled := false
		returnVal, err := testCache.GetOrSet(key, func(key string) (string, *time.Duration, error) {
			fnCalled = true
			return other, nil, nil
		})
		assert.True(t, fnCalled)
		assert.NoError(t, err)
		assert.Equals(t, returnVal, other)
		cacheMustHaveKeyAndValue(t, testCache, key, other)
		assert.Equals(t, len(testCache.keyToItem), 1)
	})

	t.Run("it should be able to handle concurrency on unique sequential operations", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const threadCount = 4
		const loopCount = 5000
		wg := sync.WaitGroup{}
		startChan := make(chan struct{})
		for i := 0; i < threadCount; i++ {
			wg.Add(1)
			go func() {
				<-startChan
				for k := 0; k < loopCount; k++ {
					key := fmt.Sprintf("key-%d-%d", i, k)
					value := fmt.Sprintf("value-%d-%d", i, k)
					testCache.Set(key, value, ptr.Of(time.Minute))
					gottenValue, gotten := testCache.Get(key)
					assert.True(t, gotten, assert.Continue())
					assert.Equals(t, value, gottenValue, assert.Continue())
					testCache.Remove(key)
					testCache.Set(key, value, ptr.Of(time.Nanosecond))
					time.Sleep(time.Nanosecond * 2)
					_, gotten = testCache.Get(key)
					assert.False(t, gotten, assert.Continue())
					other := fmt.Sprintf("other-%d-%d", i, k)
					gottenValue, err := testCache.GetOrSet(key, func(key string) (string, *time.Duration, error) {
						return other, ptr.Of(time.Minute), nil
					})
					assert.NoError(t, err, assert.Continue())
					assert.Equals(t, gottenValue, other, assert.Continue())
					testCache.Remove(key)
				}
				wg.Done()
			}()
		}
		close(startChan)
		wg.Wait()
		assert.Equals(t, len(testCache.getOrSetKeyLocks), 0)
	})

	t.Run("it should be able to handle concurrency with get and set on the same key", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const threadCount = 4
		const loopCount = 5000
		wg := sync.WaitGroup{}
		startChan := make(chan struct{})
		for i := 0; i < threadCount; i++ {
			wg.Add(1)
			go func() {
				<-startChan
				for k := 0; k < loopCount; k++ {
					const key = "key"
					const value = "value"
					testCache.Set(key, value, ptr.Of(time.Millisecond))
					gottenValue, gotten := testCache.Get(key)
					if gotten {
						assert.Equals(t, gottenValue, value, assert.Continue())
					}
					testCache.Remove(key)
					gottenValue, err := testCache.GetOrSet(key, func(key string) (string, *time.Duration, error) {
						return value, ptr.Of(time.Millisecond), nil
					})
					assert.NoError(t, err, assert.Continue())
					assert.Equals(t, gottenValue, value, assert.Continue())
				}
				wg.Done()
			}()
		}
		close(startChan)
		wg.Wait()
		assert.Equals(t, len(testCache.getOrSetKeyLocks), 0)
	})

	t.Run("it should be able to handle concurrency with get or set on the same key", func(t *testing.T) {
		t.Parallel()
		testCache := New[string, string]()
		const threadCount = 4
		const loopCount = 5000
		wg := sync.WaitGroup{}
		startChan := make(chan struct{})
		for i := 0; i < threadCount; i++ {
			wg.Add(1)
			go func() {
				<-startChan
				for k := 0; k < loopCount; k++ {
					const key = "key"
					const value = "value"
					gottenValue, err := testCache.GetOrSet(key, func(key string) (string, *time.Duration, error) {
						return value, ptr.Of(time.Nanosecond), nil
					})
					assert.NoError(t, err, assert.Continue())
					assert.Equals(t, gottenValue, value, assert.Continue())
				}
				wg.Done()
			}()
		}
		close(startChan)
		wg.Wait()
		assert.Equals(t, len(testCache.getOrSetKeyLocks), 0)
	})
}
