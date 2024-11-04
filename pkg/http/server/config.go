package server

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/TriangleSide/GoTools/pkg/config"
)

// TLSMode represents the TLS mode of the HTTP server.
type TLSMode string

const (
	// TLSModeOff represents plain HTTP without TLS.
	TLSModeOff TLSMode = "off"

	// TLSModeTLS represents HTTP over TLS.
	TLSModeTLS TLSMode = "tls"

	// TLSModeMutualTLS represents HTTP over mutual TLS.
	TLSModeMutualTLS TLSMode = "mutual_tls"
)

const (
	ConfigPrefix = "HTTP_SERVER"
)

// Config holds configuration parameters for an HTTP server.
type Config struct {
	// BindIP is the IP address the server listens on.
	BindIP string `config_format:"snake" config_default:"::1" validate:"required,ip_addr"`

	// BindPort is the port number the server listens on.
	BindPort uint16 `config_format:"snake" config_default:"0" validate:"gte=0"`

	// ReadTimeoutMilliseconds is the maximum time (in seconds) to read the request.
	// Zero or negative means no timeout.
	ReadTimeoutMilliseconds int `config_format:"snake" config_default:"120000" validate:"gte=0"`

	// WriteTimeoutMilliseconds is the maximum time (in seconds) to write the response.
	// Zero or negative means no timeout.
	WriteTimeoutMilliseconds int `config_format:"snake" config_default:"120000" validate:"gte=0"`

	// IdleTimeoutMilliseconds sets the max idle time (in seconds) between requests when keep-alives are enabled.
	// If zero, ReadTimeout is used. If both are zero, it means no timeout.
	IdleTimeoutMilliseconds int `config_format:"snake" config_default:"0" validate:"gte=0"`

	// HeaderReadTimeoutMilliseconds is the maximum time (in seconds) to read request headers.
	// If zero, ReadTimeout is used. If both are zero, it means no timeout.
	HeaderReadTimeoutMilliseconds int `config_format:"snake" config_default:"0" validate:"gte=0"`

	// TLSMode specifies the TLS mode of the server: off, tls, or mutual_tls.
	TLSMode TLSMode `config_format:"snake" config_default:"tls" validate:"oneof=off tls mutual_tls"`

	// Cert is the path to the TLS certificate file.
	Cert string `config_format:"snake" config_default:"" validate:"required_if=TLSMode tls,required_if=TLSMode mutual_tls,omitempty,filepath"`

	// Key is the path to the TLS private key file.
	Key string `config_format:"snake" config_default:"" validate:"required_if=TLSMode tls,required_if=TLSMode mutual_tls,omitempty,filepath"`

	// ClientCACerts is a list of paths to client CA certificate files (used in mutual TLS).
	ClientCACerts []string `config_format:"snake" config_default:"[]" validate:"required_if=TLSMode mutual_tls,dive,required,filepath"`

	// MaxHeaderBytes sets the maximum size in bytes of request headers. It doesn't limit the request body size.
	MaxHeaderBytes int `config_format:"snake" config_default:"1048576" validate:"gte=4096,lte=1073741824"`

	// KeepAlive controls whether HTTP keep-alives are enabled. By default, keep-alives are always enabled.
	KeepAlive bool `config_format:"snake" config_default:"true"`
}

// configure applies the options to the default serverOptions values.
func configure(opts ...Option) *serverOptions {
	srvOpts := &serverOptions{
		configProvider: func() (*Config, error) {
			return config.ProcessAndValidate[Config](config.WithPrefix(ConfigPrefix))
		},
		listenerProvider: func(bindIp string, bindPort uint16) (*net.TCPListener, error) {
			ip, err := netip.ParseAddr(bindIp)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bind IP: %w", err)
			}
			addrPort := netip.AddrPortFrom(ip, bindPort)
			tcpAddr := net.TCPAddrFromAddrPort(addrPort)
			return net.ListenTCP(tcpAddr.Network(), tcpAddr)
		},
	}

	for _, opt := range opts {
		opt(srvOpts)
	}

	return srvOpts
}
