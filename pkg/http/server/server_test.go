package server_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/TriangleSide/GoTools/pkg/http/endpoints"
	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/middleware"
	"github.com/TriangleSide/GoTools/pkg/http/responders"
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

type testResponseError struct{}

func (t *testResponseError) Error() string {
	return "test error response"
}

func init() {
	responders.MustRegisterErrorResponse(
		http.StatusInternalServerError,
		func(err *testResponseError) *responders.StandardErrorResponse {
			return &responders.StandardErrorResponse{
				Message: err.Error(),
			}
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
	allOpts = append(allOpts, server.WithRegistrars(defaultHandler(t)))
	srv, err := server.New(allOpts...)
	assert.NoError(t, err)
	assert.NotNil(t, srv)
	waitForShutdown := make(chan struct{})
	t.Cleanup(func() {
		assert.NoError(t, srv.Shutdown(context.Background()))
		<-waitForShutdown
	})
	go func() {
		assert.NoError(t, srv.Run())
		close(waitForShutdown)
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

type tlsTestFixture struct {
	TempDir                  string
	ClientCACertPath         string
	ClientCACertPaths        []string
	ServerPrivateKeyPath     string
	ServerCertificatePath    string
	ClientCertificateKeyPair tls.Certificate
	InvalidClientCert        tls.Certificate
	CACertPool               *x509.CertPool
}

func setupTLSTestFixture(t *testing.T) *tlsTestFixture {
	t.Helper()
	tempDir := t.TempDir()

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	caCertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caCertBytes, err := x509.CreateCertificate(
		rand.Reader, &caCertTemplate, &caCertTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	assert.NoError(t, err)
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})

	clientCACertPath := filepath.Join(tempDir, "ca_cert.pem")
	assert.NoError(t, os.WriteFile(clientCACertPath, caCertPEM, 0600))
	clientCaCertPaths := []string{clientCACertPath}

	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	serverPrivateKeyPEM := pem.EncodeToMemory(
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey)})

	serverCertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Server Tests Inc."},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	serverCertBytes, err := x509.CreateCertificate(
		rand.Reader, &serverCertTemplate, &caCertTemplate, &serverPrivateKey.PublicKey, caPrivateKey)
	assert.NoError(t, err)
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})

	serverPrivateKeyPath := filepath.Join(tempDir, "server_key.pem")
	assert.NoError(t, os.WriteFile(serverPrivateKeyPath, serverPrivateKeyPEM, 0600))

	serverCertificatePath := filepath.Join(tempDir, "server_cert.pem")
	assert.NoError(t, os.WriteFile(serverCertificatePath, serverCertPEM, 0600))

	clientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	clientPrivateKeyPEM := pem.EncodeToMemory(
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientPrivateKey)})

	clientCertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization: []string{"Client Tests Inc."},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	clientCertBytes, err := x509.CreateCertificate(
		rand.Reader, &clientCertTemplate, &caCertTemplate, &clientPrivateKey.PublicKey, caPrivateKey)
	assert.NoError(t, err)
	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertBytes})

	clientPrivateKeyPath := filepath.Join(tempDir, "client_key.pem")
	assert.NoError(t, os.WriteFile(clientPrivateKeyPath, clientPrivateKeyPEM, 0600))

	clientCertificatePath := filepath.Join(tempDir, "client_cert.pem")
	assert.NoError(t, os.WriteFile(clientCertificatePath, clientCertPEM, 0600))

	clientCertificateKeyPair, err := tls.LoadX509KeyPair(clientCertificatePath, clientPrivateKeyPath)
	assert.NoError(t, err)

	invalidClientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	invalidClientCertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(4),
		Subject: pkix.Name{
			Organization: []string{"Invalid Client"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	invalidClientCertBytes, err := x509.CreateCertificate(
		rand.Reader, &invalidClientCertTemplate, &invalidClientCertTemplate,
		&invalidClientPrivateKey.PublicKey, invalidClientPrivateKey)
	assert.NoError(t, err)
	invalidClientCertPEM := pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: invalidClientCertBytes})
	invalidClientKeyPEM := pem.EncodeToMemory(
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(invalidClientPrivateKey)})

	invalidClientCert, err := tls.X509KeyPair(invalidClientCertPEM, invalidClientKeyPEM)
	assert.NoError(t, err)

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(clientCertPEM)
	caCertPool.AppendCertsFromPEM(serverCertPEM)

	return &tlsTestFixture{
		TempDir:                  tempDir,
		ClientCACertPath:         clientCACertPath,
		ClientCACertPaths:        clientCaCertPaths,
		ServerPrivateKeyPath:     serverPrivateKeyPath,
		ServerCertificatePath:    serverCertificatePath,
		ClientCertificateKeyPair: clientCertificateKeyPair,
		InvalidClientCert:        invalidClientCert,
		CACertPool:               caCertPool,
	}
}

