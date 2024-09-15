package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TriangleSide/GoBase/pkg/config"
	"github.com/TriangleSide/GoBase/pkg/config/envprocessor"
	"github.com/TriangleSide/GoBase/pkg/http/api"
	"github.com/TriangleSide/GoBase/pkg/http/middleware"
	"github.com/TriangleSide/GoBase/pkg/network/tcp"
	tcplistener "github.com/TriangleSide/GoBase/pkg/network/tcp/listener"
)

// serverConfig is configured by the caller with the Option functions.
type serverConfig struct {
	configProvider   func() (*config.HTTPServer, error)
	listenerProvider func(localHost string, localPort uint16) (tcp.Listener, error)
}

// Option is used to configure the HTTP server.
type Option func(config *serverConfig)

// WithConfigProvider sets the provider for the config.HTTPServer.
func WithConfigProvider(provider func() (*config.HTTPServer, error)) Option {
	return func(config *serverConfig) {
		config.configProvider = provider
	}
}

// WithListenerProvider sets the provider for the tcp.Listener.
func WithListenerProvider(provider func(localHost string, localPort uint16) (tcp.Listener, error)) Option {
	return func(config *serverConfig) {
		config.listenerProvider = provider
	}
}

// Server handles requests via the Hypertext Transfer Protocol (HTTP) and sends back responses.
// The Server must be allocated using New since the zero value for Server is not valid configuration.
type Server struct {
	cfg      *serverConfig
	envConf  *config.HTTPServer
	listener tcp.Listener
	srv      *http.Server
	ran      *atomic.Bool
	shutdown *atomic.Bool
	wg       sync.WaitGroup
}

// New allocates and sets the required configuration for a Server.
func New(opts ...Option) (*Server, error) {
	cfg := &serverConfig{
		configProvider: func() (*config.HTTPServer, error) {
			return envprocessor.ProcessAndValidate[config.HTTPServer]()
		},
		listenerProvider: tcplistener.New,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	envConfig, err := cfg.configProvider()
	if err != nil {
		return nil, fmt.Errorf("could not load configuration (%s)", err.Error())
	}

	ran := &atomic.Bool{}
	ran.Store(false)

	shutdown := &atomic.Bool{}
	shutdown.Store(false)

	return &Server{
		cfg:      cfg,
		envConf:  envConfig,
		srv:      nil,
		ran:      ran,
		shutdown: shutdown,
		wg:       sync.WaitGroup{},
	}, nil
}

// Run configures and starts an HTTP server.
// After the HTTP server is bound to its IP and port, it invokes the callback function.
// This function blocks as long as it is serving HTTP.
func (server *Server) Run(commonMiddleware []middleware.Middleware, endpointHandlers []api.HTTPEndpointHandler, readyCallback func()) error {
	if server.ran.Swap(true) {
		panic("HTTP server can only be run once per instance")
	}

	server.wg.Add(1)
	defer func() { server.wg.Done() }()

	builder := api.NewHTTPAPIBuilder()
	for _, endpointHandler := range endpointHandlers {
		endpointHandler.AcceptHTTPAPIBuilder(builder)
	}

	serveMux := http.NewServeMux()
	for apiPath, methodToEndpointHandlerMap := range builder.Handlers() {
		for method, endpointHandler := range methodToEndpointHandlerMap {
			allMiddleware := append(commonMiddleware, endpointHandler.Middleware...)
			handlerChain := middleware.CreateChain(allMiddleware, endpointHandler.Handler)
			serveMux.HandleFunc(fmt.Sprintf("%s %s", method, apiPath), handlerChain)
		}
	}

	var tlsConfig *tls.Config = nil
	if server.envConf.HTTPServerTLS {
		serverCert, err := tls.LoadX509KeyPair(server.envConf.HTTPServerCert, server.envConf.HTTPServerKey)
		if err != nil {
			return fmt.Errorf("failed to load the server certificates (%s)", err.Error())
		}
		tlsConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{serverCert},
		}
	}

	server.srv = &http.Server{
		Handler:           serveMux,
		ReadTimeout:       time.Second * time.Duration(server.envConf.HTTPServerReadTimeoutSeconds),
		WriteTimeout:      time.Second * time.Duration(server.envConf.HTTPServerWriteTimeoutSeconds),
		IdleTimeout:       time.Second * time.Duration(server.envConf.HTTPServerIdleTimeoutSeconds),
		ReadHeaderTimeout: time.Second * time.Duration(server.envConf.HTTPServerHeaderReadTimeoutSeconds),
		MaxHeaderBytes:    server.envConf.HTTPServerMaxHeaderBytes,
		TLSConfig:         tlsConfig,
	}

	// Manually creating the listener first ensures the server can start receiving connections before
	// it is marked as ready by the callback.
	var err error
	server.listener, err = server.cfg.listenerProvider(server.envConf.HTTPServerBindIP, server.envConf.HTTPServerBindPort)
	if err != nil {
		return fmt.Errorf("failed to create the network listener (%s)", err.Error())
	}

	readyCallback()

	// ServeTLS always returns as error as it is meant to be blocking.
	if server.envConf.HTTPServerTLS {
		err = server.srv.ServeTLS(server.listener, "", "")
	} else {
		err = server.srv.Serve(server.listener)
	}

	// The server blocks until there is an error, or it is shutdown.
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return fmt.Errorf("error encountered while serving http requests (%s)", err.Error())
	}
}

// Shutdown gracefully shuts down the server and waits for it to finish.
// This function can be called concurrently, but the first will perform the shutdown action.
func (server *Server) Shutdown(ctx context.Context) error {
	var err error
	if !server.shutdown.Swap(true) {
		err = server.srv.Shutdown(ctx)
		_ = server.listener.Close()
	}
	server.wg.Wait()
	return err
}
