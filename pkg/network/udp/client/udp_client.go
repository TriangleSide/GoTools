// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

package udp_client

import (
	"fmt"
	"net"

	"intelligence/pkg/network/udp"
)

// Config is configured by the caller with the Option functions.
type Config struct {
	localAddress *net.UDPAddr
}

// Option is used to configure the UDP client.
type Option func(*Config) error

// WithLocalAddress makes the UDP client connect to a specific local host and port.
func WithLocalAddress(localHost string, localPort uint16) Option {
	return func(config *Config) error {
		resolvedAddress, err := udp.ResolveAddr(localHost, localPort)
		config.localAddress = resolvedAddress
		return err
	}
}

// New dials a remote UDP address.
func New(remoteHost string, remotePort uint16, opts ...Option) (udp.Conn, error) {
	config := &Config{
		localAddress: nil,
	}
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, fmt.Errorf("failed to configure UDP client (%s)", err.Error())
		}
	}

	resolvedAddress, err := udp.ResolveAddr(remoteHost, remotePort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the UDP address (%s)", err.Error())
	}

	conn, err := net.DialUDP("udp", config.localAddress, resolvedAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to dial the UDP address (%s)", err.Error())
	}

	return conn, nil
}
