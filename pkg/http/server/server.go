package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"

	"intelligence/pkg/config"
	"intelligence/pkg/http/api"
	"intelligence/pkg/http/middleware"
	"intelligence/pkg/logger"
	netutils "intelligence/pkg/utils/net"
)

// Server handles requests via the Hypertext Transfer Protocol (HTTP) and sends back responses.
// The Server must be allocated using New since the zero value for Server is not valid configuration.
type Server struct {
	conf     config.Server
	listener net.Listener
	srv      *http.Server
	done     chan bool
}

// New allocates and sets the required configuration for a Server.
func New(conf config.Server) *Server {
	return &Server{
		conf: conf,
		srv:  nil,
		done: nil,
	}
}

// Run configures and starts an HTTP server. When the server is bound to its IP and port, it invokes the callback function.
func (server *Server) Run(ctx context.Context, commonMiddleware []middleware.Middleware, endpointHandlers []api.HTTPEndpointHandler, readyCallback func()) error {
	logEntry := logger.LogEntry(ctx)

	// Create the done channel.
	server.done = make(chan bool)
	defer close(server.done)

	// Build the endpoints.
	builder := api.NewHTTPAPIBuilder()
	for _, endpointHandler := range endpointHandlers {
		endpointHandler.AcceptHTTPAPIBuilder(builder)
	}

	// Attach the handler routes to a serve mux.
	serveMux := http.NewServeMux()
	for apiPath, methodToEndpointHandlerMap := range builder.Handlers() {
		for method, endpointHandler := range methodToEndpointHandlerMap {
			allMiddleware := append(commonMiddleware, endpointHandler.Middleware...)
			handlerChain := middleware.CreateChain(allMiddleware, endpointHandler.Handler)
			logEntry.Debugf("Path %s has a handler for %s and %d middleware.", apiPath, method, len(allMiddleware))
			serveMux.HandleFunc(fmt.Sprintf("%s %s", method.String(), apiPath.String()), handlerChain)
		}
	}

	// Create the address to bind to.
	addr, err := netutils.FormatNetworkAddress(server.conf.ServerBindIP, server.conf.ServerBindPort)
	if err != nil {
		return err
	}
	logEntry.Debugf("The server bind address is %s.", addr)

	// Configure TLS.
	serverCert, err := tls.LoadX509KeyPair(server.conf.ServerCert, server.conf.ServerKey)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	// Configure the server.
	server.srv = &http.Server{
		Addr:              addr,
		Handler:           serveMux,
		ReadTimeout:       server.conf.ServerReadTimeout,
		WriteTimeout:      server.conf.ServerWriteTimeout,
		IdleTimeout:       0,       // Turn off the keep-alive functionality.
		ReadHeaderTimeout: 0,       // Uses the value of the read timeout.
		MaxHeaderBytes:    1 << 20, // 1MB.
		ErrorLog:          log.New(logger.LogEntry(ctx).WriterLevel(logrus.ErrorLevel), "", 0),
		TLSConfig:         tlsConfig,
	}

	// Manually creating the listener first ensures the server can start receiving connections before
	// it is marked as ready by the callback.
	server.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// Let the caller know the server is bound, and ready to serve HTTP.
	readyCallback()

	// Run the server.
	err = server.srv.ServeTLS(server.listener, "", "")
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server and waits for it to finish.
// This function can only be called once.
func (server *Server) Shutdown(ctx context.Context) error {
	logEntry := logger.LogEntry(ctx)
	logEntry.Debugf("Shutting down server.")
	err := server.srv.Shutdown(ctx)
	_ = server.listener.Close()
	<-server.done
	return err
}
