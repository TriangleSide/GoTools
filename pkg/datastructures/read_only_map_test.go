package datastructures_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/datastructures"
)

func readOnlyMapMustHaveKeyAndValueGet[Key comparable, Value any](roMap datastructures.ReadOnlyMap[Key, Value], key Key, expectedValue Value) {
	value := roMap.Get(key)
	ExpectWithOffset(1, value).To(Equal(expectedValue))
}

func readOnlyMapMustHaveKeyAndValueFetch[Key comparable, Value any](roMap datastructures.ReadOnlyMap[Key, Value], key Key, expectedValue Value) {
	value, ok := roMap.Fetch(key)
	ExpectWithOffset(1, ok).To(BeTrue())
	ExpectWithOffset(1, value).To(Equal(expectedValue))
}

var _ = Describe("ReadOnlyMap", Ordered, func() {
	var (
		builder datastructures.ReadOnlyMapBuilder[string, string]
		roMap   datastructures.ReadOnlyMap[string, string]
	)

	BeforeEach(func() {
		builder = datastructures.NewReadOnlyMapBuilder[string, string]()
	})

	AfterEach(func() {
		builder = nil
		roMap = nil
	})

	When("the builder is new", func() {
		It("should create an empty ReadOnlyMap when Build is called without any entries", func() {
			roMap = builder.Build()
			Expect(roMap.Keys()).To(BeEmpty())
			Expect(roMap.Size()).To(Equal(0))
		})

		It("should panic if Build is called twice", func() {
			roMap = builder.Build()
			Expect(func() {
				builder.Build()
			}).To(PanicWith(ContainSubstring("Build has already been called on this builder.")))
		})

		It("should panic if Set is called after Build", func() {
			roMap = builder.Build()
			Expect(func() {
				builder.Set(datastructures.ReadOnlyMapBuilderEntry[string, string]{Key: "key", Value: "value"})
			}).To(PanicWith(ContainSubstring("Build has already been called on this builder.")))
		})

		It("should panic if SetMap is called after Build", func() {
			roMap = builder.Build()
			Expect(func() {
				builder.SetMap(map[string]string{"key": "value"})
			}).To(PanicWith(ContainSubstring("Build has already been called on this builder.")))
		})
	})

	When("adding entries with Set", func() {
		const key = "key"
		const value = "value"

		BeforeEach(func() {
			builder.Set(datastructures.ReadOnlyMapBuilderEntry[string, string]{Key: key, Value: value})
			roMap = builder.Build()
		})

		It("should have the key and value in the ReadOnlyMap using Get", func() {
			readOnlyMapMustHaveKeyAndValueGet(roMap, key, value)
		})

		It("should have the key and value in the ReadOnlyMap using Fetch", func() {
			readOnlyMapMustHaveKeyAndValueFetch(roMap, key, value)
		})

		It("should return correct keys", func() {
			Expect(roMap.Keys()).To(ConsistOf(key))
		})

		It("should return false for Has on a missing key", func() {
			Expect(roMap.Has("missing")).To(BeFalse())
		})

		It("should return zero value for Get on a missing key", func() {
			value := roMap.Get("missing")
			var zeroValue string
			Expect(value).To(Equal(zeroValue))
		})

		It("should return false for Fetch on a missing key", func() {
			_, ok := roMap.Fetch("missing")
			Expect(ok).To(BeFalse())
		})

		It("should return correct size", func() {
			Expect(roMap.Size()).To(Equal(1))
		})
	})

	When("adding entries with SetMap", func() {
		var testMap map[string]string

		BeforeEach(func() {
			testMap = map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			builder.SetMap(testMap)
			roMap = builder.Build()
		})

		It("should have all keys and values from the map using Get", func() {
			for key, value := range testMap {
				readOnlyMapMustHaveKeyAndValueGet(roMap, key, value)
			}
		})

		It("should have all keys and values from the map using Fetch", func() {
			for key, value := range testMap {
				readOnlyMapMustHaveKeyAndValueFetch(roMap, key, value)
			}
		})

		It("should return all keys", func() {
			Expect(roMap.Keys()).To(ConsistOf("key1", "key2"))
		})

		It("should return correct size", func() {
			Expect(roMap.Size()).To(Equal(2))
		})
	})

	When("adding duplicate keys", func() {
		const key = "key"
		const value1 = "value1"
		const value2 = "value2"

		BeforeEach(func() {
			builder.Set(
				datastructures.ReadOnlyMapBuilderEntry[string, string]{Key: key, Value: value1},
				datastructures.ReadOnlyMapBuilderEntry[string, string]{Key: key, Value: value2},
			)
			roMap = builder.Build()
		})

		It("should have the last value for the key using Get", func() {
			readOnlyMapMustHaveKeyAndValueGet(roMap, key, value2)
			Expect(roMap.Keys()).To(HaveLen(1))
			Expect(roMap.Size()).To(Equal(1))
		})

		It("should have the last value for the key using Fetch", func() {
			readOnlyMapMustHaveKeyAndValueFetch(roMap, key, value2)
			Expect(roMap.Keys()).To(HaveLen(1))
			Expect(roMap.Size()).To(Equal(1))
		})
	})

	When("building multiple ReadOnlyMaps from different builders", func() {
		var builder1, builder2 datastructures.ReadOnlyMapBuilder[string, string]
		var roMap1, roMap2 datastructures.ReadOnlyMap[string, string]

		BeforeEach(func() {
			builder1 = datastructures.NewReadOnlyMapBuilder[string, string]()
			builder2 = datastructures.NewReadOnlyMapBuilder[string, string]()

			builder1.Set(datastructures.ReadOnlyMapBuilderEntry[string, string]{Key: "key1", Value: "value1"})
			builder2.Set(datastructures.ReadOnlyMapBuilderEntry[string, string]{Key: "key2", Value: "value2"})

			roMap1 = builder1.Build()
			roMap2 = builder2.Build()
		})

		It("should have separate maps using Get", func() {
			readOnlyMapMustHaveKeyAndValueGet(roMap1, "key1", "value1")
			Expect(roMap1.Has("key2")).To(BeFalse())

			readOnlyMapMustHaveKeyAndValueGet(roMap2, "key2", "value2")
			Expect(roMap2.Has("key1")).To(BeFalse())
		})

		It("should have separate maps using Fetch", func() {
			readOnlyMapMustHaveKeyAndValueFetch(roMap1, "key1", "value1")
			_, ok := roMap1.Fetch("key2")
			Expect(ok).To(BeFalse())

			readOnlyMapMustHaveKeyAndValueFetch(roMap2, "key2", "value2")
			_, ok = roMap2.Fetch("key1")
			Expect(ok).To(BeFalse())
		})

		It("should return correct sizes", func() {
			Expect(roMap1.Size()).To(Equal(1))
			Expect(roMap2.Size()).To(Equal(1))
		})
	})

	When("testing with different key and value types", func() {
		It("should work with integer keys and values using Get and Fetch", func() {
			intBuilder := datastructures.NewReadOnlyMapBuilder[int, int]()
			intBuilder.Set(
				datastructures.ReadOnlyMapBuilderEntry[int, int]{Key: 1, Value: 100},
				datastructures.ReadOnlyMapBuilderEntry[int, int]{Key: 2, Value: 200},
			)
			intRoMap := intBuilder.Build()

			value := intRoMap.Get(1)
			Expect(value).To(Equal(100))

			value, ok := intRoMap.Fetch(2)
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal(200))

			Expect(intRoMap.Size()).To(Equal(2))
		})

		It("should work with struct keys and values using Get and Fetch", func() {
			type KeyStruct struct {
				ID int
			}
			type ValueStruct struct {
				Name string
			}
			structBuilder := datastructures.NewReadOnlyMapBuilder[KeyStruct, ValueStruct]()
			structBuilder.Set(
				datastructures.ReadOnlyMapBuilderEntry[KeyStruct, ValueStruct]{Key: KeyStruct{ID: 1}, Value: ValueStruct{Name: "Alice"}},
				datastructures.ReadOnlyMapBuilderEntry[KeyStruct, ValueStruct]{Key: KeyStruct{ID: 2}, Value: ValueStruct{Name: "Bob"}},
			)
			structRoMap := structBuilder.Build()

			value := structRoMap.Get(KeyStruct{ID: 1})
			Expect(value).To(Equal(ValueStruct{Name: "Alice"}))

			value, ok := structRoMap.Fetch(KeyStruct{ID: 2})
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal(ValueStruct{Name: "Bob"}))

			Expect(structRoMap.Size()).To(Equal(2))
		})
	})

	When("modifying the slice returned by Keys", func() {
		const key = "key"
		const value = "value"

		BeforeEach(func() {
			builder.Set(datastructures.ReadOnlyMapBuilderEntry[string, string]{Key: key, Value: value})
			roMap = builder.Build()
		})

		It("should not affect the map", func() {
			keys := roMap.Keys()
			Expect(keys).To(ConsistOf(key))
			keys[0] = "modifiedKey"
			Expect(roMap.Has(key)).To(BeTrue())
			Expect(roMap.Has("modifiedKey")).To(BeFalse())
			Expect(roMap.Keys()).To(ConsistOf(key))
			Expect(roMap.Size()).To(Equal(1))
		})
	})

	When("adding no entries", func() {
		BeforeEach(func() {
			roMap = builder.Build()
		})

		It("should create an empty map", func() {
			Expect(roMap.Keys()).To(BeEmpty())
			Expect(roMap.Size()).To(Equal(0))
		})

		It("Get should return zero value for any key", func() {
			value := roMap.Get("anyKey")
			var zeroValue string
			Expect(value).To(Equal(zeroValue))
		})

		It("Fetch should return false for any key", func() {
			_, ok := roMap.Fetch("anyKey")
			Expect(ok).To(BeFalse())
		})
	})

	When("calling SetMap with an empty map", func() {
		BeforeEach(func() {
			emptyMap := map[string]string{}
			builder.SetMap(emptyMap)
			roMap = builder.Build()
		})

		It("should create an empty map", func() {
			Expect(roMap.Keys()).To(BeEmpty())
			Expect(roMap.Size()).To(Equal(0))
		})

		It("Get should return zero value for any key", func() {
			value := roMap.Get("anyKey")
			var zeroValue string
			Expect(value).To(Equal(zeroValue))
		})

		It("Fetch should return false for any key", func() {
			_, ok := roMap.Fetch("anyKey")
			Expect(ok).To(BeFalse())
		})
	})

	When("calling Set with no entries", func() {
		BeforeEach(func() {
			builder.Set()
			roMap = builder.Build()
		})

		It("should create an empty map", func() {
			Expect(roMap.Keys()).To(BeEmpty())
			Expect(roMap.Size()).To(Equal(0))
		})

		It("Get should return zero value for any key", func() {
			value := roMap.Get("anyKey")
			var zeroValue string
			Expect(value).To(Equal(zeroValue))
		})

		It("Fetch should return false for any key", func() {
			_, ok := roMap.Fetch("anyKey")
			Expect(ok).To(BeFalse())
		})
	})

	When("modifying the map passed to SetMap after building", func() {
		var originalMap map[string]string

		BeforeEach(func() {
			originalMap = map[string]string{"key": "value"}
			builder.SetMap(originalMap)
			roMap = builder.Build()
		})

		It("should not affect the ReadOnlyMap using Get", func() {
			originalMap["key"] = "modifiedValue"
			originalMap["newKey"] = "newValue"
			readOnlyMapMustHaveKeyAndValueGet(roMap, "key", "value")
			Expect(roMap.Has("newKey")).To(BeFalse())
			Expect(roMap.Size()).To(Equal(1))
		})

		It("should not affect the ReadOnlyMap using Fetch", func() {
			originalMap["key"] = "modifiedValue"
			originalMap["newKey"] = "newValue"
			readOnlyMapMustHaveKeyAndValueFetch(roMap, "key", "value")
			Expect(roMap.Has("newKey")).To(BeFalse())
			Expect(roMap.Size()).To(Equal(1))
		})
	})

	When("iterating over the ReadOnlyMap", func() {
		It("should iterate over all key-value pairs", func() {
			testData := map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			builder.SetMap(testData)
			roMap = builder.Build()

			collectedData := make(map[string]string)
			for key, value := range roMap.Iterator() {
				collectedData[key] = value
			}

			Expect(collectedData).To(Equal(testData))
			Expect(roMap.Size()).To(Equal(3))
		})

		It("should be able to handle a break in the iteration", func() {
			testData := map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			builder.SetMap(testData)
			roMap = builder.Build()

			collectedData := make(map[string]string)
			for key, value := range roMap.Iterator() {
				collectedData[key] = value
				break
			}

			Expect(len(collectedData)).To(Equal(1))
			Expect(roMap.Size()).To(Equal(3))
		})

		It("should stop iterating when yield returns false", func() {
			testData := map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			builder.SetMap(testData)
			roMap = builder.Build()

			collectedData := make(map[string]string)
			counter := 0
			roMap.Iterator()(func(key string, value string) bool {
				collectedData[key] = value
				counter++
				return false
			})

			Expect(counter).To(Equal(1))
			Expect(len(collectedData)).To(Equal(1))
			Expect(roMap.Size()).To(Equal(3))
		})

		It("should handle an empty map", func() {
			roMap = builder.Build()

			collectedData := make(map[string]string)
			for key, value := range roMap.Iterator() {
				collectedData[key] = value
			}

			Expect(collectedData).To(BeEmpty())
			Expect(roMap.Size()).To(Equal(0))
		})
	})

	When("testing Get and Fetch with zero values", func() {
		type CustomStruct struct {
			Field1 string
			Field2 int
		}

		It("should return zero value for Get when key is missing", func() {
			builder := datastructures.NewReadOnlyMapBuilder[string, CustomStruct]()
			roMap := builder.Build()

			value := roMap.Get("missing")
			var zeroValue CustomStruct
			Expect(value).To(Equal(zeroValue))
			Expect(roMap.Size()).To(Equal(0))
		})

		It("should return false for Fetch when key is missing", func() {
			builder := datastructures.NewReadOnlyMapBuilder[string, CustomStruct]()
			roMap := builder.Build()

			_, ok := roMap.Fetch("missing")
			Expect(ok).To(BeFalse())
			Expect(roMap.Size()).To(Equal(0))
		})
	})

	When("testing Has method consistency with Fetch", func() {
		const key = "key"
		const value = "value"

		BeforeEach(func() {
			builder.Set(datastructures.ReadOnlyMapBuilderEntry[string, string]{Key: key, Value: value})
			roMap = builder.Build()
		})

		It("should return true from Has when Fetch returns true", func() {
			_, ok := roMap.Fetch(key)
			Expect(ok).To(BeTrue())
			Expect(roMap.Has(key)).To(BeTrue())
			Expect(roMap.Size()).To(Equal(1))
		})

		It("should return false from Has when Fetch returns false", func() {
			_, ok := roMap.Fetch("missing")
			Expect(ok).To(BeFalse())
			Expect(roMap.Has("missing")).To(BeFalse())
			Expect(roMap.Size()).To(Equal(1))
		})
	})

	When("testing Get and Fetch with nil values", func() {
		It("should handle nil values correctly", func() {
			type PointerStruct struct {
				Field1 string
			}
			builder := datastructures.NewReadOnlyMapBuilder[string, *PointerStruct]()
			builder.Set(datastructures.ReadOnlyMapBuilderEntry[string, *PointerStruct]{Key: "key", Value: nil})
			roMap := builder.Build()

			value := roMap.Get("key")
			Expect(value).To(BeNil())

			value, ok := roMap.Fetch("key")
			Expect(ok).To(BeTrue())
			Expect(value).To(BeNil())

			Expect(roMap.Size()).To(Equal(1))
		})
	})

	When("testing Get and Fetch with complex types", func() {
		It("should work with slices as values", func() {
			builder := datastructures.NewReadOnlyMapBuilder[string, []int]()
			builder.Set(datastructures.ReadOnlyMapBuilderEntry[string, []int]{Key: "numbers", Value: []int{1, 2, 3}})
			roMap := builder.Build()

			value := roMap.Get("numbers")
			Expect(value).To(Equal([]int{1, 2, 3}))

			value, ok := roMap.Fetch("numbers")
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal([]int{1, 2, 3}))

			Expect(roMap.Size()).To(Equal(1))
		})

		It("should work with maps as values", func() {
			builder := datastructures.NewReadOnlyMapBuilder[string, map[string]int]()
			builder.Set(datastructures.ReadOnlyMapBuilderEntry[string, map[string]int]{Key: "mapKey", Value: map[string]int{"a": 1}})
			roMap := builder.Build()

			value := roMap.Get("mapKey")
			Expect(value).To(Equal(map[string]int{"a": 1}))

			value, ok := roMap.Fetch("mapKey")
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal(map[string]int{"a": 1}))

			Expect(roMap.Size()).To(Equal(1))
		})
	})

	When("testing thread safety (concurrent access)", func() {
		It("should allow concurrent reads", func() {
			builder := datastructures.NewReadOnlyMapBuilder[int, int]()
			const entryCount = 1000
			for i := 0; i < entryCount; i++ {
				builder.Set(datastructures.ReadOnlyMapBuilderEntry[int, int]{Key: i, Value: i * 10})
			}
			roMap := builder.Build()

			const goRoutineCount = 10
			done := make(chan bool)
			for i := 0; i < goRoutineCount; i++ {
				go func() {
					for k := 0; k < entryCount; k++ {
						expected := k * 10
						value := roMap.Get(k)
						Expect(value).To(Equal(expected))
						value, ok := roMap.Fetch(k)
						Expect(ok).To(BeTrue())
						Expect(value).To(Equal(expected))
						Expect(roMap.Has(k)).To(BeTrue())
						Expect(roMap.Keys()).To(HaveLen(entryCount))
						Expect(roMap.Size()).To(Equal(entryCount))
						count := 0
						for _, _ = range roMap.Iterator() {
							count++
						}
						Expect(count).To(Equal(entryCount))
					}
					done <- true
				}()
			}

			for i := 0; i < goRoutineCount; i++ {
				<-done
			}
		})
	})
})
