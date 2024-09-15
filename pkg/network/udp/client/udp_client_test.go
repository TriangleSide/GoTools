package udp_client_test

import (
	"github.com/TriangleSide/GoBase/pkg/network/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	udpclient "github.com/TriangleSide/GoBase/pkg/network/udp/client"
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

	When("a udp client is bound to the same local port twice", func() {
		It("should return an error", func() {
			localAddress, err := udp.ResolveAddr("::1", 7654)
			Expect(err).To(Not(HaveOccurred()))
			localAddressConfig := udpclient.WithLocalAddress(localAddress)
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
