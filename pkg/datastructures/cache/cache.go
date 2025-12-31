package cache

import (
	"sync"
)

// GetOrSetFn is used in the GetOrSet function of the Cache interface.
// If the value is not present, the function gets called.
type GetOrSetFn[Key comparable, Value any] func(Key) (Value, error)

// getOrSetKeyLock is used by the GetOrSet function to make sure the function is not executed in parallel.
type getOrSetKeyLock[Value any] struct {
	WaitChan chan struct{}
	FnValue  Value
	FnError  error
}

// Cache stores key/value pairs.
type Cache[Key comparable, Value any] struct {
	getOrSetKeyLocks sync.Map
	keyToItem        sync.Map
}

// New creates a new Cache instance. The benefit of using Cache instead of a regular map is that
// Cache is thread safe.
func New[Key comparable, Value any]() *Cache[Key, Value] {
	return &Cache[Key, Value]{}
}

// Set sets a value in the cache.
func (c *Cache[Key, Value]) Set(key Key, value Value) {
	c.keyToItem.Store(key, value)
}

// Get retrieves a value from the cache if present.
func (c *Cache[Key, Value]) Get(key Key) (Value, bool) {
	valueUncast, loaded := c.keyToItem.Load(key)
	if !loaded {
		var zeroValue Value
		return zeroValue, false
	}
	return valueUncast.(Value), true
}

// GetOrSet tries to get the value, and if not present, it calls getOrSetFn to fetch and set the value.
func (c *Cache[Key, Value]) GetOrSet(key Key, getOrSetFn GetOrSetFn[Key, Value]) (Value, error) {
	keyLockUncast, keyLockFound := c.getOrSetKeyLocks.LoadOrStore(key, &getOrSetKeyLock[Value]{
		WaitChan: make(chan struct{}),
	})
	keyLock := keyLockUncast.(*getOrSetKeyLock[Value])

	if keyLockFound {
		// In this case, there is a concurrent call with the same key.
		// Wait for the concurrent call to complete and return its fetched value.
		<-keyLock.WaitChan
		return keyLock.FnValue, keyLock.FnError
	}

	// In this case, there is no concurrent call to GetOrSet with this key.
	// If a concurrent call happens before the end of this function, it will receive the value set in keyLock.
	defer func() {
		c.getOrSetKeyLocks.Delete(key)
		close(keyLock.WaitChan)
	}()

	var valueFound bool
	keyLock.FnValue, valueFound = c.Get(key)
	if valueFound {
		return keyLock.FnValue, nil
	}

	keyLock.FnValue, keyLock.FnError = getOrSetFn(key)
	if keyLock.FnError != nil {
		return keyLock.FnValue, keyLock.FnError
	}

	c.Set(key, keyLock.FnValue)
	return keyLock.FnValue, nil
}

// Remove removes a key and its value from the cache if present.
func (c *Cache[Key, Value]) Remove(key Key) (Value, bool) {
	valueUncast, loaded := c.keyToItem.LoadAndDelete(key)
	if !loaded {
		var zeroValue Value
		return zeroValue, false
	}
	return valueUncast.(Value), true
}

// Clear removes all the values in the cache.
func (c *Cache[Key, Value]) Clear() {
	c.keyToItem.Clear()
}
