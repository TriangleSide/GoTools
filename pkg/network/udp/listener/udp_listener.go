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

package udp_listener

import (
	"fmt"
	"net"

	"intelligence/pkg/network/udp"
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
