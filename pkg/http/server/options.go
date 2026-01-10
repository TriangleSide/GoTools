package server

import (
	"fmt"
	"net"
	"strconv"

	"github.com/TriangleSide/go-toolkit/pkg/config"
	"github.com/TriangleSide/go-toolkit/pkg/http/endpoints"
	"github.com/TriangleSide/go-toolkit/pkg/http/middleware"
)

// serverOptions is configured by the caller with the Option functions.
type serverOptions struct {
	configProvider   func() (*Config, error)
	listenerProvider func(bindIP string, bindPort uint16) (*net.TCPListener, error)
	boundCallback    func(tcpAddr *net.TCPAddr)
	commonMiddleware []middleware.Middleware
	endpointHandlers []endpoints.EndpointHandler
}

// configure applies the options to the default serverOptions values.
func configure(opts ...Option) *serverOptions {
	srvOpts := &serverOptions{
		configProvider: func() (*Config, error) {
			return config.ProcessAndValidate[Config]()
		},
		listenerProvider: func(bindIp string, bindPort uint16) (*net.TCPListener, error) {
			addr := net.JoinHostPort(bindIp, strconv.FormatUint(uint64(bindPort), 10))
			tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve TCP address: %w", err)
			}
			return net.ListenTCP("tcp", tcpAddr)
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

// WithEndpoints adds endpoint handlers to the server.
func WithEndpoints(endpointHandlers ...endpoints.EndpointHandler) Option {
	return func(srvOpts *serverOptions) {
		srvOpts.endpointHandlers = append(srvOpts.endpointHandlers, endpointHandlers...)
	}
}
