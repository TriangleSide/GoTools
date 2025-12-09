package readonly_test

import (
	"sync"
	"testing"

	"github.com/TriangleSide/GoTools/pkg/datastructures/readonly"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func verifyMapKeyAndValue[Key comparable, Value any](t *testing.T, roMap *readonly.Map[Key, Value], key Key, value Value) {
	t.Helper()
	assert.True(t, roMap.Size() >= 1)
	hasKey := roMap.Has(key)
	assert.True(t, hasKey)
	actual := roMap.Get(key)
	assert.Equals(t, value, actual)
	actual, ok := roMap.Fetch(key)
	assert.True(t, ok)
	assert.Equals(t, value, actual)
	found := false
	for iKey, iValue := range roMap.All() {
		if key == iKey {
			found = true
			assert.Equals(t, value, iValue)
			break
		}
	}
	assert.True(t, found)
}

func TestMapBuilder_WhenNoEntries_ShouldCreateEmptyMap(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	roMap := builder.Build()
	assert.True(t, roMap.Size() == 0)
}

func TestMapBuilder_WhenBuildCalledTwice_ShouldPanic(t *testing.T) {
	t.Parallel()
	assert.Panic(t, func() {
		builder := readonly.NewMapBuilder[string, string]()
		builder.Build()
		builder.Build()
	})
}

func TestMapBuilder_WhenSetCalledAfterBuild_ShouldPanic(t *testing.T) {
	t.Parallel()
	assert.Panic(t, func() {
		builder := readonly.NewMapBuilder[string, string]()
		builder.Build()
		builder.Set()
	})
}

func TestMapBuilder_WhenSetMapCalledAfterBuild_ShouldPanic(t *testing.T) {
	t.Parallel()
	assert.Panic(t, func() {
		builder := readonly.NewMapBuilder[string, string]()
		builder.Build()
		builder.SetMap(map[string]string{})
	})
}

func TestMap_WhenKeyValueAdded_ShouldBeRetrievable(t *testing.T) {
	t.Parallel()
	const key = "key"
	const value = "value"
	builder := readonly.NewMapBuilder[string, string]()
	builder.Set(readonly.MapEntry[string, string]{Key: key, Value: value})
	roMap := builder.Build()
	verifyMapKeyAndValue(t, roMap, key, value)
}

func TestMapBuilder_WhenSetMapUsed_ShouldBeAvailableInMap(t *testing.T) {
	t.Parallel()
	const key = "key"
	const value = "value"
	builder := readonly.NewMapBuilder[string, string]()
	builder.SetMap(map[string]string{key: value})
	roMap := builder.Build()
	verifyMapKeyAndValue(t, roMap, key, value)
}

func TestMap_WhenQueryingNonExistingValues_ShouldReturnFalse(t *testing.T) {
	t.Parallel()
	const key = "key"
	const value = "value"
	builder := readonly.NewMapBuilder[string, string]()
	builder.SetMap(map[string]string{key: value})
	roMap := builder.Build()
	actual := roMap.Get("missing")
	assert.Equals(t, actual, "")
	fetched, ok := roMap.Fetch("missing")
	assert.False(t, ok)
	assert.Equals(t, actual, fetched)
	assert.Equals(t, roMap.Size(), 1)
}

func TestMapBuilder_WhenValuesOverwritten_ShouldOnlyHaveLastValue(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	builder.Set(readonly.MapEntry[string, string]{Key: "key1", Value: "value1"})
	builder.Set(readonly.MapEntry[string, string]{Key: "key1", Value: "value2"})
	builder.Set(readonly.MapEntry[string, string]{Key: "key2", Value: "value3"})
	builder.SetMap(map[string]string{"key2": "value4"})
	builder.SetMap(map[string]string{"key3": "value5"})
	builder.Set(readonly.MapEntry[string, string]{Key: "key3", Value: "value6"})
	roMap := builder.Build()
	verifyMapKeyAndValue(t, roMap, "key1", "value2")
	verifyMapKeyAndValue(t, roMap, "key2", "value4")
	verifyMapKeyAndValue(t, roMap, "key3", "value6")
}

func TestMap_WhenStructUsedAsKey_ShouldBeRetrievable(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Value int
	}
	builder := readonly.NewMapBuilder[testStruct, testStruct]()
	builder.SetMap(map[testStruct]testStruct{
		{Value: 1}: {Value: 2},
	})
	roMap := builder.Build()
	gotten := roMap.Get(testStruct{Value: 1})
	assert.Equals(t, gotten.Value, 2)
}

func TestMap_WhenModifyingKeysSlice_ShouldNotAffectMap(t *testing.T) {
	t.Parallel()
	const key = "key"
	const value = "value"
	builder := readonly.NewMapBuilder[string, string]()
	builder.SetMap(map[string]string{key: value})
	roMap := builder.Build()
	keys := roMap.Keys()
	assert.Equals(t, len(keys), 1)
	assert.Equals(t, keys[0], key)
	keys[0] = "modifiedKey"
	assert.True(t, roMap.Has(key))
	assert.False(t, roMap.Has("modifiedKey"))
	keys = roMap.Keys()
	assert.Equals(t, len(keys), 1)
	assert.Equals(t, keys[0], key)
}

func TestMapBuilder_WhenNoValuesAdded_ShouldCreateEmptyMap(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	roMap := builder.Build()
	assert.Equals(t, roMap.Size(), 0)
}

