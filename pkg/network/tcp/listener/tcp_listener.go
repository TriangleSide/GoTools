package tcp_listener

import (
	"fmt"
	"net"

	"intelligence/pkg/network/tcp"
)

// New creates a TCP listener with some default settings.
func New(localHost string, localPort uint16) (tcp.Listener, error) {
	resolvedAddress, err := tcp.ResolveAddr(localHost, localPort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the TCP address (%s)", err.Error())
	}

	conn, err := net.ListenTCP("tcp", resolvedAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on the TCP address (%s)", err.Error())
	}

	return conn, nil
}
