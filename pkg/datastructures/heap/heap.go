package heap

import (
	"errors"
	"sync"
)

var (
	// errEmptyHeap occurs when attempting to access an element from an empty heap.
	errEmptyHeap = errors.New("heap is empty")
)

const (
	// heapBranchingFactor is the number of children per heap node.
	heapBranchingFactor = 2
	// heapLeftChildOffset is the index delta from a parent to its left child.
	heapLeftChildOffset = 1
	// heapRightChildOffset is the index delta from a parent to its right child.
	heapRightChildOffset = 2
)

// Heap is a tree-based data structure optimized for quickly accessing the minimum or maximum element,
// depending on the comparator. It supports O(log n) insertion and deletion operations. It is commonly
// used in priority queues, heap sort, and graph algorithms.
type Heap[T any] struct {
	hasPriority func(a T, b T) bool
	tree        []T
	lock        sync.RWMutex
}

// New instantiates a Heap.
func New[T any](hasPriority func(a T, b T) bool) *Heap[T] {
	return &Heap[T]{
		hasPriority: hasPriority,
		tree:        make([]T, 0, 1),
	}
}

// Size returns the number of elements in the heap.
func (h *Heap[T]) Size() int {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return len(h.tree)
}

// Push adds a new value to the heap.
func (h *Heap[T]) Push(value T) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.tree = append(h.tree, value)

	index := len(h.tree) - 1
	for index > 0 {
		parentIndex := (index - 1) / heapBranchingFactor
		if !h.hasPriority(h.tree[index], h.tree[parentIndex]) {
			break
		}
		h.tree[index], h.tree[parentIndex] = h.tree[parentIndex], h.tree[index]
		index = parentIndex
	}
}

// bubbleDown moves the top element down the heap to maintain the heap property.
func (h *Heap[T]) bubbleDown() T {
	retValue := h.tree[0]
	h.tree[0] = h.tree[len(h.tree)-1]
	h.tree = h.tree[:len(h.tree)-1]

	index := 0
	for {
		leftIndex := (index * heapBranchingFactor) + heapLeftChildOffset
		if leftIndex >= len(h.tree) {
			break
		}
		swapLeft := h.hasPriority(h.tree[leftIndex], h.tree[index])

		rightIndex := (index * heapBranchingFactor) + heapRightChildOffset
		var swapRight bool
		if rightIndex < len(h.tree) {
			swapRight = h.hasPriority(h.tree[rightIndex], h.tree[index])
		}

		if swapLeft && swapRight {
			if h.hasPriority(h.tree[leftIndex], h.tree[rightIndex]) {
				swapRight = false
			} else {
				swapLeft = false
			}
		}

		if swapLeft {
			h.tree[index], h.tree[leftIndex] = h.tree[leftIndex], h.tree[index]
			index = leftIndex
			continue
		}

		if swapRight {
			h.tree[index], h.tree[rightIndex] = h.tree[rightIndex], h.tree[index]
			index = rightIndex
			continue
		}

		break
	}

	return retValue
}

// Pop removes the largest or smallest element (depending on the comparator) from the heap.
// It panics if the heap is empty.
func (h *Heap[T]) Pop() T {
	h.lock.Lock()
	defer h.lock.Unlock()
	if len(h.tree) == 0 {
		panic(errEmptyHeap)
	}
	return h.bubbleDown()
}

// CompareAndPop checks if the top element satisfies the condition and pops it if true.
// Returns the popped value and true if successful, or zero value and false otherwise.
func (h *Heap[T]) CompareAndPop(shouldPop func(T) bool) (T, bool) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if len(h.tree) == 0 {
		var zeroValue T
		return zeroValue, false
	}

	if !shouldPop(h.tree[0]) {
		var zeroValue T
		return zeroValue, false
	}

	return h.bubbleDown(), true
}

// Peek returns the min or max value on this heap. The access is O(1).
// It panics if the heap is empty.
func (h *Heap[T]) Peek() T {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if len(h.tree) == 0 {
		panic(errEmptyHeap)
	}
	return h.tree[0]
}
