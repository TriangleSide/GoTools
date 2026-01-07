package server_test

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/http/endpoints"
	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/middleware"
	"github.com/TriangleSide/GoTools/pkg/http/server"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

type testHandler struct {
	Path       string
	Method     string
	Middleware []middleware.Middleware
	Handler    http.HandlerFunc
}

func (t *testHandler) RegisterEndpoints(builder *endpoints.Builder) {
	builder.MustRegister(endpoints.Path(t.Path), endpoints.Method(t.Method), &endpoints.Endpoint{
		Middleware: t.Middleware,
		Handler:    t.Handler,
	})
}

func getDefaultConfig(t *testing.T) *server.Config {
	t.Helper()
	defaults := &server.Config{
		HTTPServerBindIP:                  "::1",
		HTTPServerBindPort:                0,
		HTTPServerReadTimeoutMillis:       120000,
		HTTPServerWriteTimeoutMillis:      120000,
		HTTPServerIdleTimeoutMillis:       0,
		HTTPServerHeaderReadTimeoutMillis: 0,
		HTTPServerTLSMode:                 server.TLSModeOff,
		HTTPServerCert:                    "",
		HTTPServerKey:                     "",
		HTTPServerClientCACerts:           []string{},
		HTTPServerMaxHeaderBytes:          1048576,
		HTTPServerKeepAlive:               false,
	}
	err := validation.Struct(defaults)
	assert.Nil(t, err)
	return defaults
}

func defaultHandler(t *testing.T) *testHandler {
	t.Helper()
	return &testHandler{
		Path:       "/",
		Method:     http.MethodGet,
		Middleware: nil,
		Handler: func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			_, err := io.WriteString(writer, "PONG")
			assert.NoError(t, err)
		},
	}
}

func startServer(t *testing.T, options ...server.Option) string {
	t.Helper()
	waitUntilReady := make(chan struct{})
	var address string
	allOpts := make([]server.Option, 0, len(options)+2)
	allOpts = append(allOpts, options...)
	allOpts = append(allOpts, server.WithBoundCallback(func(addr *net.TCPAddr) {
		address = addr.String()
		close(waitUntilReady)
	}))
	allOpts = append(allOpts, server.WithEndpoints(defaultHandler(t)))
	srv, err := server.New(allOpts...)
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
		defer cancel()
		assert.NoError(t, srv.Shutdown(shutdownCtx), assert.Continue())
	})
	go func() {
		assert.NoError(t, srv.Run(), assert.Continue())
	}()
	<-waitUntilReady
	return address
}

func assertRootRequestSuccess(t *testing.T, httpClient *http.Client, addr string, useTLS bool) {
	t.Helper()
	var protocol string
	if useTLS {
		protocol = "https"
	} else {
		protocol = "http"
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, protocol+"://"+addr, nil)
	assert.NoError(t, err)
	assert.NotNil(t, request)
	response, err := httpClient.Do(request)
	assert.NoError(t, err)
	assert.Equals(t, http.StatusOK, response.StatusCode)
	assert.NotNil(t, response.Body)
	bodyContents, err := io.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.Equals(t, bodyContents, []byte("PONG"))
	assert.NoError(t, response.Body.Close())
}

func TestNew_ConfigProviderError_ReturnsError(t *testing.T) {
	t.Parallel()
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		return nil, errors.New("config error")
	}))
	assert.ErrorPart(t, err, "could not load configuration: config error")
	assert.Nil(t, srv)
}

func TestNew_InvalidTLSMode_ReturnsError(t *testing.T) {
	t.Parallel()
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := getDefaultConfig(t)
		cfg.HTTPServerTLSMode = "invalid_mode"
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "invalid TLS mode: invalid_mode")
	assert.Nil(t, srv)
}

func TestRun_InvalidBindAddress_ReturnsError(t *testing.T) {
	t.Parallel()
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := getDefaultConfig(t)
		cfg.HTTPServerBindIP = "not_an_ip"
		return cfg, nil
	}))
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	err = srv.Run()
	assert.ErrorPart(t, err, "failed to create the network listener")
}

