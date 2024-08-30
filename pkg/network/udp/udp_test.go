package udp_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/network/udp"
)

var _ = Describe("udp", func() {
	When("the udp listener host is an incorrectly formatted IP", func() {
		It("should return an error", func() {
			conn, err := udp.ResolveAddr("300.300.300.300", 13579)
			Expect(conn).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to format the UDP address (invalid hostname '300.300.300.300')"))
		})
	})

	When("the udp listener host is an incorrectly formatted hostname", func() {
		It("should return an error", func() {
			conn, err := udp.ResolveAddr("doesnotexist.doesnotexist", 13579)
			Expect(conn).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no such host"))
		})
	})
})
