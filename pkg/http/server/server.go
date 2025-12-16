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
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TriangleSide/GoTools/pkg/http/api"
	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/middleware"
)

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

	serveMux := configureServeMux(srvOpts)

	var tlsConfig *tls.Config
	if envConfig.HTTPServerTLSMode != TLSModeOff {
		tlsConfig, err = configureTLS(envConfig)
		if err != nil {
			return nil, err
		}
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
	}
	return fmt.Errorf("error encountered while serving http requests (%w)", err)
}

// Shutdown gracefully shuts down the server and waits for it to finish.
// This function can be called concurrently, but the first will perform the shutdown action.
func (server *Server) Shutdown(ctx context.Context) error {
	var err error
	if !server.shutdown.Swap(true) {
		if shutdownErr := server.srv.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("failed to shutdown the server (%w)", shutdownErr)
		}
	}
	server.wg.Wait()
	return err
}

// configureServeMux creates and configures an HTTP request multiplexer with endpoint handlers.
func configureServeMux(srvOpts *serverOptions) *http.ServeMux {
	builder := api.NewHTTPAPIBuilder()
	for _, endpointHandler := range srvOpts.endpointHandlers {
		endpointHandler.AcceptHTTPAPIBuilder(builder)
	}

	serveMux := http.NewServeMux()
	for apiPath, methodToEndpointHandlerMap := range builder.Handlers() {
		methodHandlers := make(map[string]http.HandlerFunc, len(methodToEndpointHandlerMap))
		allowedMethods := make([]string, 0, len(methodToEndpointHandlerMap))
		for method, endpointHandler := range methodToEndpointHandlerMap {
			endpointHandlerMw := make([]middleware.Middleware, 0, len(srvOpts.commonMiddleware)+len(endpointHandler.Middleware))
			endpointHandlerMw = append(endpointHandlerMw, srvOpts.commonMiddleware...)
			endpointHandlerMw = append(endpointHandlerMw, endpointHandler.Middleware...)
			handlerChain := middleware.CreateChain(endpointHandlerMw, endpointHandler.Handler)
			methodHandlers[string(method)] = handlerChain
			allowedMethods = append(allowedMethods, string(method))
		}
		sort.Strings(allowedMethods)
		path := string(apiPath)
		serveMux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
			handler, ok := methodHandlers[request.Method]
			if !ok {
				writer.Header().Set(headers.Allow, strings.Join(allowedMethods, ", "))
				writer.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			handler(writer, request)
		})
	}
	return serveMux
}

// configureTLS creates a TLS configuration based on the specified TLS mode.
func configureTLS(envConfig *Config) (*tls.Config, error) {
	switch envConfig.HTTPServerTLSMode {
	case TLSModeTLS:
		serverCert, err := tls.LoadX509KeyPair(envConfig.HTTPServerCert, envConfig.HTTPServerKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load the server certificates (%w)", err)
		}
		return &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{serverCert},
		}, nil
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
		return &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCAs,
		}, nil
	default:
		return nil, fmt.Errorf("invalid TLS mode: %s", envConfig.HTTPServerTLSMode)
	}
}

// loadMutualTLSClientCAs loads client CA certificates for mutual TLS.
func loadMutualTLSClientCAs(clientCaCertPaths []string) (*x509.CertPool, error) {
	clientCAs := x509.NewCertPool()
	for _, caCertPath := range clientCaCertPaths {
		cleanPath := filepath.Clean(caCertPath)
		caCert, err := os.ReadFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("could not read client CA certificate on path %s (%w)", cleanPath, err)
		}
		if ok := clientCAs.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append client CA certificate (%s)", cleanPath)
		}
	}
	return clientCAs, nil
}