func TestRun_PortAlreadyInUse_ReturnsError(t *testing.T) {
	t.Parallel()
	const ipAddress = "::1"
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(ipAddress), Port: 0})
	assert.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, listener.Close())
	})
	addr, ok := listener.Addr().(*net.TCPAddr)
	assert.True(t, ok)
	listenerPort := addr.AddrPort().Port()
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := getDefaultConfig(t)
		cfg.HTTPServerBindIP = ipAddress
		cfg.HTTPServerBindPort = listenerPort
		return cfg, nil
	}))
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	err = srv.Run()
	assert.ErrorPart(t, err, "address already in use")
}

func TestRun_CalledTwice_Panics(t *testing.T) {
	t.Parallel()
	waitUntilReady := make(chan bool)
	srv, err := server.New(server.WithBoundCallback(func(*net.TCPAddr) {
		close(waitUntilReady)
	}), server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}))
	assert.NoError(t, err)
	assert.NotNil(t, srv)

	go func() {
		assert.NoError(t, srv.Run())
	}()
	<-waitUntilReady

	assert.PanicPart(t, func() {
		_ = srv.Run()
	}, "http server can only be run once per instance")

	shutdownErr := srv.Shutdown(t.Context())
	assert.NoError(t, shutdownErr)
}

func TestShutdown_CalledMultipleTimes_Succeeds(t *testing.T) {
	t.Parallel()
	waitUntilReady := make(chan bool)
	srv, err := server.New(server.WithBoundCallback(func(*net.TCPAddr) {
		close(waitUntilReady)
	}), server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}))
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	go func() {
		assert.NoError(t, srv.Run())
	}()
	<-waitUntilReady
	for range 3 {
		assert.NoError(t, srv.Shutdown(t.Context()))
	}
}

func TestShutdown_WithExpiredContext_ReturnsError(t *testing.T) {
	t.Parallel()

	handlerStarted := make(chan struct{})
	handlerBlocking := make(chan struct{})

	waitUntilReady := make(chan struct{})
	var serverAddr string
	srv, err := server.New(
		server.WithBoundCallback(func(addr *net.TCPAddr) {
			serverAddr = addr.String()
			close(waitUntilReady)
		}),
		server.WithConfigProvider(func() (*server.Config, error) {
			return getDefaultConfig(t), nil
		}),
		server.WithEndpoints(&testHandler{
			Path:   "/slow",
			Method: http.MethodGet,
			Handler: func(writer http.ResponseWriter, _ *http.Request) {
				close(handlerStarted)
				<-handlerBlocking
				writer.WriteHeader(http.StatusOK)
			},
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, srv)

	serverDone := make(chan struct{})
	go func() {
		_ = srv.Run()
		close(serverDone)
	}()
	<-waitUntilReady

	clientChan := make(chan struct{})
	go func() {
		req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+serverAddr+"/slow", nil)
		response, clientErr := http.DefaultClient.Do(req)
		assert.NoError(t, clientErr)
		assert.NoError(t, response.Body.Close())
		close(clientChan)
	}()

	<-handlerStarted

	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Millisecond)
	defer cancel()

	err = srv.Shutdown(ctx)
	assert.ErrorPart(t, err, "failed to shutdown the server")

	close(handlerBlocking)
	<-serverDone
	<-clientChan
}

func TestRun_ListenerClosedUnexpectedly_ReturnsError(t *testing.T) {
	t.Parallel()
	listener, err := net.ListenTCP("tcp6", &net.TCPAddr{IP: net.ParseIP("::1"), Port: 0})
	assert.NoError(t, err)
	waitUntilReady := make(chan bool)
	srv, err := server.New(server.WithListenerProvider(func(string, uint16) (*net.TCPListener, error) {
		return listener, nil
	}), server.WithBoundCallback(func(*net.TCPAddr) {
		close(waitUntilReady)
	}), server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}))
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	srvErrChan := make(chan error, 1)
	go func() {
		srvErrChan <- srv.Run()
	}()
	<-waitUntilReady
	assert.NoError(t, listener.Close())
	err = <-srvErrChan
	assert.ErrorPart(t, err, "error encountered while serving http requests")
}

