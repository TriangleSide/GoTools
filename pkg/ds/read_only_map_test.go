package ds_test

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/ds"
	"github.com/TriangleSide/GoBase/pkg/test/assert"
)

func verifyKeyAndValue[Key comparable, Value any](t *testing.T, roMap ds.ReadOnlyMap[Key, Value], key Key, value Value) {
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
	for iKey, iValue := range roMap.Iterator() {
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

	newBuilder := func() ds.ReadOnlyMapBuilder[string, string] {
		return ds.NewReadOnlyMapBuilder[string, string]()
	}

	t.Run("when the builder doesnt have entries it should create an empty ReadOnlyMap", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
		roMap := builder.Build()
		assert.True(t, roMap.Size() == 0)
	})

	t.Run("when build gets called on a builder twice it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			builder := newBuilder()
			builder.Build()
			builder.Build()
		})
	})

	t.Run("when set gets called after build it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			builder := newBuilder()
			builder.Build()
			builder.Set()
		})
	})

	t.Run("when set map gets called after build it should panic", func(t *testing.T) {
		t.Parallel()
		assert.Panic(t, func() {
			builder := newBuilder()
			builder.Build()
			builder.SetMap(map[string]string{})
		})
	})

	t.Run("when a key and value is added to a read only map it should be retrievable", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		const value = "value"
		builder := newBuilder()
		builder.Set(ds.ReadOnlyMapBuilderEntry[string, string]{Key: key, Value: value})
		roMap := builder.Build()
		verifyKeyAndValue(t, roMap, key, value)
	})

	t.Run("when set map is used with the builder it should be available in the map", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		const value = "value"
		builder := newBuilder()
		builder.SetMap(map[string]string{key: value})
		roMap := builder.Build()
		verifyKeyAndValue(t, roMap, key, value)
	})

	t.Run("when querying for non existing values it should return false", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		const value = "value"
		builder := newBuilder()
		builder.SetMap(map[string]string{key: value})
		roMap := builder.Build()
		actual := roMap.Get("missing")
		assert.Equals(t, actual, "")
		fetched, ok := roMap.Fetch("missing")
		assert.False(t, ok)
		assert.Equals(t, actual, fetched)
		assert.Equals(t, roMap.Size(), 1)
	})

	t.Run("when values are overwritten in the builder it should only have the last value", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
		builder.Set(ds.ReadOnlyMapBuilderEntry[string, string]{Key: "key1", Value: "value1"})
		builder.Set(ds.ReadOnlyMapBuilderEntry[string, string]{Key: "key1", Value: "value2"})
		builder.Set(ds.ReadOnlyMapBuilderEntry[string, string]{Key: "key2", Value: "value3"})
		builder.SetMap(map[string]string{"key2": "value4"})
		builder.SetMap(map[string]string{"key3": "value5"})
		builder.Set(ds.ReadOnlyMapBuilderEntry[string, string]{Key: "key3", Value: "value6"})
		roMap := builder.Build()
		verifyKeyAndValue(t, roMap, "key1", "value2")
		verifyKeyAndValue(t, roMap, "key2", "value4")
		verifyKeyAndValue(t, roMap, "key3", "value6")
	})

	t.Run("when a struct is used as a key it should be retrievable", func(t *testing.T) {
		t.Parallel()
		type testStruct struct {
			Value int
		}
		builder := ds.NewReadOnlyMapBuilder[testStruct, testStruct]()
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
		builder := newBuilder()
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

	t.Run("when adding no values to the builder it should create an empty map", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
		roMap := builder.Build()
		assert.Equals(t, roMap.Size(), 0)
	})

	t.Run("when the builder uses set with nothing and set map with an empty map it should create an empty map", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
		builder.Set()
		builder.SetMap(map[string]string{})
		roMap := builder.Build()
		assert.Equals(t, roMap.Size(), 0)
	})

	t.Run("when modifying the map used in the builder it should not affect the built map", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
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
			builder := newBuilder()
			builder.SetMap(iteratingMap)
			roMap := builder.Build()
			count := 0
			for _, _ = range roMap.Iterator() {
				count++
			}
			assert.Equals(t, count, len(iteratingMap))
		})

		t.Run("it should be able to handle a break", func(t *testing.T) {
			t.Parallel()
			builder := newBuilder()
			builder.SetMap(iteratingMap)
			roMap := builder.Build()
			count := 0
			for _, _ = range roMap.Iterator() {
				count++
				break
			}
			assert.Equals(t, count, 1)
		})

		t.Run("it should be able to handle a false yield", func(t *testing.T) {
			t.Parallel()
			builder := newBuilder()
			builder.SetMap(iteratingMap)
			roMap := builder.Build()
			count := 0
			roMap.Iterator()(func(key string, value string) bool {
				count++
				return false
			})
			assert.Equals(t, count, 1)
		})
	})

	t.Run("when iterating over an empty map it should do nothing", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
		builder.SetMap(map[string]string{})
		roMap := builder.Build()
		count := 0
		for _, _ = range roMap.Iterator() {
			count++
		}
		assert.Equals(t, count, 0)
	})

	t.Run("when many threads use the map it should have no issues", func(t *testing.T) {
		t.Parallel()

		builder := ds.NewReadOnlyMapBuilder[int, int]()
		const entryCount = 1000
		for i := 0; i < entryCount; i++ {
			builder.Set(ds.ReadOnlyMapBuilderEntry[int, int]{Key: i, Value: i * 10})
		}
		roMap := builder.Build()

		const goRoutineCount = 8
		done := make(chan bool)
		for i := 0; i < goRoutineCount; i++ {
			go func() {
				for k := 0; k < entryCount; k++ {
					expected := k * 10
					verifyKeyAndValue(t, roMap, k, expected)
				}
				done <- true
			}()
		}
	})
}
