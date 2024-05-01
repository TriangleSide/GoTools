package ptr_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/utils/ptr"
)

var _ = Describe("to pointer utility", func() {
	It("should convert a uint to a pointer", func() {
		ptrVal := ptr.Of[uint](123)
		Expect(ptrVal).To(Not(BeNil()))
		Expect(reflect.TypeOf(ptrVal).Elem().Kind()).To(Equal(reflect.Uint))
		Expect(*ptrVal).To(Equal(uint(123)))
	})

	It("should convert a float32 to a pointer", func() {
		ptrVal := ptr.Of[float32](123.45)
		Expect(ptrVal).To(Not(BeNil()))
		Expect(reflect.TypeOf(ptrVal).Elem().Kind()).To(Equal(reflect.Float32))
		Expect(*ptrVal).To(BeNumerically("~", float32(123.45), 0.001))
	})

	It("should convert a struct to a pointer", func() {
		type testStruct struct {
			Value int
		}
		ptrVal := ptr.Of[testStruct](testStruct{Value: 123})
		Expect(ptrVal).To(Not(BeNil()))
		Expect(reflect.TypeOf(ptrVal).Elem().Kind()).To(Equal(reflect.Struct))
		Expect(ptrVal.Value).To(Equal(123))
	})

	It("should panic with a pointer type", func() {
		Expect(func() {
			ptr.Of[*int](nil)
		}).Should(Panic())
	})
})