func TestMapBuilder_WhenSetWithNothingAndSetMapWithEmptyMap_ShouldCreateEmptyMap(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	builder.Set()
	builder.SetMap(map[string]string{})
	roMap := builder.Build()
	assert.Equals(t, roMap.Size(), 0)
}

func TestMapBuilder_WhenModifyingSourceMap_ShouldNotAffectBuiltMap(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	mapToSet := map[string]string{
		"key1": "value1",
	}
	builder.SetMap(mapToSet)
	roMap := builder.Build()
	mapToSet["key1"] = "value2"
	assert.Equals(t, roMap.Get("key1"), "value1")
}

func TestMap_WhenIterating_ShouldHaveAllData(t *testing.T) {
	t.Parallel()
	iteratingMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	builder := readonly.NewMapBuilder[string, string]()
	builder.SetMap(iteratingMap)
	roMap := builder.Build()
	count := 0
	for range roMap.All() {
		count++
	}
	assert.Equals(t, count, len(iteratingMap))
}

func TestMap_WhenIteratingWithBreak_ShouldHandleBreak(t *testing.T) {
	t.Parallel()
	iteratingMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	builder := readonly.NewMapBuilder[string, string]()
	builder.SetMap(iteratingMap)
	roMap := builder.Build()
	count := 0
	for range roMap.All() {
		count++
		break
	}
	assert.Equals(t, count, 1)
}

func TestMap_WhenIteratingWithFalseYield_ShouldHandleFalseYield(t *testing.T) {
	t.Parallel()
	iteratingMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	builder := readonly.NewMapBuilder[string, string]()
	builder.SetMap(iteratingMap)
	roMap := builder.Build()
	count := 0
	roMap.All()(func(key string, value string) bool {
		count++
		return false
	})
	assert.Equals(t, count, 1)
}

func TestMap_WhenIteratingOverEmptyMap_ShouldDoNothing(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	builder.SetMap(map[string]string{})
	roMap := builder.Build()
	count := 0
	for range roMap.All() {
		count++
	}
	assert.Equals(t, count, 0)
}

func TestMap_WhenManyThreadsUseMap_ShouldHaveNoIssues(t *testing.T) {
	t.Parallel()

	const entryCount = 1000
	const goRoutineCount = 8
	wg := sync.WaitGroup{}
	waitToStart := make(chan struct{})

	builder := readonly.NewMapBuilder[int, int]()
	for i := range entryCount {
		builder.Set(readonly.MapEntry[int, int]{Key: i, Value: i * 10})
	}
	roMap := builder.Build()

	for range goRoutineCount {
		wg.Go(func() {
			<-waitToStart
			for k := range entryCount {
				expected := k * 10
				verifyMapKeyAndValue(t, roMap, k, expected)
			}
		})
	}

	close(waitToStart)
	wg.Wait()
}

func TestMap_WhenHasCalledOnNonExistingKey_ShouldReturnFalse(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	builder.Set(readonly.MapEntry[string, string]{Key: "key", Value: "value"})
	roMap := builder.Build()
	assert.False(t, roMap.Has("nonexistent"))
}

func TestMap_WhenKeysCalledOnEmptyMap_ShouldReturnEmptySlice(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	roMap := builder.Build()
	keys := roMap.Keys()
	assert.Equals(t, len(keys), 0)
}

func TestMapBuilder_WhenSetCalledWithMultipleEntries_ShouldAddAllEntries(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, string]()
	builder.Set(
		readonly.MapEntry[string, string]{Key: "key1", Value: "value1"},
		readonly.MapEntry[string, string]{Key: "key2", Value: "value2"},
		readonly.MapEntry[string, string]{Key: "key3", Value: "value3"},
	)
	roMap := builder.Build()
	assert.Equals(t, roMap.Size(), 3)
	assert.Equals(t, roMap.Get("key1"), "value1")
	assert.Equals(t, roMap.Get("key2"), "value2")
	assert.Equals(t, roMap.Get("key3"), "value3")
}

func TestMapBuilder_WhenMethodChaining_ShouldWork(t *testing.T) {
	t.Parallel()
	roMap := readonly.NewMapBuilder[string, string]().
		Set(readonly.MapEntry[string, string]{Key: "key1", Value: "value1"}).
		SetMap(map[string]string{"key2": "value2"}).
		Set(readonly.MapEntry[string, string]{Key: "key3", Value: "value3"}).
		Build()
	assert.Equals(t, roMap.Size(), 3)
	assert.Equals(t, roMap.Get("key1"), "value1")
	assert.Equals(t, roMap.Get("key2"), "value2")
	assert.Equals(t, roMap.Get("key3"), "value3")
}

func TestMap_WhenNilValueStored_ShouldBeRetrievable(t *testing.T) {
	t.Parallel()
	builder := readonly.NewMapBuilder[string, *string]()
	builder.Set(readonly.MapEntry[string, *string]{Key: "nilKey", Value: nil})
	str := "value"
	builder.Set(readonly.MapEntry[string, *string]{Key: "nonNilKey", Value: &str})
	roMap := builder.Build()
	assert.Equals(t, roMap.Size(), 2)
	assert.True(t, roMap.Has("nilKey"))
	nilValue, ok := roMap.Fetch("nilKey")
	assert.True(t, ok)
	assert.True(t, nilValue == nil)
	nonNilValue := roMap.Get("nonNilKey")
	assert.Equals(t, *nonNilValue, "value")
}
