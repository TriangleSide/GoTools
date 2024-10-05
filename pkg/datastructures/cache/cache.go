package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// GetOrSetFn is used in the GetOrSet function of the Cache interface.
// If the value is not present, or if it's expired, the function gets called.
type GetOrSetFn[Key comparable, Value any] func(Key) (Value, *time.Duration, error)

// getOrSetKeyLock is used by the GetOrSet function to make sure the function is not executed in parallel.
type getOrSetKeyLock[Value any] struct {
	Count    atomic.Int32
	WaitChan chan struct{}
	FnValue  Value
	FnError  error
}

// Cache is an implementation of the Cache interface.
type Cache[Key comparable, Value any] struct {
	rwMutex          sync.RWMutex
	getOrSetLock     sync.Mutex
	getOrSetKeyLocks map[Key]*getOrSetKeyLock[Value]
	keyToItem        map[Key]*item[Value]
}

// New creates a new instance of the Cache interface.
func New[Key comparable, Value any]() *Cache[Key, Value] {
	return &Cache[Key, Value]{
		rwMutex:          sync.RWMutex{},
		getOrSetLock:     sync.Mutex{},
		getOrSetKeyLocks: make(map[Key]*getOrSetKeyLock[Value]),
		keyToItem:        make(map[Key]*item[Value]),
	}
}

// item are the values that are held in the Cache's map.
type item[Value any] struct {
	value  Value
	expiry *time.Time
}

// Set is the implementation of the Cache interface.
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
	c.rwMutex.Lock()
	c.keyToItem[key] = itemToAdd
	c.rwMutex.Unlock()
}

// Get is the implementation of the Cache interface.
func (c *Cache[Key, Value]) Get(key Key) (Value, bool) {
	c.rwMutex.RLock()
	itemValue, loaded := c.keyToItem[key]
	c.rwMutex.RUnlock()

	if loaded {
		if itemValue.expiry != nil && time.Now().After(*itemValue.expiry) {
			c.clearIfExpired(key)
			var zeroValue Value
			return zeroValue, false
		}
		return itemValue.value, true
	} else {
		var zeroValue Value
		return zeroValue, false
	}
}

// clearIfExpired removes the key from the Cache if it is expired.
func (c *Cache[Key, Value]) clearIfExpired(key Key) {
	c.rwMutex.Lock()
	itemValue, loaded := c.keyToItem[key]
	if loaded && itemValue.expiry != nil && time.Now().After(*itemValue.expiry) {
		delete(c.keyToItem, key)
	}
	c.rwMutex.Unlock()
}

// GetOrSet is the implementation of the Cache interface.
func (c *Cache[Key, Value]) GetOrSet(key Key, fn GetOrSetFn[Key, Value]) (Value, error) {
	c.getOrSetLock.Lock()
	keyLock, keyLockFound := c.getOrSetKeyLocks[key]
	if !keyLockFound {
		keyLock = &getOrSetKeyLock[Value]{
			Count:    atomic.Int32{},
			WaitChan: make(chan struct{}),
		}
		keyLock.Count.Store(1)
		c.getOrSetKeyLocks[key] = keyLock
	} else {
		keyLock.Count.Add(1)
	}
	c.getOrSetLock.Unlock()

	defer func() {
		// The decrement is outside the lock for performance in concurrent scenarios.
		if keyLock.Count.Add(-1) == 0 {
			c.getOrSetLock.Lock()
			// Zero is checked again since the decrement is outside the lock.
			if keyLock.Count.Load() == 0 {
				delete(c.getOrSetKeyLocks, key)
			}
			c.getOrSetLock.Unlock()
		}
	}()

	if keyLockFound {
		<-keyLock.WaitChan
		return keyLock.FnValue, keyLock.FnError
	}

	var valueFound bool
	keyLock.FnValue, valueFound = c.Get(key)
	if valueFound {
		close(keyLock.WaitChan)
		return keyLock.FnValue, nil
	}

	var ttl *time.Duration
	keyLock.FnValue, ttl, keyLock.FnError = fn(key)
	defer close(keyLock.WaitChan)
	if keyLock.FnError != nil {
		return keyLock.FnValue, keyLock.FnError
	}

	c.Set(key, keyLock.FnValue, ttl)
	return keyLock.FnValue, nil
}

// Remove is the implementation of the Cache interface.
func (c *Cache[Key, Value]) Remove(key Key) {
	c.rwMutex.Lock()
	delete(c.keyToItem, key)
	c.rwMutex.Unlock()
}

// Reset is the implementation of the Cache interface.
func (c *Cache[Key, Value]) Reset() {
	c.rwMutex.Lock()
	c.keyToItem = make(map[Key]*item[Value])
	c.rwMutex.Unlock()
}
