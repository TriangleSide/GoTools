package udp_listener

import (
	"fmt"
	"net"

	"github.com/TriangleSide/GoBase/pkg/network/udp"
)

// New creates a local UDP listener with some default settings.
func New(localHost string, localPort uint16) (udp.Conn, error) {
	resolvedAddress, err := udp.ResolveAddr(localHost, localPort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the address (%s)", err.Error())
	}

	conn, err := net.ListenUDP("udp", resolvedAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on the address (%s)", err.Error())
	}

	return conn, nil
}
