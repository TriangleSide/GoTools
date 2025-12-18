package heap_test

import (
	"crypto/rand"
	"math"
	"math/big"
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/datastructures/heap"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func getRandomInt(t *testing.T, maxValue int) int {
	t.Helper()
	randomValueBig, err := rand.Int(rand.Reader, big.NewInt(int64(maxValue)))
	assert.Nil(t, err)
	return int(randomValueBig.Int64())
}

func TestNew_WhenCreatingNewHeap_ShouldInitializeCorrectly(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	assert.Equals(t, 0, maxHeap.Size())
}

func TestHeap_WhenSingleElement_ShouldWorkCorrectly(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	maxHeap.Push(42)
	assert.Equals(t, 1, maxHeap.Size())
	assert.Equals(t, 42, maxHeap.Peek())
	assert.Equals(t, 42, maxHeap.Pop())
	assert.Equals(t, 0, maxHeap.Size())
}

func TestHeap_WhenTwoElements_ShouldBubbleDownCorrectly(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	maxHeap.Push(10)
	maxHeap.Push(20)
	assert.Equals(t, 2, maxHeap.Size())
	assert.Equals(t, 20, maxHeap.Pop())
	assert.Equals(t, 1, maxHeap.Size())
	assert.Equals(t, 10, maxHeap.Peek())

	minHeap := heap.New(func(a, b int) bool { return a < b })
	minHeap.Push(20)
	minHeap.Push(10)
	assert.Equals(t, 2, minHeap.Size())
	assert.Equals(t, 10, minHeap.Pop())
	assert.Equals(t, 1, minHeap.Size())
	assert.Equals(t, 20, minHeap.Peek())
}

func TestPush_WhenPushingElements_ShouldMaintainHeapProperty(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	maxHeap.Push(10)
	assert.Equals(t, maxHeap.Peek(), 10)
	maxHeap.Push(20)
	assert.Equals(t, maxHeap.Peek(), 20)
	maxHeap.Push(5)
	assert.Equals(t, maxHeap.Peek(), 20)
	assert.Equals(t, 3, maxHeap.Size())

	minHeap := heap.New(func(a, b int) bool { return a < b })
	minHeap.Push(10)
	assert.Equals(t, minHeap.Peek(), 10)
	minHeap.Push(20)
	assert.Equals(t, minHeap.Peek(), 10)
	minHeap.Push(5)
	assert.Equals(t, minHeap.Peek(), 5)
	assert.Equals(t, 3, minHeap.Size())
}

func TestPush_WhenPushingMultipleElements_ShouldHandleDuplicatesCorrectly(t *testing.T) {
	t.Parallel()

	const dupCount = 100

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	for range dupCount {
		maxHeap.Push(10)
		maxHeap.Push(5)
		maxHeap.Push(20)
	}
	for range dupCount {
		assert.Equals(t, 20, maxHeap.Pop())
	}
	for range dupCount {
		assert.Equals(t, 10, maxHeap.Pop())
	}
	for range dupCount {
		assert.Equals(t, 5, maxHeap.Pop())
	}

	minHeap := heap.New(func(a, b int) bool { return a < b })
	for range dupCount {
		minHeap.Push(10)
		minHeap.Push(5)
		minHeap.Push(20)
	}
	for range dupCount {
		assert.Equals(t, 5, minHeap.Pop())
	}
	for range dupCount {
		assert.Equals(t, 10, minHeap.Pop())
	}
	for range dupCount {
		assert.Equals(t, 20, minHeap.Pop())
	}
}

func TestPop_WhenPoppingElements_ShouldMaintainHeapProperty(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	maxHeap.Push(10)
	assert.Equals(t, maxHeap.Peek(), 10)
	maxHeap.Push(20)
	assert.Equals(t, maxHeap.Peek(), 20)
	maxHeap.Push(5)
	assert.Equals(t, maxHeap.Peek(), 20)
	assert.Equals(t, 20, maxHeap.Pop())
	assert.Equals(t, 10, maxHeap.Pop())
	assert.Equals(t, 5, maxHeap.Pop())
	assert.Equals(t, 0, maxHeap.Size())

	minHeap := heap.New(func(a, b int) bool { return a < b })
	minHeap.Push(10)
	assert.Equals(t, minHeap.Peek(), 10)
	minHeap.Push(20)
	assert.Equals(t, minHeap.Peek(), 10)
	minHeap.Push(5)
	assert.Equals(t, minHeap.Peek(), 5)
	assert.Equals(t, 5, minHeap.Pop())
	assert.Equals(t, 10, minHeap.Pop())
	assert.Equals(t, 20, minHeap.Pop())
	assert.Equals(t, 0, minHeap.Size())
}

