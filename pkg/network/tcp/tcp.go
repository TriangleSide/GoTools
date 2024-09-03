package tcp

import (
	"fmt"
	"net"

	"github.com/TriangleSide/GoBase/pkg/network"
)

// Listener exists because the GoLang standard library doesn't have an interface for it.
type Listener interface {
	net.Listener
}

// ResolveAddr validates and formats a net.TCPAddr.
func ResolveAddr(host string, port uint16) (*net.TCPAddr, error) {
	formattedAddress, err := network.FormatNetworkAddress(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to format the TCP address (%s)", err.Error())
	}

	return net.ResolveTCPAddr("tcp", formattedAddress)
}