func certPathsConfigProvider(t *testing.T, fixture *tlsTestFixture) *server.Config {
	t.Helper()
	cfg := getDefaultConfig(t)
	cfg.HTTPServerKey = fixture.ServerPrivateKeyPath
	cfg.HTTPServerCert = fixture.ServerCertificatePath
	cfg.HTTPServerClientCACerts = fixture.ClientCACertPaths
	cfg.HTTPServerKeepAlive = false
	return cfg
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

func TestNew_TLSModeWithMissingKeys_ReturnsError(t *testing.T) {
	t.Parallel()
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := getDefaultConfig(t)
		cfg.HTTPServerTLSMode = server.TLSModeTLS
		cfg.HTTPServerKey = ""
		cfg.HTTPServerCert = ""
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "failed to load the server certificates")
	assert.Nil(t, srv)
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
		server.WithRegistrars(&testHandler{
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
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+serverAddr+"/slow", nil)
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
	), server.WithRegistrars(&testHandler{
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
	}), server.WithRegistrars(
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

func TestNew_TLSModeWithMissingCert_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg := certPathsConfigProvider(t, fixture)
			cfg.HTTPServerCert = ""
			cfg.HTTPServerTLSMode = mode
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	}
}

func TestNew_TLSModeWithMissingKey_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg := certPathsConfigProvider(t, fixture)
			cfg.HTTPServerKey = ""
			cfg.HTTPServerTLSMode = mode
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	}
}

func TestNew_MutualTLSModeWithMissingClientCA_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerClientCACerts = []string{}
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "no client CAs provided")
	assert.Nil(t, srv)
}

func TestNew_MutualTLSModeWithNonexistentClientCA_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerClientCACerts = []string{"does_not_exist.pem"}
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "could not read client CA certificate")
	assert.Nil(t, srv)
}

func TestNew_TLSModeWithInvalidCert_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	invalidCertPath := filepath.Join(fixture.TempDir, "invalid_cert.pem")
	assert.NoError(t, os.WriteFile(invalidCertPath, []byte("invalid data"), 0600))
	for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg := certPathsConfigProvider(t, fixture)
			cfg.HTTPServerTLSMode = mode
			cfg.HTTPServerCert = invalidCertPath
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	}
}

func TestNew_TLSModeWithInvalidKey_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	invalidKeyPath := filepath.Join(fixture.TempDir, "invalid_key.pem")
	assert.NoError(t, os.WriteFile(invalidKeyPath, []byte("invalid data"), 0600))
	for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg := certPathsConfigProvider(t, fixture)
			cfg.HTTPServerTLSMode = mode
			cfg.HTTPServerKey = invalidKeyPath
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	}
}

func TestNew_MutualTLSModeWithInvalidClientCA_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	invalidCertPath := filepath.Join(fixture.TempDir, "invalid_ca.pem")
	assert.NoError(t, os.WriteFile(invalidCertPath, []byte("invalid data"), 0600))
	srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		cfg.HTTPServerClientCACerts = []string{invalidCertPath}
		return cfg, nil
	}))
	assert.ErrorPart(t, err, "failed to load client CA certificates")
	assert.Nil(t, srv)
}

func TestRun_TLSModeWithUntrustedClient_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddr := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+serverAddr, nil)
	assert.NoError(t, err)
	response, err := httpClient.Do(request)
	if response != nil {
		assert.Nil(t, response.Body.Close())
	}
	assert.ErrorPart(t, err, "certificate")
	assert.Nil(t, response)
}

func TestRun_TLSModeWithProperClient_Succeeds(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddress, true)
}

func TestRun_TLSModeWithClientCA_Succeeds(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeTLS
		return cfg, nil
	}))
	caCertPEM, err := os.ReadFile(fixture.ClientCACertPath)
	assert.NoError(t, err)
	caCertPool := x509.NewCertPool()
	assert.True(t, caCertPool.AppendCertsFromPEM(caCertPEM))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caCertPool,
				MinVersion: tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddress, true)
}

func TestRun_MutualTLSModeWithoutClientCert_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+serverAddress, nil)
	assert.NoError(t, err)
	assert.NotNil(t, request)
	response, err := httpClient.Do(request)
	if response != nil {
		assert.Nil(t, response.Body.Close())
	}
	assert.ErrorPart(t, err, "tls: certificate required")
	assert.Nil(t, response)
}

func TestRun_MutualTLSModeWithValidClientCert_Succeeds(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				Certificates:       []tls.Certificate{fixture.ClientCertificateKeyPair},
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddress, true)
}

func TestRun_MutualTLSModeWithInvalidClientCert_ReturnsError(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				Certificates:       []tls.Certificate{fixture.InvalidClientCert},
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://"+serverAddress, nil)
	assert.NoError(t, err)
	response, err := httpClient.Do(request)
	if response != nil {
		assert.Nil(t, response.Body.Close())
	}
	assert.ErrorPart(t, err, "tls: certificate required")
	assert.Nil(t, response)
}

