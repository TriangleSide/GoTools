package udp_client

import (
	"fmt"
	"net"

	"github.com/TriangleSide/GoBase/pkg/network/udp"
)

// config is configured by the caller with the Option functions.
type config struct {
	localAddress *net.UDPAddr
}

// Option is used to configure the UDP client.
type Option func(*config)

// WithLocalAddress makes the UDP client connect to a specific local host and port.
func WithLocalAddress(localAddress *net.UDPAddr) Option {
	return func(cfg *config) {
		cfg.localAddress = localAddress
	}
}

// New dials a remote UDP address.
func New(remoteHost string, remotePort uint16, opts ...Option) (udp.Conn, error) {
	cfg := &config{
		localAddress: nil,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	resolvedAddress, err := udp.ResolveAddr(remoteHost, remotePort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the UDP address (%s)", err.Error())
	}

	conn, err := net.DialUDP("udp", cfg.localAddress, resolvedAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to dial the UDP address (%s)", err.Error())
	}

	return conn, nil
}
