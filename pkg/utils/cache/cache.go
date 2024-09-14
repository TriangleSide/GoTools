package cache

import (
	"reflect"
	"sync"
	"time"
)

const (
	// DoesNotExpire is about 100 years. Hopefully your service does not run for that long.
	DoesNotExpire = time.Hour * 1024 * 1024
)

// GetOrSetFn is used in the GetOrSet function of the Cache interface.
// If the value is not present, or if it's expired, the function gets called.
type GetOrSetFn[Key comparable, Value any] func(Key) (Value, time.Duration, error)

// Cache is a set of methods for a caching mechanism.
type Cache[Key comparable, Value any] interface {
	// Set sets the value with an expiry in the cache.
	Set(Key, Value, time.Duration)

	// Get returns the value if present and not expired from the cache.
	Get(Key) (Value, bool)

	// GetOrSet either gets a value from the cache, or runs a function to fetch and store the value.
	GetOrSet(Key, GetOrSetFn[Key, Value]) (Value, error)

	// Remove removes an item, if present, from the cache.
	Remove(Key)

	// Reset removes all items from the cache.
	Reset()
}

// cache is an implementation of the Cache interface.
type cache[Key comparable, Value any] struct {
	mu        sync.RWMutex
	keyToItem map[Key]*item[Value]
}

// New creates a new instance of the Cache interface.
func New[Key comparable, Value any]() Cache[Key, Value] {
	if reflect.TypeOf(new(Value)).Elem().Kind() == reflect.Pointer {
		panic("Value must not be a pointer because it could be modified while in the cache.")
	}
	return &cache[Key, Value]{
		mu:        sync.RWMutex{},
		keyToItem: make(map[Key]*item[Value]),
	}
}

// item are the values that are held in the cache's map.
type item[Value any] struct {
	value  Value
	expiry time.Time
}

// Set is the implementation of the Cache interface.
func (c *cache[Key, Value]) Set(key Key, value Value, ttl time.Duration) {
	itemToAdd := &item[Value]{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
	c.mu.Lock()
	c.keyToItem[key] = itemToAdd
	c.mu.Unlock()
}

// Get is the implementation of the Cache interface.
func (c *cache[Key, Value]) Get(key Key) (Value, bool) {
	c.mu.RLock()
	itemValue, loaded := c.keyToItem[key]
	c.mu.RUnlock()

	if loaded {
		if time.Now().Before(itemValue.expiry) {
			return itemValue.value, true
		}
		c.clearIfExpired(key)
	}

	var zeroValue Value
	return zeroValue, false
}

// clearIfExpired removes the key from the cache if it is expired.
func (c *cache[Key, Value]) clearIfExpired(key Key) {
	c.mu.Lock()
	itemValue, loaded := c.keyToItem[key]
	if loaded && time.Now().After(itemValue.expiry) {
		delete(c.keyToItem, key)
	}
	c.mu.Unlock()
}

// GetOrSet is the implementation of the Cache interface.
func (c *cache[Key, Value]) GetOrSet(key Key, fn GetOrSetFn[Key, Value]) (Value, error) {
	value, valueLoaded := c.Get(key)
	if valueLoaded {
		return value, nil
	}

	fnValue, ttl, err := fn(key)
	if err != nil {
		return fnValue, err
	}

	c.Set(key, fnValue, ttl)
	return fnValue, nil
}

// Remove is the implementation of the Cache interface.
func (c *cache[Key, Value]) Remove(key Key) {
	c.mu.Lock()
	delete(c.keyToItem, key)
	c.mu.Unlock()
}

// Reset is the implementation of the Cache interface.
func (c *cache[Key, Value]) Reset() {
	c.mu.Lock()
	c.keyToItem = make(map[Key]*item[Value])
	c.mu.Unlock()
}