func TestRun_MutualTLSModeWithMultipleClientCAs_AcceptsClientsFromAnyCA(t *testing.T) {
	t.Parallel()
	fixture := setupTLSTestFixture(t)
	secondCACertPath := filepath.Join(fixture.TempDir, "second_ca_cert.pem")
	secondCAPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	secondCACertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject: pkix.Name{
			Organization: []string{"Second Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	secondCACertBytes, err := x509.CreateCertificate(
		rand.Reader, &secondCACertTemplate, &secondCACertTemplate,
		&secondCAPrivateKey.PublicKey, secondCAPrivateKey)
	assert.NoError(t, err)
	secondCACertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: secondCACertBytes})
	assert.NoError(t, os.WriteFile(secondCACertPath, secondCACertPEM, 0600))
	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		cfg := certPathsConfigProvider(t, fixture)
		cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
		cfg.HTTPServerClientCACerts = append(cfg.HTTPServerClientCACerts, secondCACertPath)
		return cfg, nil
	}))
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				RootCAs:            fixture.CACertPool,
				Certificates:       []tls.Certificate{fixture.ClientCertificateKeyPair},
				MinVersion:         tls.VersionTLS13,
			},
		},
	}
	assertRootRequestSuccess(t, httpClient, serverAddress, true)
}

func TestRun_ConcurrentRequests_NoErrors(t *testing.T) {
	t.Parallel()

	serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
		return getDefaultConfig(t), nil
	}), server.WithRegistrars(
		&testHandler{
			Path:   "/status",
			Method: http.MethodGet,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				type params struct {
					Value string `json:"-" urlQuery:"value" validate:"required"`
				}
				responders.Status[params](writer, request, func(*params) (int, error) {
					return http.StatusOK, nil
				})
			},
		},
		&testHandler{
			Path:   "/error",
			Method: http.MethodGet,
			Handler: func(writer http.ResponseWriter, _ *http.Request) {
				responders.Error(writer, &testResponseError{})
			},
		},
		&testHandler{
			Path:   "/json/{id}",
			Method: http.MethodPost,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				type requestParams struct {
					ID   string `json:"-"    urlPath:"id"        validate:"required"`
					Data string `json:"data" validate:"required"`
				}
				type response struct {
					ID string
				}
				responders.JSON(writer, request, func(params *requestParams) (*response, int, error) {
					return &response{
						ID: params.ID,
					}, http.StatusOK, nil
				})
			},
		},
		&testHandler{
			Path:   "/jsonstream",
			Method: http.MethodGet,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				type requestParams struct{}
				type response struct {
					ID string
				}
				responders.JSONStream(writer, request, func(*requestParams) (<-chan *response, int, error) {
					responseChan := make(chan *response)
					go func() {
						defer close(responseChan)
						responseChan <- &response{ID: "1"}
						responseChan <- &response{ID: "2"}
						responseChan <- &response{ID: "3"}
					}()
					return responseChan, http.StatusOK, nil
				})
			},
		},
	))

	testCases := []struct {
		method      string
		path        string
		body        func() io.Reader
		contentType string
		expected    int
	}{
		{http.MethodGet, "/error",
			nil,
			"", http.StatusInternalServerError},
		{http.MethodGet, "/status?value=test",
			nil,
			"", http.StatusOK},
		{http.MethodGet, "/status",
			nil,
			"", http.StatusBadRequest},
		{http.MethodPost, "/json/testId",
			func() io.Reader { return bytes.NewBufferString(`{"data":"value"}`) },
			headers.ContentTypeApplicationJSON, http.StatusOK},
		{http.MethodPost, "/json/testId",
			func() io.Reader { return bytes.NewBufferString(`{"data":""}`) },
			headers.ContentTypeApplicationJSON, http.StatusBadRequest},
		{http.MethodGet, "/jsonstream",
			nil,
			"", http.StatusOK},
	}

	var waitGroup sync.WaitGroup
	waitToStart := make(chan struct{})
	totalGoRoutinesPerOperation := 2
	totalRequestsPerGoRoutine := 1000

	for _, testCase := range testCases {
		for range totalGoRoutinesPerOperation {
			waitGroup.Go(func() {
				<-waitToStart
				for range totalRequestsPerGoRoutine {
					var body io.Reader
					if testCase.body != nil {
						body = testCase.body()
					}
					ctx, cancel := context.WithCancel(t.Context())
					request, err := http.NewRequestWithContext(ctx, testCase.method, "http://"+serverAddress+testCase.path, body)
					if err != nil {
						cancel()
						assert.NoError(t, err, assert.Continue())
						continue
					}
					if testCase.contentType != "" {
						request.Header.Set(headers.ContentType, testCase.contentType)
					}
					response, err := http.DefaultClient.Do(request)
					cancel()
					if err != nil {
						assert.NoError(t, err, assert.Continue())
						continue
					}
					assert.Equals(t, response.StatusCode, testCase.expected, assert.Continue())
					assert.NoError(t, response.Body.Close(), assert.Continue())
				}
			})
		}
	}

	close(waitToStart)
	waitGroup.Wait()
}
