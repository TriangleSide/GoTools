package network_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/network"
)

var _ = Describe("format network address", func() {
	When("the port 12345 is used", func() {
		const (
			port uint16 = 12345
		)

		When("an IPv4 address is used", func() {
			It("should return a formatted address", func() {
				formatted, err := network.FormatNetworkAddress("127.0.0.1", port)
				Expect(err).ToNot(HaveOccurred())
				Expect(formatted).To(Equal("127.0.0.1:12345"))
			})
		})

		When("an IPv6 address  is used", func() {
			It("should return a formatted address", func() {
				formatted, err := network.FormatNetworkAddress("::1", port)
				Expect(err).ToNot(HaveOccurred())
				Expect(formatted).To(Equal("[::1]:12345"))
			})
		})

		When("the hostname localhost is used", func() {
			It("should return a formatted address", func() {
				formatted, err := network.FormatNetworkAddress("localhost", port)
				Expect(err).ToNot(HaveOccurred())
				Expect(formatted).To(Equal("localhost:12345"))
			})
		})

		When("the fqdn example.com is used", func() {
			It("should return a formatted address", func() {
				formatted, err := network.FormatNetworkAddress("example.com", port)
				Expect(err).ToNot(HaveOccurred())
				Expect(formatted).To(Equal("example.com:12345"))
			})
		})

		When("an incorrectly formatted hostname is used", func() {
			It("should fail", func() {
				formatted, err := network.FormatNetworkAddress("[=+--]", port)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid hostname"))
				Expect(formatted).To(BeEmpty())
			})
		})

		When("an incorrectly formatted IP is used", func() {
			It("should fail", func() {
				formatted, err := network.FormatNetworkAddress("256.100.50.25", port)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid hostname '256.100.50.25'"))
				Expect(formatted).To(BeEmpty())
			})
		})
	})
})
