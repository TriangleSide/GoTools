package ds_test

import (
	"testing"

	"github.com/TriangleSide/GoBase/pkg/ds"
)

func TestReadOnlyMap(t *testing.T) {
	t.Parallel()

	newBuilder := func() ds.ReadOnlyMapBuilder[string, string] {
		return ds.NewReadOnlyMapBuilder[string, string]()
	}

	getEquals := func(t *testing.T, roMap ds.ReadOnlyMap[string, string], key string, value string) {
		actual := roMap.Get(key)
		if value != actual {
			t.Fatalf("roMap.Get(key) = %v, want %v", actual, key)
		}
	}

	fetchEquals := func(t *testing.T, roMap ds.ReadOnlyMap[string, string], key string, value string) {
		actual, ok := roMap.Fetch(key)
		if !ok {
			t.Fatalf("roMap.Fetch(key) = false, want true")
		}
		if value != actual {
			t.Fatalf("roMap.Fetch(key) = %v, want %v", actual, key)
		}
	}

	t.Run("when the builder doesnt have entries it should create an empty ReadOnlyMap", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
		roMap := builder.Build()
		if roMap.Size() != 0 {
			t.Fatalf("roMap has size %d but expected 0", roMap.Size())
		}
	})

	t.Run("when build gets called on a builder twice it should panic", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		builder := newBuilder()
		builder.Build()
		builder.Build()
	})

	t.Run("when set gets called after build it should panic", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		builder := newBuilder()
		builder.Build()
		builder.Set()
	})

	t.Run("when set map gets called after build it should panic", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		builder := newBuilder()
		builder.Build()
		builder.SetMap(map[string]string{})
	})

	t.Run("when a key and value is added to a read only map it should be retrievable", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		const value = "value"
		builder := newBuilder()
		builder.Set(ds.ReadOnlyMapBuilderEntry[string, string]{Key: key, Value: value})
		roMap := builder.Build()
		getEquals(t, roMap, key, value)
		fetchEquals(t, roMap, key, value)
		if roMap.Size() != 1 {
			t.Fatalf("roMap has size %d but expected 1", roMap.Size())
		}
		keys := roMap.Keys()
		if keys[0] != key {
			t.Fatalf("roMap.Keys() = %v, want %v", keys, []string{key})
		}
		has := roMap.Has(key)
		if !has {
			t.Fatalf("roMap.Has() = false, want true")
		}
		count := 0
		for _, _ = range roMap.Iterator() {
			count++
		}
		if count != 1 {
			t.Fatalf("iterator should have 1 value")
		}
	})

	t.Run("when set map is used with the builder it should be available in the map", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		const value = "value"
		builder := newBuilder()
		builder.SetMap(map[string]string{key: value})
		roMap := builder.Build()
		getEquals(t, roMap, key, value)
		fetchEquals(t, roMap, key, value)
		if roMap.Size() != 1 {
			t.Fatalf("roMap has size %d but expected 1", roMap.Size())
		}
	})

	t.Run("when querying for non existing values it should return false", func(t *testing.T) {
		t.Parallel()
		const key = "key"
		const value = "value"
		builder := newBuilder()
		builder.SetMap(map[string]string{key: value})
		roMap := builder.Build()
		actual := roMap.Get("missing")
		if actual != "" {
			t.Fatalf("missing values should be empty")
		}
		fetched, ok := roMap.Fetch("missing")
		if ok {
			t.Fatalf("missing keys should return false")
		}
		if fetched != "" {
			t.Fatalf("missing values should be empty")
		}
		if roMap.Size() != 1 {
			t.Fatalf("roMap has size %d but expected 1", roMap.Size())
		}
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
		getEquals(t, roMap, "key1", "value2")
		getEquals(t, roMap, "key2", "value4")
		getEquals(t, roMap, "key3", "value6")
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
		if gotten.Value != 2 {
			t.Fatalf("got %v, want %v", gotten.Value, 2)
		}
	})

	t.Run("when modifying the slice returned by keys it should have no impact on the map", func(t *testing.T) {
		const key = "key"
		const value = "value"
		builder := newBuilder()
		builder.SetMap(map[string]string{key: value})
		roMap := builder.Build()
		keys := roMap.Keys()
		if len(keys) != 1 {
			t.Fatalf("key size %v but want %v", len(keys), 1)
		}
		if keys[0] != key {
			t.Fatalf("keys are %v but want %v", keys, []string{key})
		}
		keys[0] = "modifiedKey"
		if !roMap.Has(key) {
			t.Fatalf("the map should have the key")
		}
		if roMap.Has("modifiedKey") {
			t.Fatalf("the map should not have the modifiedKey")
		}
		keys = roMap.Keys()
		if len(keys) != 1 {
			t.Fatalf("key size %v but want %v", len(keys), 1)
		}
		if keys[0] != key {
			t.Fatalf("keys are %v but want %v", keys, []string{key})
		}
	})

	t.Run("when adding no values to the builder it should create an empty map", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
		roMap := builder.Build()
		if roMap.Size() != 0 {
			t.Fatalf("roMap has size %d but expected 0", roMap.Size())
		}
	})

	t.Run("when the builder uses set with nothing and set map with an empty map it should create an empty map", func(t *testing.T) {
		t.Parallel()
		builder := newBuilder()
		builder.Set()
		builder.SetMap(map[string]string{})
		roMap := builder.Build()
		if roMap.Size() != 0 {
			t.Fatalf("roMap has size %d but expected 0", roMap.Size())
		}
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
		if roMap.Get("key1") != "value1" {
			t.Fatalf("the map should not be modifiable")
		}
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
			if count != 3 {
				t.Fatalf("iterator should have seen 3 values but got %d", count)
			}
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
			if count != 1 {
				t.Fatalf("iterator should have seen 1 value but got %d", count)
			}
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
			if count != 1 {
				t.Fatalf("iterator should have seen 1 value but got %d", count)
			}
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
		if count != 0 {
			t.Fatalf("iterator should have seen 0 values but got %d", count)
		}
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
					value := roMap.Get(k)
					if value != expected {
						t.Errorf("the value from get is unexpected")
					}
					value, ok := roMap.Fetch(k)
					if !ok {
						t.Errorf("the value in fetch should be present")
					}
					if value != expected {
						t.Errorf("the value from fetch is unexpected")
					}
					if !roMap.Has(k) {
						t.Errorf("the map should have the key")
					}
					if len(roMap.Keys()) != entryCount {
						t.Errorf("the map should have %d keys", entryCount)
					}
					if roMap.Size() != entryCount {
						t.Errorf("the map should have %d entries", entryCount)
					}
					count := 0
					for _, _ = range roMap.Iterator() {
						count++
					}
					if count != entryCount {
						t.Errorf("the map should have %d entries while iterating", entryCount)
					}
				}
				done <- true
			}()
		}
	})
}
