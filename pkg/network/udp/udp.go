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

package udp

import (
	"fmt"
	"net"

	"intelligence/pkg/network"
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
