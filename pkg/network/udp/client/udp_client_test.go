package udp_client_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	udpclient "intelligence/pkg/network/udp/client"
)

var _ = Describe("udp client", func() {
	When("the udp client host is an incorrectly formatted IP", func() {
		It("should return an error", func() {
			conn, err := udpclient.New("300.300.300.300", 13579)
			Expect(conn).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to format the UDP address (invalid hostname '300.300.300.300')"))
		})
	})

	When("configuring the udp client fails", func() {
		It("should return an error", func() {
			conn, err := udpclient.New("::1", 13579, func(config *udpclient.Config) error {
				return errors.New("failed")
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to configure UDP client (failed)"))
			Expect(conn).To(BeNil())
		})
	})

	When("a udp client is bound to the same local port twice", func() {
		It("should return an error", func() {
			localAddressConfig := udpclient.WithLocalAddress("::1", 7654)
			conn, err := udpclient.New("::1", 13579, localAddressConfig)
			Expect(err).To(Not(HaveOccurred()))
			Expect(conn).To(Not(BeNil()))
			secondConn, err := udpclient.New("::1", 13579, localAddressConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("address already in use"))
			Expect(secondConn).To(BeNil())
			Expect(conn.Close()).To(Succeed())
		})
	})
})
