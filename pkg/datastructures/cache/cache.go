package cache

import (
	"sync"
	"time"
)

// GetOrSetFn is used in the GetOrSet function of the Cache interface.
// If the value is not present, or if it's expired, the function gets called.
type GetOrSetFn[Key comparable, Value any] func(Key) (Value, *time.Duration, error)

// getOrSetKeyLock is used by the GetOrSet function to make sure the function is not executed in parallel.
type getOrSetKeyLock[Value any] struct {
	WaitChan chan struct{}
	FnValue  Value
	FnError  error
}

// Cache stores key/value pairs with optional expiration.
type Cache[Key comparable, Value any] struct {
	getOrSetKeyLocks sync.Map
	keyToItem        sync.Map
}

// New creates a new Cache instance. The benefit of using Cache instead of a regular map is that
// Cache is thread safe. It also handles expiring items.
func New[Key comparable, Value any]() *Cache[Key, Value] {
	return &Cache[Key, Value]{
		getOrSetKeyLocks: sync.Map{},
		keyToItem:        sync.Map{},
	}
}

// item is what is stored in the internal map of the cache.
type item[Value any] struct {
	value  Value
	expiry *time.Time
}

// Set sets a value and time-to-live in the cache.
func (c *Cache[Key, Value]) Set(key Key, value Value, ttl *time.Duration) {
	var itemToAdd *item[Value]
	if ttl != nil {
		expireTime := time.Now().Add(*ttl)
		itemToAdd = &item[Value]{
			value:  value,
			expiry: &expireTime,
		}
	} else {
		itemToAdd = &item[Value]{
			value:  value,
			expiry: nil,
		}
	}
	c.keyToItem.Store(key, itemToAdd)
}

// Get retrieves a value from the cache if present.
func (c *Cache[Key, Value]) Get(key Key) (Value, bool) {
	itemValueUncast, loaded := c.keyToItem.Load(key)
	if !loaded {
		var zeroValue Value
		return zeroValue, false
	}
	itemValue := itemValueUncast.(*item[Value])
	if itemValue.expiry != nil && time.Now().After(*itemValue.expiry) {
		c.keyToItem.CompareAndDelete(key, itemValue)
		var zeroValue Value
		return zeroValue, false
	}
	return itemValue.value, true
}

// GetOrSet tries to get the value, and if not present, it calls fn to fetch and set the value.
func (c *Cache[Key, Value]) GetOrSet(key Key, fn GetOrSetFn[Key, Value]) (Value, error) {
	keyLockUncast, keyLockFound := c.getOrSetKeyLocks.LoadOrStore(key, &getOrSetKeyLock[Value]{
		WaitChan: make(chan struct{}),
	})
	keyLock := keyLockUncast.(*getOrSetKeyLock[Value])

	if keyLockFound {
		// In this case, there is a concurrent call with the same key.
		// Wait for the concurrent call to complete and return its fetched value.
		<-keyLock.WaitChan
		return keyLock.FnValue, keyLock.FnError
	} else {
		// In this case, there is no concurrent call to GetOrSet with this key.
		// If a concurrent call happens before the end of this function, it will receive the value set in keyLock.
		defer func() {
			c.getOrSetKeyLocks.Delete(key)
			close(keyLock.WaitChan)
		}()
	}

	var valueFound bool
	keyLock.FnValue, valueFound = c.Get(key)
	if valueFound {
		return keyLock.FnValue, nil
	}

	var ttl *time.Duration
	keyLock.FnValue, ttl, keyLock.FnError = fn(key)
	if keyLock.FnError != nil {
		return keyLock.FnValue, keyLock.FnError
	}

	c.Set(key, keyLock.FnValue, ttl)
	return keyLock.FnValue, nil
}

// Remove removes a key and its value from the cache if present.
func (c *Cache[Key, Value]) Remove(key Key) {
	c.keyToItem.Delete(key)
}

// Clear removes all the values in the cache.
func (c *Cache[Key, Value]) Clear() {
	c.keyToItem.Clear()
}