func TestPop_WhenPoppingFromEmptyHeap_ShouldPanic(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	assert.PanicExact(t, func() { maxHeap.Pop() }, "heap is empty")

	minHeap := heap.New(func(a, b int) bool { return a < b })
	assert.PanicExact(t, func() { minHeap.Pop() }, "heap is empty")
}

func TestPeek_WhenPeekingElements_ShouldReturnRootWithoutRemoving(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	maxHeap.Push(15)
	maxHeap.Push(25)
	maxHeap.Push(10)
	assert.Equals(t, 25, maxHeap.Peek())
	assert.Equals(t, 3, maxHeap.Size())

	minHeap := heap.New(func(a, b int) bool { return a < b })
	minHeap.Push(15)
	minHeap.Push(25)
	minHeap.Push(10)
	assert.Equals(t, 10, minHeap.Peek())
	assert.Equals(t, 3, minHeap.Size())
}

func TestPeek_WhenPeekingOnEmptyHeap_ShouldPanic(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	assert.PanicExact(t, func() { maxHeap.Peek() }, "heap is empty")

	minHeap := heap.New(func(a, b int) bool { return a < b })
	assert.PanicExact(t, func() { minHeap.Peek() }, "heap is empty")
}

func TestHeap_WhenValuesAreAddedRandomly_ShouldRetainHeapProperties(t *testing.T) {
	t.Parallel()

	const count = 10000
	minHeap := heap.New(func(a, b int) bool { return a < b })
	valueToCount := make(map[int]int, count)

	for range count {
		randomValue := getRandomInt(t, count/10)
		minHeap.Push(randomValue)
		valueToCount[randomValue]++

		randomValue = getRandomInt(t, count/10)
		minHeap.Push(randomValue)
		valueToCount[randomValue]++

		valueRemoved := minHeap.Pop()
		valueToCount[valueRemoved]--
		if valueToCount[valueRemoved] < 0 {
			t.Fatalf("Value %d was not added.", valueRemoved)
		}
		if valueToCount[valueRemoved] == 0 {
			delete(valueToCount, valueRemoved)
		}
	}

	lastValue := -1
	for range count {
		valueRemoved := minHeap.Pop()
		assert.True(t, valueRemoved >= lastValue)

		valueToCount[valueRemoved]--
		if valueToCount[valueRemoved] < 0 {
			t.Fatalf("Value %d was not added.", valueRemoved)
		}
		if valueToCount[valueRemoved] == 0 {
			delete(valueToCount, valueRemoved)
		}

		lastValue = valueRemoved
	}

	assert.Equals(t, 0, minHeap.Size())
	assert.Equals(t, 0, len(valueToCount))
}

func TestHeap_WhenAccessedConcurrently_ShouldHaveNoIssues(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	const countPerRoutine = 5000
	const routineCount = 4

	var waitGroup sync.WaitGroup
	waitToStart := make(chan struct{})

	for range routineCount {
		waitGroup.Go(func() {
			<-waitToStart
			for range countPerRoutine {
				randomValue := getRandomInt(t, countPerRoutine/10)
				maxHeap.Push(randomValue)
				randomValue = getRandomInt(t, countPerRoutine/10)
				maxHeap.Push(randomValue)

				_ = maxHeap.Size()

				if maxHeap.Size() > 0 {
					_ = maxHeap.Peek()
				}

				if getRandomInt(t, 2) == 0 {
					maxHeap.Pop()
				} else {
					maxHeap.CompareAndPop(func(v int) bool { return v > -1 })
				}
			}
		})
	}

	close(waitToStart)
	waitGroup.Wait()

	lastValue := math.MaxInt
	for range countPerRoutine * routineCount {
		value := maxHeap.Pop()
		assert.True(t, value <= lastValue)
		lastValue = value
	}

	assert.Equals(t, 0, maxHeap.Size())
}

