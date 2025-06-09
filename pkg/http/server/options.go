package server

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/http/api"
	"github.com/TriangleSide/GoTools/pkg/http/middleware"
)

// serverOptions is configured by the caller with the Option functions.
type serverOptions struct {
	configProvider   func() (*Config, error)
	listenerProvider func(bindIP string, bindPort uint16) (*net.TCPListener, error)
	boundCallback    func(tcpAddr *net.TCPAddr)
	commonMiddleware []middleware.Middleware
	endpointHandlers []api.HTTPEndpointHandler
}

// configure applies the options to the default serverOptions values.
func configure(opts ...Option) *serverOptions {
	srvOpts := &serverOptions{
		configProvider: func() (*Config, error) {
			return config.ProcessAndValidate[Config]()
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

// Option is used to configure the HTTP server.
type Option func(srvOpts *serverOptions)

// WithConfigProvider sets the provider for the config.Config.
func WithConfigProvider(provider func() (*Config, error)) Option {
	return func(srvOpts *serverOptions) {
		srvOpts.configProvider = provider
	}
}

// WithListenerProvider sets the provider for the tcp.Listener.
func WithListenerProvider(provider func(bindIP string, bindPort uint16) (*net.TCPListener, error)) Option {
	return func(srvOpts *serverOptions) {
		srvOpts.listenerProvider = provider
	}
}

// WithBoundCallback sets the bound callback for the server.
// The callback is invoked when the network listener is bound to the configured IP and port.
func WithBoundCallback(callback func(tcpAddr *net.TCPAddr)) Option {
	return func(srvOpts *serverOptions) {
		srvOpts.boundCallback = callback
	}
}

// WithCommonMiddleware adds common middleware for the server.
// The middleware gets executed on every request to the server.
func WithCommonMiddleware(commonMiddleware ...middleware.Middleware) Option {
	return func(srvOpts *serverOptions) {
		srvOpts.commonMiddleware = append(srvOpts.commonMiddleware, commonMiddleware...)
	}
}

// WithEndpointHandlers adds the handlers to the server.
func WithEndpointHandlers(endpointHandlers ...api.HTTPEndpointHandler) Option {
	return func(srvOpts *serverOptions) {
		srvOpts.endpointHandlers = append(srvOpts.endpointHandlers, endpointHandlers...)
	}
}
