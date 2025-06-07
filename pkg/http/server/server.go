package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

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

// Server handles requests via the Hypertext Transfer Protocol (HTTP) and sends back responses.
// The Server must be allocated using New since the zero value for Server is not valid configuration.
type Server struct {
	srv              http.Server
	ran              atomic.Bool
	shutdown         atomic.Bool
	wg               sync.WaitGroup
	listenerProvider func() (*net.TCPListener, error)
	boundCallback    func(tcpAddr *net.TCPAddr)
}

// New configures an HTTP server with the provided options.
func New(opts ...Option) (*Server, error) {
	srvOpts := configure(opts...)

	envConfig, err := srvOpts.configProvider()
	if err != nil {
		return nil, fmt.Errorf("could not load configuration (%w)", err)
	}

	builder := api.NewHTTPAPIBuilder()
	for _, endpointHandler := range srvOpts.endpointHandlers {
		endpointHandler.AcceptHTTPAPIBuilder(builder)
	}

	serveMux := http.NewServeMux()
	for apiPath, methodToEndpointHandlerMap := range builder.Handlers() {
		for method, endpointHandler := range methodToEndpointHandlerMap {
			endpointHandlerMw := append(srvOpts.commonMiddleware, endpointHandler.Middleware...)
			handlerChain := middleware.CreateChain(endpointHandlerMw, endpointHandler.Handler)
			serveMux.HandleFunc(fmt.Sprintf("%s %s", method, apiPath), handlerChain)
		}
	}

	var tlsConfig *tls.Config
	switch envConfig.HTTPServerTLSMode {
	case TLSModeOff:
		tlsConfig = nil
	case TLSModeTLS:
		serverCert, err := tls.LoadX509KeyPair(envConfig.HTTPServerCert, envConfig.HTTPServerKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load the server certificates (%w)", err)
		}
		tlsConfig = &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{serverCert},
		}
	case TLSModeMutualTLS:
		if len(envConfig.HTTPServerClientCACerts) == 0 {
			return nil, errors.New("no client CAs provided")
		}
		serverCert, err := tls.LoadX509KeyPair(envConfig.HTTPServerCert, envConfig.HTTPServerKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load the server certificates (%w)", err)
		}
		clientCAs, err := loadMutualTLSClientCAs(envConfig.HTTPServerClientCACerts)
		if err != nil {
			return nil, fmt.Errorf("failed to load client CA certificates (%w)", err)
		}
		tlsConfig = &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCAs,
		}
	default:
		return nil, fmt.Errorf("invalid TLS mode: %s", envConfig.HTTPServerTLSMode)
	}

	srv := &Server{
		srv: http.Server{
			Handler:           serveMux,
			ReadTimeout:       time.Millisecond * time.Duration(envConfig.HTTPServerReadTimeoutMillis),
			WriteTimeout:      time.Millisecond * time.Duration(envConfig.HTTPServerWriteTimeoutMillis),
			IdleTimeout:       time.Millisecond * time.Duration(envConfig.HTTPServerIdleTimeoutMillis),
			ReadHeaderTimeout: time.Millisecond * time.Duration(envConfig.HTTPServerHeaderReadTimeoutMillis),
			MaxHeaderBytes:    envConfig.HTTPServerMaxHeaderBytes,
			TLSConfig:         tlsConfig,
		},
		ran:      atomic.Bool{},
		shutdown: atomic.Bool{},
		wg:       sync.WaitGroup{},
		listenerProvider: func() (*net.TCPListener, error) {
			return srvOpts.listenerProvider(envConfig.HTTPServerBindIP, envConfig.HTTPServerBindPort)
		},
		boundCallback: srvOpts.boundCallback,
	}

	srv.srv.SetKeepAlivesEnabled(envConfig.HTTPServerKeepAlive)
	srv.ran.Store(false)
	srv.shutdown.Store(false)

	return srv, nil
}

// Run starts an HTTP server.
// This function blocks as long as its serving HTTP requests.
func (server *Server) Run() error {
	if server.ran.Swap(true) {
		panic("HTTP server can only be run once per instance")
	}
	server.wg.Add(1)
	defer func() { server.wg.Done() }()

	listener, err := server.listenerProvider()
	if err != nil {
		return fmt.Errorf("failed to create the network listener (%w)", err)
	}

	if server.boundCallback != nil {
		tcpAddr := listener.Addr().(*net.TCPAddr)
		server.boundCallback(tcpAddr)
	}

	if server.srv.TLSConfig == nil {
		err = server.srv.Serve(listener)
	} else {
		err = server.srv.ServeTLS(listener, "", "")
	}

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return fmt.Errorf("error encountered while serving http requests (%w)", err)
	}
}

// Shutdown gracefully shuts down the server and waits for it to finish.
// This function can be called concurrently, but the first will perform the shutdown action.
func (server *Server) Shutdown(ctx context.Context) error {
	var err error
	if !server.shutdown.Swap(true) {
		err = server.srv.Shutdown(ctx)
	}
	server.wg.Wait()
	return err
}

// loadMutualTLSClientCAs loads client CA certificates for mutual TLS.
func loadMutualTLSClientCAs(clientCaCertPaths []string) (*x509.CertPool, error) {
	clientCAs := x509.NewCertPool()
	for _, caCertPath := range clientCaCertPaths {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("could not read client CA certificate on path %s (%w)", caCertPath, err)
		}
		if ok := clientCAs.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append client CA certificate (%s)", caCertPath)
		}
	}
	return clientCAs, nil
}
