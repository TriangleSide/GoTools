package readonly

import (
	"errors"
	"iter"
	"sync/atomic"
)

// Slice provides a read-only wrapper around a slice.
type Slice[T any] struct {
	internalSlice []T
}

// At retrieves the element at the given index.
func (s *Slice[T]) At(index int) T {
	return s.internalSlice[index]
}

// Size returns the number of elements in the slice.
func (s *Slice[T]) Size() int {
	return len(s.internalSlice)
}

// All iterates over the elements of the slice.
func (s *Slice[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, value := range s.internalSlice {
			if !yield(i, value) {
				return
			}
		}
	}
}

// SliceBuilder builds a Slice.
type SliceBuilder[T any] struct {
	built         atomic.Bool
	internalSlice []T
}

// NewSliceBuilder returns a new SliceBuilder.
func NewSliceBuilder[T any]() *SliceBuilder[T] {
	return &SliceBuilder[T]{}
}

// Append adds elements to the SliceBuilder.
func (b *SliceBuilder[T]) Append(elements ...T) *SliceBuilder[T] {
	if b.built.Load() {
		panic(errors.New("build has already been called on this SliceBuilder"))
	}
	b.internalSlice = append(b.internalSlice, elements...)
	return b
}

// Build creates a Slice from the SliceBuilder's elements.
func (b *SliceBuilder[T]) Build() *Slice[T] {
	if b.built.Swap(true) {
		panic(errors.New("build has already been called on this SliceBuilder"))
	}
	internalSlice := b.internalSlice
	b.internalSlice = nil // This ensures the SliceBuilder no longer has access to the internal slice passed to Slice.
	return &Slice[T]{
		internalSlice: internalSlice,
	}
}
