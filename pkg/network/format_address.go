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
