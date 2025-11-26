package readonly_test

import (
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/datastructures/readonly"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func verifySliceValue[T any](t *testing.T, roSlice *readonly.Slice[T], index int, expectedValue T) {
	t.Helper()
	assert.True(t, roSlice.Size() > 0)
	actualValue := roSlice.At(index)
	assert.Equals(t, expectedValue, actualValue)
}

func TestReadOnlySlice(t *testing.T) {
	t.Parallel()

	t.Run("when the SliceBuilder doesn't have entries it should create an empty Slice", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewSliceBuilder[string]()
		roSlice := builder.Build()
		assert.True(t, roSlice.Size() == 0)
	})

	t.Run("when build gets called on a SliceBuilder twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			builder := readonly.NewSliceBuilder[string]()
			builder.Build()
			builder.Build()
		})
	})

	t.Run("when append gets called after build it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			builder := readonly.NewSliceBuilder[string]()
			builder.Build()
			builder.Append()
		})
	})

	t.Run("when a value is added to a read-only slice it should be retrievable", func(t *testing.T) {
		t.Parallel()
		const value = "value"
		builder := readonly.NewSliceBuilder[string]()
		builder.Append(value)
		roSlice := builder.Build()
		verifySliceValue(t, roSlice, 0, value)
	})

	t.Run("when values are appended multiple times it should contain all of them in order", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewSliceBuilder[string]()
		builder.Append("value1")
		builder.Append("value2")
		roSlice := builder.Build()
		verifySliceValue(t, roSlice, 0, "value1")
		verifySliceValue(t, roSlice, 1, "value2")
	})

	t.Run("when a struct is added to a read-only slice it should be retrievable", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Value int
		}
		builder := readonly.NewSliceBuilder[testStruct]()
		builder.Append(testStruct{Value: 1})
		roSlice := builder.Build()
		gotten := roSlice.At(0)
		assert.Equals(t, gotten.Value, 1)
	})

	t.Run("when adding no values to the SliceBuilder it should create an empty slice", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewSliceBuilder[string]()
		roSlice := builder.Build()
		assert.Equals(t, roSlice.Size(), 0)
	})

	t.Run("when iterating over some values", func(t *testing.T) {
		t.Parallel()
		values := []string{"value1", "value2", "value3"}

		t.Run("it should have all the data", func(t *testing.T) {
			t.Parallel()
			builder := readonly.NewSliceBuilder[string]()
			builder.Append(values...)
			roSlice := builder.Build()
			count := 0
			for _, value := range roSlice.All() {
				assert.Equals(t, value, values[count])
				count++
			}
			assert.Equals(t, count, len(values))
		})

		t.Run("it should be able to handle a break", func(t *testing.T) {
			t.Parallel()
			builder := readonly.NewSliceBuilder[string]()
			builder.Append(values...)
			roSlice := builder.Build()
			count := 0
			for _, value := range roSlice.All() {
				assert.Equals(t, value, values[count])
				count++
				break
			}
			assert.Equals(t, count, 1)
		})
	})

	t.Run("when iterating over an empty slice it should do nothing", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewSliceBuilder[string]()
		roSlice := builder.Build()
		count := 0
		for range roSlice.All() {
			count++
		}
		assert.Equals(t, count, 0)
	})

	t.Run("when many threads use the slice it should have no issues", func(t *testing.T) {
		t.Parallel()

		const entryCount = 1000
		const goRoutineCount = 4
		wg := sync.WaitGroup{}
		waitToStart := make(chan struct{})

		builder := readonly.NewSliceBuilder[int]()
		for i := range entryCount {
			builder.Append(i)
		}
		roSlice := builder.Build()

		for range goRoutineCount {
			wg.Go(func() {
				<-waitToStart
				for k := range entryCount {
					verifySliceValue(t, roSlice, k, k)
				}
			})
		}

		close(waitToStart)
		wg.Wait()
	})
}