func TestCompareAndPop_WhenCalledOnEmptyHeap_ShouldReturnZeroValueAndFalse(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	value, found := maxHeap.CompareAndPop(func(int) bool { return true })
	assert.Equals(t, 0, value)
	assert.False(t, found)

	minHeap := heap.New(func(a, b int) bool { return a < b })
	value, found = minHeap.CompareAndPop(func(int) bool { return true })
	assert.Equals(t, 0, value)
	assert.False(t, found)
}

func TestCompareAndPop_WhenCalled_ShouldBehaveCorrectly(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		isMaxHeap     bool
		pushValues    []int
		condition     func(int) bool
		shouldPop     bool
		expectedValue int
		expectedSize  int
		expectedPeek  int
	}{
		{
			name:          "when condition is met on max heap it should pop",
			isMaxHeap:     true,
			pushValues:    []int{10, 20, 5},
			condition:     func(v int) bool { return v == 20 },
			shouldPop:     true,
			expectedValue: 20,
			expectedSize:  2,
			expectedPeek:  10,
		},
		{
			name:          "when condition is not met on max heap it should not pop",
			isMaxHeap:     true,
			pushValues:    []int{10, 20, 5},
			condition:     func(v int) bool { return v == 15 },
			shouldPop:     false,
			expectedValue: 0,
			expectedSize:  3,
			expectedPeek:  20,
		},
		{
			name:          "when condition is met on min heap it should pop",
			isMaxHeap:     false,
			pushValues:    []int{10, 20, 5},
			condition:     func(v int) bool { return v == 5 },
			shouldPop:     true,
			expectedValue: 5,
			expectedSize:  2,
			expectedPeek:  10,
		},
		{
			name:          "when condition is not met on min heap it should not pop",
			isMaxHeap:     false,
			pushValues:    []int{10, 20, 5},
			condition:     func(v int) bool { return v == 15 },
			shouldPop:     false,
			expectedValue: 0,
			expectedSize:  3,
			expectedPeek:  5,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testHeap *heap.Heap[int]
			if testCase.isMaxHeap {
				testHeap = heap.New(func(a, b int) bool { return a > b })
			} else {
				testHeap = heap.New(func(a, b int) bool { return a < b })
			}

			for _, v := range testCase.pushValues {
				testHeap.Push(v)
			}

			value, found := testHeap.CompareAndPop(testCase.condition)
			assert.Equals(t, testCase.shouldPop, found)
			assert.Equals(t, testCase.expectedValue, value)
			assert.Equals(t, testCase.expectedSize, testHeap.Size())
			assert.Equals(t, testCase.expectedPeek, testHeap.Peek())
		})
	}
}

func TestCompareAndPop_WhenCalled_ShouldMaintainHeapProperty(t *testing.T) {
	t.Parallel()

	maxHeap := heap.New(func(a, b int) bool { return a > b })
	maxHeap.Push(30)
	maxHeap.Push(20)
	maxHeap.Push(25)
	maxHeap.Push(10)
	maxHeap.Push(15)

	value, found := maxHeap.CompareAndPop(func(v int) bool { return v > 25 })
	assert.True(t, found)
	assert.Equals(t, 30, value)
	assert.Equals(t, 25, maxHeap.Peek())

	assert.Equals(t, 25, maxHeap.Pop())
	assert.Equals(t, 20, maxHeap.Pop())
	assert.Equals(t, 15, maxHeap.Pop())
	assert.Equals(t, 10, maxHeap.Pop())
	assert.Equals(t, 0, maxHeap.Size())

	minHeap := heap.New(func(a, b int) bool { return a < b })
	minHeap.Push(10)
	minHeap.Push(20)
	minHeap.Push(15)
	minHeap.Push(30)
	minHeap.Push(25)

	value, found = minHeap.CompareAndPop(func(v int) bool { return v < 15 })
	assert.True(t, found)
	assert.Equals(t, 10, value)
	assert.Equals(t, 15, minHeap.Peek())

	assert.Equals(t, 15, minHeap.Pop())
	assert.Equals(t, 20, minHeap.Pop())
	assert.Equals(t, 25, minHeap.Pop())
	assert.Equals(t, 30, minHeap.Pop())
	assert.Equals(t, 0, minHeap.Size())
}
