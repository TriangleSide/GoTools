package udp_listener_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	udplistener "intelligence/pkg/network/udp/listener"
)

var _ = Describe("udp listener", func() {
	When("the udp listener host is an incorrectly formatted IP", func() {
		It("should return an error", func() {
			conn, err := udplistener.New("300.300.300.300", 13579)
			Expect(conn).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to format the UDP address (invalid hostname '300.300.300.300')"))
		})
	})

	When("a udp listener is bound to the same port twice", func() {
		It("should return an error", func() {
			conn, err := udplistener.New("::1", 13579)
			Expect(err).To(Not(HaveOccurred()))
			Expect(conn).To(Not(BeNil()))
			secondConn, err := udplistener.New("::1", 13579)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to listen on the address"))
			Expect(secondConn).To(BeNil())
			Expect(conn.Close()).To(Succeed())
		})
	})
})
