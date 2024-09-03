package udp

import (
	"fmt"
	"net"

	"github.com/TriangleSide/GoBase/pkg/network"
)

// Conn exists because the GoLang standard library doesn't have an interface for it.
type Conn interface {
	net.Conn
	ReadFromUDP(b []byte) (int, *net.UDPAddr, error)
	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	Close() error
	LocalAddr() net.Addr
	SetReadBuffer(bytes int) error
	SetWriteBuffer(bytes int) error
}

// ResolveAddr validates and formats a net.UDPAddr.
func ResolveAddr(host string, port uint16) (*net.UDPAddr, error) {
	formattedAddress, err := network.FormatNetworkAddress(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to format the UDP address (%s)", err.Error())
	}

	return net.ResolveUDPAddr("udp", formattedAddress)
}
