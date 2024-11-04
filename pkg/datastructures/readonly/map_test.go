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

func TestReadOnlyMap(t *testing.T) {
	t.Parallel()

	t.Run("when the MapBuilder doesnt have entries it should create an empty Map", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewMapBuilder[string, string]()
		roMap := builder.Build()
		assert.True(t, roMap.Size() == 0)
	})

	t.Run("when build gets called on a MapBuilder twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			builder := readonly.NewMapBuilder[string, string]()
			builder.Build()
			builder.Build()
		})
	})

	t.Run("when set gets called after build it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			builder := readonly.NewMapBuilder[string, string]()
			builder.Build()
			builder.Set()
		})
	})

	t.Run("when set map gets called after build it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			builder := readonly.NewMapBuilder[string, string]()
			builder.Build()
			builder.SetMap(map[string]string{})
		})
	})

	t.Run("when a key and value is added to a read only map it should be retrievable", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		const value = "value"
		builder := readonly.NewMapBuilder[string, string]()
		builder.Set(readonly.MapEntry[string, string]{Key: key, Value: value})
		roMap := builder.Build()
		verifyMapKeyAndValue(t, roMap, key, value)
	})

	t.Run("when set map is used with the MapBuilder it should be available in the map", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		const value = "value"
		builder := readonly.NewMapBuilder[string, string]()
		builder.SetMap(map[string]string{key: value})
		roMap := builder.Build()
		verifyMapKeyAndValue(t, roMap, key, value)
	})

	t.Run("when querying for non existing values it should return false", func(t *testing.T) {
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
	})

	t.Run("when values are overwritten in the MapBuilder it should only have the last value", func(t *testing.T) {
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
	})

	t.Run("when a struct is used as a key it should be retrievable", func(t *testing.T) {
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
	})

	t.Run("when modifying the slice returned by keys it should have no impact on the map", func(t *testing.T) {
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
	})

	t.Run("when adding no values to the MapBuilder it should create an empty map", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewMapBuilder[string, string]()
		roMap := builder.Build()
		assert.Equals(t, roMap.Size(), 0)
	})

	t.Run("when the MapBuilder uses set with nothing and set map with an empty map it should create an empty map", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewMapBuilder[string, string]()
		builder.Set()
		builder.SetMap(map[string]string{})
		roMap := builder.Build()
		assert.Equals(t, roMap.Size(), 0)
	})

	t.Run("when modifying the map used in the MapBuilder it should not affect the built map", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewMapBuilder[string, string]()
		mapToSet := map[string]string{
			"key1": "value1",
		}
		builder.SetMap(mapToSet)
		roMap := builder.Build()
		mapToSet["key1"] = "value2"
		assert.Equals(t, roMap.Get("key1"), "value1")
	})

	t.Run("when iterating over some values", func(t *testing.T) {
		t.Parallel()
		iteratingMap := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		t.Run("it should have all the data", func(t *testing.T) {
			t.Parallel()
			builder := readonly.NewMapBuilder[string, string]()
			builder.SetMap(iteratingMap)
			roMap := builder.Build()
			count := 0
			for _, _ = range roMap.All() {
				count++
			}
			assert.Equals(t, count, len(iteratingMap))
		})

		t.Run("it should be able to handle a break", func(t *testing.T) {
			t.Parallel()
			builder := readonly.NewMapBuilder[string, string]()
			builder.SetMap(iteratingMap)
			roMap := builder.Build()
			count := 0
			for _, _ = range roMap.All() {
				count++
				break
			}
			assert.Equals(t, count, 1)
		})

		t.Run("it should be able to handle a false yield", func(t *testing.T) {
			t.Parallel()
			builder := readonly.NewMapBuilder[string, string]()
			builder.SetMap(iteratingMap)
			roMap := builder.Build()
			count := 0
			roMap.All()(func(key string, value string) bool {
				count++
				return false
			})
			assert.Equals(t, count, 1)
		})
	})

	t.Run("when iterating over an empty map it should do nothing", func(t *testing.T) {
		t.Parallel()
		builder := readonly.NewMapBuilder[string, string]()
		builder.SetMap(map[string]string{})
		roMap := builder.Build()
		count := 0
		for _, _ = range roMap.All() {
			count++
		}
		assert.Equals(t, count, 0)
	})

	t.Run("when many threads use the map it should have no issues", func(t *testing.T) {
		t.Parallel()

		const entryCount = 1000
		const goRoutineCount = 8
		wg := sync.WaitGroup{}
		waitToStart := make(chan struct{})

		builder := readonly.NewMapBuilder[int, int]()
		for i := 0; i < entryCount; i++ {
			builder.Set(readonly.MapEntry[int, int]{Key: i, Value: i * 10})
		}
		roMap := builder.Build()

		for i := 0; i < goRoutineCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-waitToStart
				for k := 0; k < entryCount; k++ {
					expected := k * 10
					verifyMapKeyAndValue(t, roMap, k, expected)
				}
			}()
		}

		close(waitToStart)
		wg.Wait()
	})
}
