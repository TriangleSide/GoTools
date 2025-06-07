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

	"github.com/TriangleSide/GoTools/pkg/config"
	"github.com/TriangleSide/GoTools/pkg/http/api"
	"github.com/TriangleSide/GoTools/pkg/http/headers"
	"github.com/TriangleSide/GoTools/pkg/http/middleware"
	"github.com/TriangleSide/GoTools/pkg/http/responders"
	"github.com/TriangleSide/GoTools/pkg/http/server"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

type testHandler struct {
	Path       string
	Method     string
	Middleware []middleware.Middleware
	Handler    http.HandlerFunc
}

func (t *testHandler) AcceptHTTPAPIBuilder(builder *api.HTTPAPIBuilder) {
	builder.MustRegister(api.Path(t.Path), api.Method(t.Method), &api.Handler{
		Middleware: t.Middleware,
		Handler:    t.Handler,
	})
}

type testErrorResponse struct{}

func (t *testErrorResponse) Error() string {
	return "test error response"
}

func init() {
	responders.MustRegisterErrorResponse(http.StatusInternalServerError, func(err *testErrorResponse) *responders.StandardErrorResponse {
		return &responders.StandardErrorResponse{
			Message: err.Error(),
		}
	})
}

func TestServer(t *testing.T) {
	t.Setenv("HTTP_SERVER_KEEP_ALIVE", "false")
	t.Setenv("HTTP_SERVER_TLS_MODE", string(server.TLSModeOff))
	http.DefaultTransport.(*http.Transport).DisableKeepAlives = true

	handler := &testHandler{
		Path:       "/",
		Method:     http.MethodGet,
		Middleware: nil,
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			_, err := io.WriteString(writer, "PONG")
			assert.NoError(t, err)
		},
	}

	startServer := func(t *testing.T, options ...server.Option) string {
		t.Helper()
		waitUntilReady := make(chan struct{})
		var address string
		allOpts := append(options, server.WithBoundCallback(func(addr *net.TCPAddr) {
			address = addr.String()
			close(waitUntilReady)
		}), server.WithEndpointHandlers(handler))
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

	assertRootRequestSuccess := func(t *testing.T, httpClient *http.Client, addr string, tls bool) {
		t.Helper()
		var protocol string
		if tls {
			protocol = "https"
		} else {
			protocol = "http"
		}
		request, err := http.NewRequest(http.MethodGet, protocol+"://"+addr, nil)
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

	t.Run("when a server is instantiated it should fail if there's an error when parsing the environment variables", func(t *testing.T) {
		t.Parallel()
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			return nil, errors.New("config error")
		}))
		assert.ErrorPart(t, err, "could not load configuration (config error)")
		assert.Nil(t, srv)
	})

	t.Run("when a server is started it should fail if the TLS mode is invalid", func(t *testing.T) {
		t.Parallel()
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg, err := config.ProcessAndValidate[server.Config]()
			assert.NoError(t, err)
			cfg.HTTPServerTLSMode = "invalid_mode"
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "invalid TLS mode: invalid_mode")
		assert.Nil(t, srv)
	})

	t.Run("when a server is run it should fail if when the address is incorrectly formatted", func(t *testing.T) {
		t.Parallel()
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg, err := config.ProcessAndValidate[server.Config]()
			assert.NoError(t, err)
			cfg.HTTPServerBindIP = "not_an_ip"
			return cfg, nil
		}))
		assert.NoError(t, err)
		assert.NotNil(t, srv)
		err = srv.Run()
		assert.ErrorPart(t, err, "failed to create the network listener")
	})

	t.Run("when a server is started it should fail if the keys are missing when the TLS mode is TLS", func(t *testing.T) {
		t.Parallel()
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg, err := config.ProcessAndValidate[server.Config]()
			assert.NoError(t, err)
			cfg.HTTPServerTLSMode = server.TLSModeTLS
			cfg.HTTPServerKey = ""
			cfg.HTTPServerCert = ""
			return cfg, nil
		}))
		assert.ErrorPart(t, err, "failed to load the server certificates")
		assert.Nil(t, srv)
	})

	t.Run("when the server bind port is already take it should fail when starting", func(t *testing.T) {
		t.Parallel()
		const ip = "::1"
		listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(ip), Port: 0})
		assert.NoError(t, err)
		t.Cleanup(func() {
			assert.NoError(t, listener.Close())
		})
		addr, ok := listener.Addr().(*net.TCPAddr)
		assert.True(t, ok)
		listenerPort := addr.AddrPort().Port()
		srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
			cfg, err := config.ProcessAndValidate[server.Config]()
			assert.NoError(t, err)
			cfg.HTTPServerBindIP = ip
			cfg.HTTPServerBindPort = listenerPort
			return cfg, nil
		}))
		assert.NoError(t, err)
		assert.NotNil(t, srv)
		err = srv.Run()
		assert.ErrorPart(t, err, "address already in use")
	})

	t.Run("when the server is started twice it should panic", func(t *testing.T) {
		t.Parallel()
		waitUntilReady := make(chan bool)
		srv, err := server.New(server.WithBoundCallback(func(*net.TCPAddr) {
			close(waitUntilReady)
		}))
		assert.NoError(t, err)
		assert.NotNil(t, srv)

		go func() {
			assert.NoError(t, srv.Run())
		}()
		<-waitUntilReady

		assert.PanicPart(t, func() {
			_ = srv.Run()
		}, "HTTP server can only be run once per instance")
	})

	t.Run("when a server is started it should be able to be shutdown multiple times", func(t *testing.T) {
		t.Parallel()
		waitUntilReady := make(chan bool)
		srv, err := server.New(server.WithBoundCallback(func(*net.TCPAddr) {
			close(waitUntilReady)
		}))
		assert.NoError(t, err)
		assert.NotNil(t, srv)
		go func() {
			assert.NoError(t, srv.Run())
		}()
		<-waitUntilReady
		for i := 0; i < 3; i++ {
			assert.NoError(t, srv.Shutdown(context.Background()))
		}
	})

	t.Run("when a server is started it should return an error when the TCP listener is closed unexpectedly", func(t *testing.T) {
		t.Parallel()
		listener, err := net.ListenTCP("tcp6", &net.TCPAddr{IP: net.ParseIP("::1"), Port: 0})
		waitUntilReady := make(chan bool)
		srv, err := server.New(server.WithListenerProvider(func(bindIP string, bindPort uint16) (*net.TCPListener, error) {
			return listener, err
		}), server.WithBoundCallback(func(*net.TCPAddr) {
			close(waitUntilReady)
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
	})

	t.Run("when common middleware is added to the server it should execute in order", func(t *testing.T) {
		t.Parallel()
		seq := make([]string, 0)
		serverAddr := startServer(t, server.WithCommonMiddleware(
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
		), server.WithEndpointHandlers(&testHandler{
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
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				seq = append(seq, "4")
				writer.WriteHeader(http.StatusOK)
			},
		}))
		httpClient := &http.Client{}
		request, err := http.NewRequest(http.MethodGet, "http://"+serverAddr+"/test", nil)
		assert.NoError(t, err)
		response, err := httpClient.Do(request)
		t.Cleanup(func() {
			assert.NoError(t, response.Body.Close())
		})
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equals(t, seq, []string{"0", "1", "2", "3", "4"})
	})

	t.Run("when a server is started without TLS an HTTP client should be able to make requests", func(t *testing.T) {
		t.Parallel()
		serverAddr := startServer(t)
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
					MinVersion:         tls.VersionTLS13,
				},
			},
		}
		assertRootRequestSuccess(t, httpClient, serverAddr, false)
	})

	t.Run("when certificates are generated for TLS and mTLS", func(t *testing.T) {
		t.Parallel()
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

		caCertBytes, err := x509.CreateCertificate(rand.Reader, &caCertTemplate, &caCertTemplate, &caPrivateKey.PublicKey, caPrivateKey)
		assert.NoError(t, err)
		caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})

		clientCACertPath := filepath.Join(tempDir, "ca_cert.pem")
		assert.NoError(t, os.WriteFile(clientCACertPath, caCertPEM, 0644))
		clientCaCertPaths := []string{clientCACertPath}

		serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)
		serverPrivateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey)})

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
		serverCertBytes, err := x509.CreateCertificate(rand.Reader, &serverCertTemplate, &caCertTemplate, &serverPrivateKey.PublicKey, caPrivateKey)
		assert.NoError(t, err)
		serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})

		serverPrivateKeyPath := filepath.Join(tempDir, "server_key.pem")
		assert.NoError(t, os.WriteFile(serverPrivateKeyPath, serverPrivateKeyPEM, 0600))

		serverCertificatePath := filepath.Join(tempDir, "server_cert.pem")
		assert.NoError(t, os.WriteFile(serverCertificatePath, serverCertPEM, 0644))

		clientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)
		clientPrivateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientPrivateKey)})

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
		clientCertBytes, err := x509.CreateCertificate(rand.Reader, &clientCertTemplate, &caCertTemplate, &clientPrivateKey.PublicKey, caPrivateKey)
		assert.NoError(t, err)
		clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertBytes})

		clientPrivateKeyPath := filepath.Join(tempDir, "client_key.pem")
		assert.NoError(t, os.WriteFile(clientPrivateKeyPath, clientPrivateKeyPEM, 0600))

		clientCertificatePath := filepath.Join(tempDir, "client_cert.pem")
		assert.NoError(t, os.WriteFile(clientCertificatePath, clientCertPEM, 0644))

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
		invalidClientCertBytes, err := x509.CreateCertificate(rand.Reader, &invalidClientCertTemplate, &invalidClientCertTemplate, &invalidClientPrivateKey.PublicKey, invalidClientPrivateKey)
		assert.NoError(t, err)
		invalidClientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: invalidClientCertBytes})
		invalidClientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(invalidClientPrivateKey)})

		invalidClientCert, err := tls.X509KeyPair(invalidClientCertPEM, invalidClientKeyPEM)
		assert.NoError(t, err)

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(clientCertPEM)
		caCertPool.AppendCertsFromPEM(serverCertPEM)

		certPathsConfigProvider := func(t *testing.T) *server.Config {
			cfg, configErr := config.ProcessAndValidate[server.Config]()
			assert.NoError(t, configErr)
			cfg.HTTPServerKey = serverPrivateKeyPath
			cfg.HTTPServerCert = serverCertificatePath
			cfg.HTTPServerClientCACerts = clientCaCertPaths
			return cfg
		}

		t.Run("when the server certificate is missing it should fail to create the server", func(t *testing.T) {
			t.Parallel()
			for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
				srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
					cfg := certPathsConfigProvider(t)
					cfg.HTTPServerCert = ""
					cfg.HTTPServerTLSMode = mode
					return cfg, nil
				}))
				assert.ErrorPart(t, err, "failed to load the server certificates")
				assert.Nil(t, srv)
			}
		})

		t.Run("when the server key is missing it should fail to create the server", func(t *testing.T) {
			t.Parallel()
			for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
				srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
					cfg := certPathsConfigProvider(t)
					cfg.HTTPServerKey = ""
					cfg.HTTPServerTLSMode = mode
					return cfg, nil
				}))
				assert.ErrorPart(t, err, "failed to load the server certificates")
				assert.Nil(t, srv)
			}
		})

		t.Run("when the client CA is missing it should fail to create the server", func(t *testing.T) {
			t.Parallel()
			srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerClientCACerts = []string{}
				cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
				return cfg, nil
			}))
			assert.ErrorPart(t, err, "no client CAs provided")
			assert.Nil(t, srv)
		})

		t.Run("when the client CA doesn't exist should fail to create the server", func(t *testing.T) {
			t.Parallel()
			srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerClientCACerts = []string{"does_not_exist.pem"}
				cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
				return cfg, nil
			}))
			assert.ErrorPart(t, err, "could not read client CA certificate")
			assert.Nil(t, srv)
		})

		t.Run("when the server certificate is invalid it should fail to be created", func(t *testing.T) {
			t.Parallel()
			invalidCertPath := filepath.Join(tempDir, "invalid_cert.pem")
			assert.NoError(t, os.WriteFile(invalidCertPath, []byte("invalid data"), 0644))
			for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
				srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
					cfg := certPathsConfigProvider(t)
					cfg.HTTPServerTLSMode = mode
					cfg.HTTPServerCert = invalidCertPath
					return cfg, nil
				}))
				assert.ErrorPart(t, err, "failed to load the server certificates")
				assert.Nil(t, srv)
			}
		})

		t.Run("when the server key is invalid it should fail to be created", func(t *testing.T) {
			t.Parallel()
			invalidKeyPath := filepath.Join(tempDir, "invalid_key.pem")
			assert.NoError(t, os.WriteFile(invalidKeyPath, []byte("invalid data"), 0644))
			for _, mode := range []server.TLSMode{server.TLSModeTLS, server.TLSModeMutualTLS} {
				srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
					cfg := certPathsConfigProvider(t)
					cfg.HTTPServerTLSMode = mode
					cfg.HTTPServerKey = invalidKeyPath
					return cfg, nil
				}))
				assert.ErrorPart(t, err, "failed to load the server certificates")
				assert.Nil(t, srv)
			}
		})

		t.Run("when the client CA is invalid it should fail to be created", func(t *testing.T) {
			t.Parallel()
			invalidCertPath := filepath.Join(tempDir, "invalid_ca.pem")
			assert.NoError(t, os.WriteFile(invalidCertPath, []byte("invalid data"), 0644))
			srv, err := server.New(server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
				cfg.HTTPServerClientCACerts = []string{invalidCertPath}
				return cfg, nil
			}))
			assert.ErrorPart(t, err, "failed to load client CA certificates")
			assert.Nil(t, srv)
		})

		t.Run("when a server is run with TLS it should fail with client that don't recognize the CA", func(t *testing.T) {
			t.Parallel()
			serverAddr := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerTLSMode = server.TLSModeTLS
				return cfg, nil
			}))
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: false,
					},
				},
			}
			request, err := http.NewRequest(http.MethodGet, "https://"+serverAddr, nil)
			assert.NoError(t, err)
			response, err := httpClient.Do(request)
			assert.ErrorPart(t, err, "unknown authority")
			assert.Nil(t, response)
		})

		t.Run("when a server is run with TLS it should succeed if the client is properly configured", func(t *testing.T) {
			t.Parallel()
			serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerTLSMode = server.TLSModeTLS
				return cfg, nil
			}))
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: false,
						RootCAs:            caCertPool,
					},
				},
			}
			assertRootRequestSuccess(t, httpClient, serverAddress, true)
		})

		t.Run("when a server is run with TLS it should succeed if the client doesn't trust the CA but insecure is set", func(t *testing.T) {
			t.Parallel()
			serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerTLSMode = server.TLSModeTLS
				return cfg, nil
			}))
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
			assertRootRequestSuccess(t, httpClient, serverAddress, true)
		})

		t.Run("when an HTTP client is created without a client certificate it should fail to connect to an mTLS server", func(t *testing.T) {
			t.Parallel()
			serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
				return cfg, nil
			}))
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: false,
						RootCAs:            caCertPool,
					},
				},
			}
			request, err := http.NewRequest(http.MethodGet, "https://"+serverAddress, nil)
			assert.NoError(t, err)
			assert.NotNil(t, request)
			response, err := httpClient.Do(request)
			assert.ErrorPart(t, err, "tls: certificate required")
			assert.Nil(t, response)
		})

		t.Run("when an HTTP client is created with a client certificate signed by the trusted CA it should be able to get the contents of an mTLS server", func(t *testing.T) {
			t.Parallel()
			serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
				return cfg, nil
			}))
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: false,
						RootCAs:            caCertPool,
						Certificates:       []tls.Certificate{clientCertificateKeyPair},
					},
				},
			}
			assertRootRequestSuccess(t, httpClient, serverAddress, true)
		})

		t.Run("when an HTTP client is created with an invalid client certificate it should fail to connect to an mTLS server", func(t *testing.T) {
			t.Parallel()
			serverAddress := startServer(t, server.WithConfigProvider(func() (*server.Config, error) {
				cfg := certPathsConfigProvider(t)
				cfg.HTTPServerTLSMode = server.TLSModeMutualTLS
				return cfg, nil
			}))
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: false,
						RootCAs:            caCertPool,
						Certificates:       []tls.Certificate{invalidClientCert},
					},
				},
			}
			request, err := http.NewRequest(http.MethodGet, "https://"+serverAddress, nil)
			assert.NoError(t, err)
			response, err := httpClient.Do(request)
			assert.ErrorPart(t, err, "tls: certificate required")
			assert.Nil(t, response)
		})
	})

	t.Run("when the server handles concurrent requests there should be no error", func(t *testing.T) {
		t.Parallel()

		serverAddress := startServer(t, server.WithEndpointHandlers(
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
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					responders.Error(writer, &testErrorResponse{})
				},
			},
			&testHandler{
				Path:   "/json/{id}",
				Method: http.MethodPost,
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					type requestParams struct {
						Id   string `json:"-" urlPath:"id" validate:"required"`
						Data string `json:"data" validate:"required"`
					}
					type response struct {
						Id string
					}
					responders.JSON(writer, request, func(params *requestParams) (*response, int, error) {
						return &response{
							Id: params.Id,
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
						Id string
					}
					responders.JSONStream(writer, request, func(params *requestParams) (<-chan *response, int, error) {
						responseChan := make(chan *response)
						go func() {
							defer close(responseChan)
							responseChan <- &response{Id: "1"}
							responseChan <- &response{Id: "2"}
							responseChan <- &response{Id: "3"}
						}()
						return responseChan, http.StatusOK, nil
					})
				},
			},
		))

		wg := sync.WaitGroup{}
		waitToStart := make(chan struct{})
		totalGoRoutinesPerOperation := 2
		totalRequestsPerGoRoutine := 1000

		performRequest := func(t *testing.T, method, url string, body io.Reader, expected int) {
			t.Helper()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			request, err := http.NewRequestWithContext(ctx, method, url, body)
			if err != nil {
				assert.NoError(t, err, assert.Continue())
				return
			}
			if method == http.MethodPost {
				request.Header.Set(headers.ContentType, headers.ContentTypeApplicationJson)
			}
			response, err := http.DefaultClient.Do(request)
			assert.NoError(t, err, assert.Continue())
			if err != nil {
				return
			}
			assert.Equals(t, response.StatusCode, expected, assert.Continue())
			assert.NoError(t, response.Body.Close(), assert.Continue())
		}

		// Error endpoint.
		for routineI := 0; routineI < totalGoRoutinesPerOperation; routineI++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-waitToStart
				for i := 0; i < totalRequestsPerGoRoutine; i++ {
					performRequest(t, http.MethodGet, "http://"+serverAddress+"/error", nil, http.StatusInternalServerError)
				}
			}()
		}

		// Status endpoint good.
		for routineI := 0; routineI < totalGoRoutinesPerOperation; routineI++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-waitToStart
				for i := 0; i < totalRequestsPerGoRoutine; i++ {
					performRequest(t, http.MethodGet, "http://"+serverAddress+"/status?value=test", nil, http.StatusOK)
				}
			}()
		}

		// Status endpoint bad decode.
		for routineI := 0; routineI < totalGoRoutinesPerOperation; routineI++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-waitToStart
				for i := 0; i < totalRequestsPerGoRoutine; i++ {
					performRequest(t, http.MethodGet, "http://"+serverAddress+"/status", nil, http.StatusBadRequest)
				}
			}()
		}

		// JSON endpoint good.
		for routineI := 0; routineI < totalGoRoutinesPerOperation; routineI++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-waitToStart
				for i := 0; i < totalRequestsPerGoRoutine; i++ {
					bodyData := bytes.NewBuffer([]byte(`{"data":"value"}`))
					performRequest(t, http.MethodPost, "http://"+serverAddress+"/json/testId", bodyData, http.StatusOK)
				}
			}()
		}

		// JSON endpoint invalid.
		for routineI := 0; routineI < totalGoRoutinesPerOperation; routineI++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-waitToStart
				for i := 0; i < totalRequestsPerGoRoutine; i++ {
					bodyData := bytes.NewBuffer([]byte(`{"data":""}`))
					performRequest(t, http.MethodPost, "http://"+serverAddress+"/json/testId", bodyData, http.StatusBadRequest)
				}
			}()
		}

		// JSONStream endpoint good.
		for routineI := 0; routineI < totalGoRoutinesPerOperation; routineI++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-waitToStart
				for i := 0; i < totalRequestsPerGoRoutine; i++ {
					performRequest(t, http.MethodGet, "http://"+serverAddress+"/jsonstream", nil, http.StatusOK)
				}
			}()
		}

		close(waitToStart)
		wg.Wait()
	})
}
