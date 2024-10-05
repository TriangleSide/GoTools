package cache_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/utils/cache"
)

func cacheMustHaveKeyAndValue[Key comparable, Value any](cache cache.Cache[Key, Value], key Key, value Value) {
	gottenValue, gotten := cache.Get(key)
	ExpectWithOffset(1, gotten).To(BeTrue())
	ExpectWithOffset(1, value).To(Equal(gottenValue))
}

var _ = Describe("cache", Ordered, func() {
	var (
		testCache = cache.New[string, string]()
	)

	AfterEach(func() {
		testCache.Reset()
	})

	It("should panic if value is a pointer", func() {
		Expect(func() {
			cache.New[string, *string]()
		}).To(PanicWith(ContainSubstring("must not be a pointer")))
	})

	It("should be able to reset the cache repeatedly", func() {
		for i := 0; i < 3; i++ {
			testCache.Reset()
		}
	})

	It("should be able to remove an arbitrary key repeatedly", func() {
		const key = "key"
		for i := 0; i < 3; i++ {
			testCache.Remove(key)
		}
	})

	When("there is no values in the cache", func() {
		It("should return false when getting a key", func() {
			const key = "key"
			_, gotten := testCache.Get(key)
			Expect(gotten).To(BeFalse())
		})

		It("should call the fn with get or set", func() {
			const key = "key"
			const value = "value"
			fnCalled := false
			_, gotten := testCache.Get(key)
			Expect(gotten).To(BeFalse())
			returnVal, err := testCache.GetOrSet(key, func(key string) (string, time.Duration, error) {
				fnCalled = true
				return value, cache.DoesNotExpire, nil
			})
			Expect(err).To(Not(HaveOccurred()))
			Expect(fnCalled).To(BeTrue())
			Expect(returnVal).To(Equal(value))
			_, gotten = testCache.Get(key)
			Expect(gotten).To(BeTrue())
		})

		It("should return an error if it occurs in get or set", func() {
			const key = "key"
			const value = "value"
			fnCalled := false
			_, gotten := testCache.Get(key)
			Expect(gotten).To(BeFalse())
			returnVal, err := testCache.GetOrSet(key, func(key string) (string, time.Duration, error) {
				fnCalled = true
				return value, cache.DoesNotExpire, errors.New("error")
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fnCalled).To(BeTrue())
			Expect(returnVal).To(Equal(value))
			_, gotten = testCache.Get(key)
			Expect(gotten).To(BeFalse())
		})
	})

	When("an item is cached without an expiry time", func() {
		const (
			key   = "key"
			value = "value"
		)

		BeforeEach(func() {
			testCache.Set(key, value, cache.DoesNotExpire)
		})

		It("should be available to get", func() {
			cacheMustHaveKeyAndValue(testCache, key, value)
		})

		It("should not call the function in get or set since it's not expired", func() {
			const otherValue = "other"
			fnCalled := false
			returnVal, err := testCache.GetOrSet(key, func(key string) (string, time.Duration, error) {
				fnCalled = true
				return otherValue, cache.DoesNotExpire, nil
			})
			Expect(fnCalled).To(BeFalse())
			Expect(err).To(Not(HaveOccurred()))
			Expect(returnVal).To(Equal(value))
			cacheMustHaveKeyAndValue(testCache, key, value)
		})

		It("should be able to be overwritten by set", func() {
			const newValue = "newValue"
			cacheMustHaveKeyAndValue(testCache, key, value)
			testCache.Set(key, newValue, cache.DoesNotExpire)
			cacheMustHaveKeyAndValue(testCache, key, newValue)
		})

		It("should be able to have another value with set", func() {
			const newKey = "newKey"
			const newValue = "newValue"
			testCache.Set(newKey, newValue, cache.DoesNotExpire)
			cacheMustHaveKeyAndValue(testCache, key, value)
			cacheMustHaveKeyAndValue(testCache, newKey, newValue)
		})

		It("should be available to be removed", func() {
			cacheMustHaveKeyAndValue(testCache, key, value)
			testCache.Remove(key)
			_, gotten := testCache.Get(key)
			Expect(gotten).To(BeFalse())
		})
	})

	When("a cache item expires", func() {
		const (
			key   = "key"
			value = "value"
		)

		BeforeEach(func() {
			testCache.Set(key, value, time.Nanosecond)
		})

		It("should not be available to get", func() {
			time.Sleep(time.Millisecond)
			_, gotten := testCache.Get(key)
			Expect(gotten).To(BeFalse())
		})

		It("should call the function in get or set since it's expired", func() {
			const otherValue = "other"
			fnCalled := false
			returnVal, err := testCache.GetOrSet(key, func(key string) (string, time.Duration, error) {
				fnCalled = true
				return otherValue, cache.DoesNotExpire, nil
			})
			Expect(fnCalled).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Expect(returnVal).To(Equal(otherValue))
			cacheMustHaveKeyAndValue(testCache, key, otherValue)
		})
	})
})
