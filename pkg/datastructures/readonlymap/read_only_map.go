package readonlymap

import (
	"iter"
)

// ReadOnlyMap provides a read-only wrapper around a map.
type ReadOnlyMap[Key comparable, Value any] struct {
	internalMap map[Key]Value
}

// Get retrieves the value associated with the given key.
// Returns the zero value of Value if the key does not exist.
func (r *ReadOnlyMap[Key, Value]) Get(key Key) Value {
	value, ok := r.internalMap[key]
	if !ok {
		var zeroValue Value
		return zeroValue
	}
	return value
}

// Fetch retrieves the value associated with the given key and a boolean indicating if the key exists.
func (r *ReadOnlyMap[Key, Value]) Fetch(key Key) (Value, bool) {
	value, ok := r.internalMap[key]
	return value, ok
}

// Has checks if the key exists in the map.
func (r *ReadOnlyMap[Key, Value]) Has(key Key) bool {
	_, ok := r.internalMap[key]
	return ok
}

// Keys returns a slice of all keys in the map.
func (r *ReadOnlyMap[Key, Value]) Keys() []Key {
	keys := make([]Key, 0, len(r.internalMap))
	for k := range r.internalMap {
		keys = append(keys, k)
	}
	return keys
}

// Iterator iterates over the values of the map.
func (r *ReadOnlyMap[Key, Value]) Iterator() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		for key, value := range r.internalMap {
			if !yield(key, value) {
				return
			}
		}
	}
}

// Size returns the number of entries in the map.
func (r *ReadOnlyMap[Key, Value]) Size() int {
	return len(r.internalMap)
}

// BuilderEntry is a key-value pair for the Builder.
type BuilderEntry[Key comparable, Value any] struct {
	Key   Key
	Value Value
}

// Builder builds a ReadOnlyMap.
type Builder[Key comparable, Value any] struct {
	internalMap map[Key]Value
}

// NewBuilder returns a new Builder.
func NewBuilder[Key comparable, Value any]() *Builder[Key, Value] {
	return &Builder[Key, Value]{
		internalMap: make(map[Key]Value),
	}
}

// Set adds entries to the Builder.
func (r *Builder[Key, Value]) Set(entries ...BuilderEntry[Key, Value]) *Builder[Key, Value] {
	if r.internalMap == nil {
		panic("Build has already been called on this Builder.")
	}
	for _, entry := range entries {
		r.internalMap[entry.Key] = entry.Value
	}
	return r
}

// SetMap adds entries from another map to the Builder.
func (r *Builder[Key, Value]) SetMap(otherMap map[Key]Value) *Builder[Key, Value] {
	if r.internalMap == nil {
		panic("Build has already been called on this Builder.")
	}
	for key, value := range otherMap {
		r.internalMap[key] = value
	}
	return r
}

// Build creates a ReadOnlyMap from the Builder's entries.
func (r *Builder[Key, Value]) Build() *ReadOnlyMap[Key, Value] {
	if r.internalMap == nil {
		panic("Build has already been called on this Builder.")
	}
	internalMap := r.internalMap
	r.internalMap = nil // This ensures the Builder no longer has access to the internal map passed to ReadOnlyMap.
	return &ReadOnlyMap[Key, Value]{
		internalMap: internalMap,
	}
}
