package readonly

import (
	"iter"
	"maps"
	"sync/atomic"
)

// Map provides a read-only wrapper around a map.
type Map[Key comparable, Value any] struct {
	internalMap map[Key]Value
}

// Get retrieves the value associated with the given key.
// Returns the zero value of Value if the key does not exist.
func (r *Map[Key, Value]) Get(key Key) Value {
	value, ok := r.internalMap[key]
	if !ok {
		var zeroValue Value
		return zeroValue
	}
	return value
}

// Fetch retrieves the value associated with the given key and a boolean indicating if the key exists.
func (r *Map[Key, Value]) Fetch(key Key) (Value, bool) {
	value, ok := r.internalMap[key]
	return value, ok
}

// Has checks if the key exists in the map.
func (r *Map[Key, Value]) Has(key Key) bool {
	_, ok := r.internalMap[key]
	return ok
}

// Keys returns a slice of all keys in the map.
func (r *Map[Key, Value]) Keys() []Key {
	keys := make([]Key, 0, len(r.internalMap))
	for k := range r.internalMap {
		keys = append(keys, k)
	}
	return keys
}

// All iterates over the key/value pairs of the map.
func (r *Map[Key, Value]) All() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		for key, value := range r.internalMap {
			if !yield(key, value) {
				return
			}
		}
	}
}

// Size returns the number of entries in the map.
func (r *Map[Key, Value]) Size() int {
	return len(r.internalMap)
}

// MapEntry is a key-value pair for the MapBuilder.
type MapEntry[Key comparable, Value any] struct {
	Key   Key
	Value Value
}

// MapBuilder builds a Map.
type MapBuilder[Key comparable, Value any] struct {
	built       atomic.Bool
	internalMap map[Key]Value
}

// NewMapBuilder returns a new MapBuilder.
func NewMapBuilder[Key comparable, Value any]() *MapBuilder[Key, Value] {
	builder := &MapBuilder[Key, Value]{
		built:       atomic.Bool{},
		internalMap: make(map[Key]Value),
	}
	builder.built.Store(false)
	return builder
}

// Set adds entries to the MapBuilder.
func (b *MapBuilder[Key, Value]) Set(entries ...MapEntry[Key, Value]) *MapBuilder[Key, Value] {
	if b.built.Load() {
		panic("Build has already been called on this MapBuilder.")
	}
	for _, entry := range entries {
		b.internalMap[entry.Key] = entry.Value
	}
	return b
}

// SetMap adds entries from another map to the MapBuilder.
func (b *MapBuilder[Key, Value]) SetMap(otherMap map[Key]Value) *MapBuilder[Key, Value] {
	if b.built.Load() {
		panic("Build has already been called on this MapBuilder.")
	}
	maps.Copy(b.internalMap, otherMap)
	return b
}

// Build creates a Map from the MapBuilder's entries.
func (b *MapBuilder[Key, Value]) Build() *Map[Key, Value] {
	if b.built.Swap(true) {
		panic("Build has already been called on this MapBuilder.")
	}
	internalMap := b.internalMap
	b.internalMap = nil // This ensures the MapBuilder no longer has access to the internal map passed to Map.
	return &Map[Key, Value]{
		internalMap: internalMap,
	}
}
