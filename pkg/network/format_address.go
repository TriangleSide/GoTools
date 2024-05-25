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

package network

import (
	"fmt"
	"net"
	"strings"

	"intelligence/pkg/validation"
)

// FormatNetworkAddress takes a host address and a port number, and returns a correctly formatted server address.
func FormatNetworkAddress(host string, port uint16) (string, error) {
	parsedIP := net.ParseIP(host)
	if parsedIP != nil {
		if strings.Contains(host, ":") {
			return fmt.Sprintf("[%s]:%d", host, port), nil
		} else {
			return fmt.Sprintf("%s:%d", host, port), nil
		}
	}

	if err := validation.Var(&host, "required,hostname"); err != nil {
		return "", fmt.Errorf("invalid hostname '%s'", host)
	}

	return fmt.Sprintf("%s:%d", host, port), nil
}
