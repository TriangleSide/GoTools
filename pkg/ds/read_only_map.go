package ds

import (
	"iter"
)

// ReadOnlyMap represents a generic read-only map.
// It is the responsibility of the author to ensure that the Value generic is read-only as well.
type ReadOnlyMap[Key comparable, Value any] interface {
	// Get retrieves the value associated with the given key.
	// Returns the zero value of Value if the key does not exist.
	Get(Key) Value

	// Fetch retrieves the value associated with the given key and a boolean indicating if the key exists.
	Fetch(Key) (Value, bool)

	// Has checks if the key exists in the map.
	Has(Key) bool

	// Keys returns a slice of all keys in the map.
	Keys() []Key

	// Iterator iterates over the values of the map.
	Iterator() iter.Seq2[Key, Value]

	// Size returns the number of entries in the map.
	Size() int
}

// readOnlyMap provides a read-only wrapper around a map.
type readOnlyMap[Key comparable, Value any] struct {
	internalMap map[Key]Value
}

// Get retrieves the value associated with the given key.
// Returns the zero value of Value if the key does not exist.
func (r *readOnlyMap[Key, Value]) Get(key Key) Value {
	value, ok := r.internalMap[key]
	if !ok {
		var zeroValue Value
		return zeroValue
	}
	return value
}

// Fetch retrieves the value associated with the given key and a boolean indicating if the key exists.
func (r *readOnlyMap[Key, Value]) Fetch(key Key) (Value, bool) {
	value, ok := r.internalMap[key]
	return value, ok
}

// Has checks if the key exists in the map.
func (r *readOnlyMap[Key, Value]) Has(key Key) bool {
	_, ok := r.internalMap[key]
	return ok
}

// Keys returns a slice of all keys in the map.
func (r *readOnlyMap[Key, Value]) Keys() []Key {
	keys := make([]Key, 0, len(r.internalMap))
	for k := range r.internalMap {
		keys = append(keys, k)
	}
	return keys
}

// Iterator iterates over the values of the map.
func (r *readOnlyMap[Key, Value]) Iterator() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		for key, value := range r.internalMap {
			if !yield(key, value) {
				return
			}
		}
	}
}

// Size returns the number of entries in the map.
func (r *readOnlyMap[Key, Value]) Size() int {
	return len(r.internalMap)
}

// ReadOnlyMapBuilderEntry is a key-value pair for the builder.
type ReadOnlyMapBuilderEntry[Key comparable, Value any] struct {
	Key   Key
	Value Value
}

// ReadOnlyMapBuilder builds a ReadOnlyMap.
type ReadOnlyMapBuilder[Key comparable, Value any] interface {
	// Set adds entries to the builder.
	Set(...ReadOnlyMapBuilderEntry[Key, Value]) ReadOnlyMapBuilder[Key, Value]

	// SetMap adds entries from another map to the builder.
	SetMap(map[Key]Value) ReadOnlyMapBuilder[Key, Value]

	// Build creates a ReadOnlyMap from the builder's entries.
	Build() ReadOnlyMap[Key, Value]
}

// readOnlyMapBuilder is a builder for ReadOnlyMap.
type readOnlyMapBuilder[Key comparable, Value any] struct {
	internalMap map[Key]Value
}

// Set adds entries to the builder.
func (r *readOnlyMapBuilder[Key, Value]) Set(entries ...ReadOnlyMapBuilderEntry[Key, Value]) ReadOnlyMapBuilder[Key, Value] {
	if r.internalMap == nil {
		panic("Build has already been called on this builder.")
	}
	for _, entry := range entries {
		r.internalMap[entry.Key] = entry.Value
	}
	return r
}

// SetMap adds entries from another map to the builder.
func (r *readOnlyMapBuilder[Key, Value]) SetMap(otherMap map[Key]Value) ReadOnlyMapBuilder[Key, Value] {
	if r.internalMap == nil {
		panic("Build has already been called on this builder.")
	}
	for key, value := range otherMap {
		r.internalMap[key] = value
	}
	return r
}

// Build creates a ReadOnlyMap from the builder's entries.
func (r *readOnlyMapBuilder[Key, Value]) Build() ReadOnlyMap[Key, Value] {
	if r.internalMap == nil {
		panic("Build has already been called on this builder.")
	}
	internalMap := r.internalMap
	r.internalMap = nil // This ensures the builder no longer has access to the internal map passed to ReadOnlyMap.
	return &readOnlyMap[Key, Value]{
		internalMap: internalMap,
	}
}

// NewReadOnlyMapBuilder returns a new ReadOnlyMapBuilder.
func NewReadOnlyMapBuilder[Key comparable, Value any]() ReadOnlyMapBuilder[Key, Value] {
	return &readOnlyMapBuilder[Key, Value]{
		internalMap: make(map[Key]Value),
	}
}
