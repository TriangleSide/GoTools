// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

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
		}).Should(PanicWith(ContainSubstring("type cannot be a pointer")))
	})
})
