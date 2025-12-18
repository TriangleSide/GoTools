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

func TestSliceBuilder_NoEntries_CreatesEmptySlice(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[string]()
	roSlice := builder.Build()
	assert.True(t, roSlice.Size() == 0)
}

func TestSliceBuilder_BuildCalledTwice_Panics(t *testing.T) {
	t.Parallel()
	assert.Panic(t, func() {
		builder := readonly.NewSliceBuilder[string]()
		builder.Build()
		builder.Build()
	})
}

func TestSliceBuilder_AppendAfterBuild_Panics(t *testing.T) {
	t.Parallel()
	assert.Panic(t, func() {
		builder := readonly.NewSliceBuilder[string]()
		builder.Build()
		builder.Append()
	})
}

func TestSlice_SingleValue_IsRetrievable(t *testing.T) {
	t.Parallel()
	const value = "value"
	builder := readonly.NewSliceBuilder[string]()
	builder.Append(value)
	roSlice := builder.Build()
	verifySliceValue(t, roSlice, 0, value)
}

func TestSlice_MultipleAppends_ContainsAllInOrder(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[string]()
	builder.Append("value1")
	builder.Append("value2")
	roSlice := builder.Build()
	verifySliceValue(t, roSlice, 0, "value1")
	verifySliceValue(t, roSlice, 1, "value2")
}

func TestSlice_StructValue_IsRetrievable(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value int
	}
	builder := readonly.NewSliceBuilder[testStruct]()
	builder.Append(testStruct{Value: 1})
	roSlice := builder.Build()
	gotten := roSlice.At(0)
	assert.Equals(t, gotten.Value, 1)
}

func TestSliceBuilder_NoValues_CreatesEmptySlice(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[string]()
	roSlice := builder.Build()
	assert.Equals(t, roSlice.Size(), 0)
}

func TestSlice_All_IteratesAllValues(t *testing.T) {
	t.Parallel()
	values := []string{"value1", "value2", "value3"}
	builder := readonly.NewSliceBuilder[string]()
	builder.Append(values...)
	roSlice := builder.Build()
	count := 0
	for _, value := range roSlice.All() {
		assert.Equals(t, value, values[count])
		count++
	}
	assert.Equals(t, count, len(values))
}

func TestSlice_All_HandlesBreak(t *testing.T) {
	t.Parallel()
	values := []string{"value1", "value2", "value3"}
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
}

func TestSlice_All_EmptySlice_DoesNothing(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[string]()
	roSlice := builder.Build()
	count := 0
	for range roSlice.All() {
		count++
	}
	assert.Equals(t, count, 0)
}

func TestSlice_ConcurrentAccess_NoIssues(t *testing.T) {
	t.Parallel()

	const entryCount = 1000
	const goRoutineCount = 4
	var waitGroup sync.WaitGroup
	waitToStart := make(chan struct{})

	builder := readonly.NewSliceBuilder[int]()
	for i := range entryCount {
		builder.Append(i)
	}
	roSlice := builder.Build()

	for range goRoutineCount {
		waitGroup.Go(func() {
			<-waitToStart
			for k := range entryCount {
				verifySliceValue(t, roSlice, k, k)
			}
		})
	}

	close(waitToStart)
	waitGroup.Wait()
}

func TestSlice_At_NegativeIndex_Panics(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[string]()
	builder.Append("value")
	roSlice := builder.Build()
	assert.Panic(t, func() {
		roSlice.At(-1)
	})
}

func TestSlice_At_IndexOutOfBounds_Panics(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[string]()
	builder.Append("value")
	roSlice := builder.Build()
	assert.Panic(t, func() {
		roSlice.At(1)
	})
}

func TestSlice_At_EmptySlice_Panics(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[string]()
	roSlice := builder.Build()
	assert.Panic(t, func() {
		roSlice.At(0)
	})
}

func TestSliceBuilder_ChainedAppend_ReturnsBuilder(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[int]()
	result := builder.Append(1).Append(2).Append(3)
	assert.Equals(t, builder, result)
	roSlice := result.Build()
	assert.Equals(t, roSlice.Size(), 3)
	assert.Equals(t, roSlice.At(0), 1)
	assert.Equals(t, roSlice.At(1), 2)
	assert.Equals(t, roSlice.At(2), 3)
}

func TestSlice_All_ReturnsCorrectIndices(t *testing.T) {
	t.Parallel()
	builder := readonly.NewSliceBuilder[string]()
	builder.Append("a", "b", "c")
	roSlice := builder.Build()
	expectedIndices := []int{0, 1, 2}
	actualIndices := make([]int, 0, 3)
	for idx := range roSlice.All() {
		actualIndices = append(actualIndices, idx)
	}
	assert.Equals(t, len(expectedIndices), len(actualIndices))
	for i := range expectedIndices {
		assert.Equals(t, expectedIndices[i], actualIndices[i])
	}
}
