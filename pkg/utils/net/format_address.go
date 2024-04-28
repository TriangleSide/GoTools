package net

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

	type HostnameValidator struct {
		Hostname string `validate:"required,hostname"`
	}
	hostnameValidator := HostnameValidator{
		Hostname: host,
	}
	err := validation.Validate(hostnameValidator)
	if err != nil {
		return "", fmt.Errorf("invalid hostname %s", host)
	}
	return fmt.Sprintf("%s:%d", host, port), nil
}