func TestRun_ListenerProviderError_ReturnsError(t *testing.T) {
	t.Parallel()
	srv, err := server.New(server.WithListenerProvider(func(string, uint16) (*net.TCPListener, error) {
		return nil, errors.New("listener error")
	}), server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}))
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	err = srv.Run()
	assert.ErrorPart(t, err, "failed to create the network listener: listener error")
}

func TestRun_WithCommonMiddleware_ExecutesInOrder(t *testing.T) {
	t.Parallel()
	seq := make([]string, 0)
	serverAddr := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}), server.WithCommonMiddleware(
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				seq = append(seq, "0")
				next(writer, request)
			}
		},
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				seq = append(seq, "1")
				next(writer, request)
			}
		},
	), server.WithEndpoints(&testHandler{
		Path:   "/test",
		Method: http.MethodGet,
		Middleware: []middleware.Middleware{
			func(next http.HandlerFunc) http.HandlerFunc {
				return func(writer http.ResponseWriter, request *http.Request) {
					seq = append(seq, "2")
					next(writer, request)
				}
			},
			func(next http.HandlerFunc) http.HandlerFunc {
				return func(writer http.ResponseWriter, request *http.Request) {
					seq = append(seq, "3")
					next(writer, request)
				}
			},
		},
		Handler: func(writer http.ResponseWriter, _ *http.Request) {
			seq = append(seq, "4")
			writer.WriteHeader(http.StatusOK)
		},
	}))
	httpClient := &http.Client{}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+serverAddr+"/test", nil)
	assert.NoError(t, err)
	response, err := httpClient.Do(request)
	t.Cleanup(func() {
		assert.NoError(t, response.Body.Close())
	})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equals(t, seq, []string{"0", "1", "2", "3", "4"})
}

func TestRun_MultipleMethodsOnSamePath_EnforcesMethodRouting(t *testing.T) {
	t.Parallel()

	serverAddr := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}), server.WithEndpoints(
		&testHandler{
			Path:   "/resource",
			Method: http.MethodGet,
			Handler: func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusOK)
				_, err := writer.Write([]byte("get"))
				assert.NoError(t, err)
			},
		},
		&testHandler{
			Path:   "/resource",
			Method: http.MethodPost,
			Handler: func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusCreated)
				_, err := writer.Write([]byte("post"))
				assert.NoError(t, err)
			},
		},
	))

	assertRequest := func(method string, expectedStatus int, expectedBody string, expectedAllowHeader string) {
		t.Helper()

		request, err := http.NewRequestWithContext(t.Context(), method, "http://"+serverAddr+"/resource", nil)
		assert.NoError(t, err)

		response, err := http.DefaultClient.Do(request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equals(t, response.StatusCode, expectedStatus)

		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)
		assert.Equals(t, string(body), expectedBody)

		if expectedAllowHeader != "" {
			assert.Equals(t, response.Header.Get(headers.Allow), expectedAllowHeader)
		}

		assert.NoError(t, response.Body.Close())
	}

	assertRequest(http.MethodGet, http.StatusOK, "get", "")
	assertRequest(http.MethodPost, http.StatusCreated, "post", "")
	assertRequest(http.MethodPut, http.StatusMethodNotAllowed, "", "GET, POST")
}

func TestRun_WithoutTLS_AllowsHTTPRequests(t *testing.T) {
	t.Parallel()
	serverAddr := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddr, false)
}

func TestNew_DefaultConfig_FailsDueToCertPathsMissing(t *testing.T) {
	t.Parallel()
	srv, err := server.New()
	assert.ErrorPart(t, err, "validation failed on field 'HTTPServerKey'")
	assert.Nil(t, srv)
}
