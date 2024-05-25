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

package tcp

import (
	"fmt"
	"net"

	"intelligence/pkg/network"
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
