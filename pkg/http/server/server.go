package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"intelligence/pkg/config"
	"intelligence/pkg/http/api"
	"intelligence/pkg/http/middleware"
	netutils "intelligence/pkg/utils/net"
)

// Server handles requests via the Hypertext Transfer Protocol (HTTP) and sends back responses.
// The Server must be allocated using New since the zero value for Server is not valid configuration.
type Server struct {
	conf     config.Server
	listener net.Listener
	srv      *http.Server
	ran      *atomic.Bool
	shutdown *atomic.Bool
	wg       sync.WaitGroup
}

// New allocates and sets the required configuration for a Server.
func New(conf config.Server) *Server {
	ran := &atomic.Bool{}
	ran.Store(false)

	shutdown := &atomic.Bool{}
	shutdown.Store(false)

	return &Server{
		conf:     conf,
		srv:      nil,
		ran:      ran,
		shutdown: shutdown,
		wg:       sync.WaitGroup{},
	}
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
			serveMux.HandleFunc(fmt.Sprintf("%s %s", method.String(), apiPath.String()), handlerChain)
		}
	}

	addr, err := netutils.FormatNetworkAddress(server.conf.ServerBindIP, server.conf.ServerBindPort)
	if err != nil {
		return fmt.Errorf("failed to format the server network address (%s)", err.Error())
	}

	serverCert, err := tls.LoadX509KeyPair(server.conf.ServerCert, server.conf.ServerKey)
	if err != nil {
		return fmt.Errorf("failed to load the server certificate (%s)", err.Error())
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	server.srv = &http.Server{
		Addr:              addr,
		Handler:           serveMux,
		ReadTimeout:       server.conf.ServerReadTimeout,
		WriteTimeout:      server.conf.ServerWriteTimeout,
		IdleTimeout:       0,       // Turn off the keep-alive functionality.
		ReadHeaderTimeout: 0,       // Uses the value of the read timeout.
		MaxHeaderBytes:    1 << 20, // 1MB.
		TLSConfig:         tlsConfig,
	}

	// Manually creating the listener first ensures the server can start receiving connections before
	// it is marked as ready by the callback.
	server.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen to tcp address (%s)", err.Error())
	}

	readyCallback()

	err = server.srv.ServeTLS(server.listener, "", "")
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("error encountered while serving http requests (%s)", err.Error())
	}

	return nil
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
