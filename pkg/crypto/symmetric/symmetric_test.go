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

package symmetric_test

import (
	"crypto/rand"
	mathrand "math/rand"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/crypto/symmetric"
)

var _ = Describe("symmetric encryption", func() {
	When("an encryptor is created", func() {
		var encryptor *symmetric.Encryptor

		BeforeEach(func() {
			var err error
			encryptor, err = symmetric.New("encryptionKey" + strconv.Itoa(mathrand.Int()))
			Expect(err).ToNot(HaveOccurred())
		})

		When("data of different size is generated", func() {
			It("should be able to be encrypted and decrypted", func() {
				for dataSize := 1; dataSize <= 1024; dataSize++ {
					data := make([]byte, dataSize)
					_, err := rand.Read(data)
					Expect(err).NotTo(HaveOccurred())
					encrypted, err := encryptor.Encrypt(data)
					Expect(err).NotTo(HaveOccurred())
					Expect(encrypted).To(Not(Equal(data)))
					decrypted, err := encryptor.Decrypt(encrypted)
					Expect(err).NotTo(HaveOccurred())
					Expect(decrypted).To(Equal(data))
				}
			})
		})

		When("the same data is encrypted", func() {
			It("should have different cypher-text", func() {
				data := []byte{0x00, 0x01, 0x02}
				encrypted1, err := encryptor.Encrypt(data)
				Expect(err).NotTo(HaveOccurred())
				encrypted2, err := encryptor.Encrypt(data)
				Expect(err).NotTo(HaveOccurred())
				Expect(encrypted1).To(Not(Equal(encrypted2)))
			})
		})

		When("nil bytes are decrypted", func() {
			It("should return an error", func() {
				decrypted, err := encryptor.Decrypt(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("shorter then the minimum length"))
				Expect(decrypted).To(BeNil())
			})
		})

		When("an empty slice of bytes are encrypted and decrypted", func() {
			It("should return an empty slice", func() {
				encrypted, err := encryptor.Encrypt([]byte{})
				Expect(err).NotTo(HaveOccurred())
				decrypted, err := encryptor.Decrypt(encrypted)
				Expect(err).NotTo(HaveOccurred())
				Expect(decrypted).To(Not(BeNil()))
				Expect(decrypted).To(BeEmpty())
			})
		})

		When("an nil bytes are encrypted and decrypted", func() {
			It("should return an empty slice", func() {
				encrypted, err := encryptor.Encrypt(nil)
				Expect(err).NotTo(HaveOccurred())
				decrypted, err := encryptor.Decrypt(encrypted)
				Expect(err).NotTo(HaveOccurred())
				Expect(decrypted).To(Not(BeNil()))
				Expect(decrypted).To(BeEmpty())
			})
		})
	})

	When("an encryptor is created with an empty key", func() {
		It("should return an error", func() {
			encryptor, err := symmetric.New("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid key"))
			Expect(encryptor).To(BeNil())
		})
	})
})
