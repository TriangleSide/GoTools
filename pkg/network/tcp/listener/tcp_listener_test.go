package tcp_listener_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tcplistener "github.com/TriangleSide/GoBase/pkg/network/tcp/listener"
)

var _ = Describe("tcp listener ", func() {
	When("the tcp listener host is an incorrectly formatted IP", func() {
		It("should return an error", func() {
			conn, err := tcplistener.New("300.300.300.300", 13579)
			Expect(conn).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to format the TCP address (invalid hostname '300.300.300.300')"))
		})
	})

	When("a tcp listener is bound to the same port twice", func() {
		It("should return an error", func() {
			conn, err := tcplistener.New("::1", 13579)
			Expect(err).To(Not(HaveOccurred()))
			Expect(conn).To(Not(BeNil()))
			secondConn, err := tcplistener.New("::1", 13579)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to listen on the TCP address"))
			Expect(secondConn).To(BeNil())
			Expect(conn.Close()).To(Succeed())
		})
	})
})
