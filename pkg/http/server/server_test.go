package server_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/config"
	"github.com/TriangleSide/GoBase/pkg/config/envprocessor"
	"github.com/TriangleSide/GoBase/pkg/http/api"
	"github.com/TriangleSide/GoBase/pkg/http/middleware"
	"github.com/TriangleSide/GoBase/pkg/http/server"
	"github.com/TriangleSide/GoBase/pkg/network/tcp"
	tcplistener "github.com/TriangleSide/GoBase/pkg/network/tcp/listener"
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

var _ = Describe("http server", Ordered, func() {
	AfterEach(func() {
		unsetEnvironmentVariables()
	})

	When("a handler with middleware and common middleware is created", func() {
		const (
			path             = "/"
			body             = "PONG"
			mwSetValue       = "set"
			commonMwValueSet = "commonMwValueSet"
		)

		var (
			ctx            context.Context
			handlerMwValue string
			handlerMw      []middleware.Middleware
			handlers       []api.HTTPEndpointHandler
			commonMwValue  string
			commonMw       []middleware.Middleware
		)

		BeforeEach(func() {
			ctx = context.Background()

			handlerMwValue = ""
			handlerMw = []middleware.Middleware{
				func(next http.HandlerFunc) http.HandlerFunc {
					return func(writer http.ResponseWriter, request *http.Request) {
						handlerMwValue = mwSetValue
						next(writer, request)
					}
				},
			}

			handlers = []api.HTTPEndpointHandler{
				&testHandler{
					Path:       path,
					Method:     http.MethodGet,
					Middleware: handlerMw,
					Handler: func(writer http.ResponseWriter, request *http.Request) {
						writer.WriteHeader(http.StatusOK)
						_, err := io.WriteString(writer, body)
						Expect(err).ToNot(HaveOccurred())
					},
				},
			}

			commonMwValue = ""
			commonMw = []middleware.Middleware{
				func(next http.HandlerFunc) http.HandlerFunc {
					return func(writer http.ResponseWriter, request *http.Request) {
						commonMwValue = commonMwValueSet
						next(writer, request)
					}
				},
			}
		})

		When("server error cases", func() {
			var (
				err error
				srv *server.Server
			)

			BeforeEach(func() {
				err = nil
				srv = nil
				Expect(os.Setenv(string(config.HTTPServerTLSModeEnvName), string(config.HTTPServerTLSModeOff))).To(Succeed())
			})

			AfterEach(func() {
				if srv != nil {
					Expect(srv.Shutdown(ctx)).To(Succeed())
				}
			})

			It("should fail if the environment variables fail to be parsed", func() {
				srv, err = server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
					return nil, errors.New("config error")
				}))
				Expect(srv).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not load configuration (config error)"))
			})

			It("should fail if an invalid tls mode is provided", func() {
				srv, err = server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
					cfg, err := envprocessor.ProcessAndValidate[config.HTTPServer]()
					Expect(err).ToNot(HaveOccurred())
					cfg.HTTPServerTLSMode = "invalid_mode"
					return cfg, nil
				}))
				Expect(srv).To(Not(BeNil()))
				Expect(err).NotTo(HaveOccurred())
				err = srv.Run(commonMw, handlers, func() {})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid TLS mode: invalid_mode"))
			})

			It("should fail if the server address is incorrectly formatted", func() {
				srv, err = server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
					cfg, err := envprocessor.ProcessAndValidate[config.HTTPServer]()
					Expect(err).ToNot(HaveOccurred())
					cfg.HTTPServerBindIP = "not_an_ip"
					return cfg, nil
				}))
				Expect(srv).To(Not(BeNil()))
				Expect(err).NotTo(HaveOccurred())
				err = srv.Run(commonMw, handlers, func() {})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create the network listener"))
			})

			It("should fail if the server keys are missing when the tls mode is tls", func() {
				srv, err = server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
					cfg, err := envprocessor.ProcessAndValidate[config.HTTPServer]()
					Expect(err).ToNot(HaveOccurred())
					cfg.HTTPServerTLSMode = config.HTTPServerTLSModeTLS
					cfg.HTTPServerKey = ""
					cfg.HTTPServerCert = ""
					return cfg, nil
				}))
				Expect(srv).To(Not(BeNil()))
				Expect(err).NotTo(HaveOccurred())
				err = srv.Run(commonMw, handlers, func() {})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to load the server certificates"))
			})

			It("should fail if the port is already bound", func() {
				const ip = "::1"
				const port = 6789
				ln, err := tcplistener.New(ip, port)
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					Expect(ln.Close()).To(Succeed())
				}()
				srv, err = server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
					cfg, err := envprocessor.ProcessAndValidate[config.HTTPServer]()
					Expect(err).ToNot(HaveOccurred())
					cfg.HTTPServerBindIP = ip
					cfg.HTTPServerBindPort = port
					return cfg, nil
				}))
				Expect(srv).To(Not(BeNil()))
				Expect(err).NotTo(HaveOccurred())
				err = srv.Run(commonMw, handlers, func() {})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("address already in use"))
			})

			It("should panic when the server is started twice", func() {
				srv, err = server.New()
				Expect(srv).To(Not(BeNil()))
				Expect(err).NotTo(HaveOccurred())

				waitUntilReady := make(chan bool)
				go func() {
					defer GinkgoRecover()
					Expect(srv.Run(commonMw, handlers, func() {
						close(waitUntilReady)
					})).To(Succeed())
				}()
				<-waitUntilReady

				Expect(err).NotTo(HaveOccurred())
				Expect(func() {
					_ = srv.Run(commonMw, handlers, func() {})
				}).Should(PanicWith(ContainSubstring("HTTP server can only be run once per instance")))
			})

			It("should be able to be shutdown multiple times", func() {
				srv, err = server.New()
				Expect(srv).To(Not(BeNil()))
				Expect(err).NotTo(HaveOccurred())

				waitUntilReady := make(chan bool)
				go func() {
					defer GinkgoRecover()
					Expect(srv.Run(commonMw, handlers, func() {
						close(waitUntilReady)
					})).To(Succeed())
				}()
				<-waitUntilReady

				for i := 0; i < 3; i++ {
					Expect(srv.Shutdown(ctx)).To(Succeed())
				}
			})

			It("should return an error when the TCP listener is closed unexpectedly while the server is running", func() {
				var listener tcp.Listener
				srv, err = server.New(server.WithListenerProvider(func(host string, port uint16) (tcp.Listener, error) {
					listener, err = tcplistener.New(host, port)
					return listener, err
				}))
				Expect(srv).To(Not(BeNil()))
				Expect(err).NotTo(HaveOccurred())

				srvErrChan := make(chan error, 1)
				waitUntilReady := make(chan bool)
				go func() {
					defer GinkgoRecover()
					srvErrChan <- srv.Run(commonMw, handlers, func() {
						close(waitUntilReady)
					})
				}()
				<-waitUntilReady

				Expect(listener.Close()).To(Succeed())
				err = <-srvErrChan
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error encountered while serving http requests"))
			})
		})

		expectSuccessfulRootGet := func(httpClient *http.Client, host string, port uint16, protocol string) {
			request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s:%d%s", protocol, host, port, path), nil)
			Expect(err).NotTo(HaveOccurred())
			response, err := httpClient.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			Expect(response.Body).To(Not(BeNil()))
			responseBody, err := io.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(responseBody)).To(Equal(body))
			Expect(commonMwValue).To(Equal(commonMwValueSet))
			Expect(handlerMwValue).To(Equal(mwSetValue))
			Expect(response.Body.Close()).To(Succeed())
		}

		generateServerTests := func(host string, port uint16, clientTests func(host string, port uint16)) {
			When(fmt.Sprintf("a server is bound to IP %s and port %d is started", host, port), func() {
				var (
					srv        *server.Server
					srvErrChan chan error
				)

				BeforeEach(func() {
					var err error

					srvErrChan = make(chan error, 1)
					waitUntilReady := make(chan bool)

					srv, err = server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
						cfg, err := envprocessor.ProcessAndValidate[config.HTTPServer]()
						Expect(err).ToNot(HaveOccurred())
						cfg.HTTPServerBindIP = host
						cfg.HTTPServerBindPort = port
						return cfg, nil
					}))
					Expect(srv).To(Not(BeNil()))
					Expect(err).ToNot(HaveOccurred())

					go func() {
						defer GinkgoRecover()
						srvErrChan <- srv.Run(commonMw, handlers, func() {
							close(waitUntilReady)
						})
					}()
					<-waitUntilReady
				})

				AfterEach(func() {
					Expect(srv).To(Not(BeNil()))
					Expect(srv.Shutdown(ctx)).To(Succeed())
					Expect(<-srvErrChan).To(Not(HaveOccurred()))
				})

				clientTests(host, port)
			})
		}

		When("the tls mode is set to off", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.HTTPServerTLSModeEnvName), string(config.HTTPServerTLSModeOff))).To(Succeed())
			})

			generateInsecureClientTests := func(host string, port uint16) {
				When("an HTTPS client is created for the HTTP server", func() {
					var (
						strictHttpClient *http.Client
					)

					BeforeEach(func() {
						strictHttpClient = &http.Client{
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: false,
								},
							},
						}
					})

					It("should fail to connect to the server", func() {
						request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s:%d%s", host, port, path), nil)
						Expect(err).NotTo(HaveOccurred())
						response, err := strictHttpClient.Do(request)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("server gave HTTP response to HTTPS client"))
						Expect(response).To(BeNil())
					})
				})

				When("an HTTP client is created", func() {
					var (
						httpClient *http.Client
					)

					BeforeEach(func() {
						httpClient = &http.Client{}
					})

					It("should be able to get the root contents", func() {
						expectSuccessfulRootGet(httpClient, host, port, "http")
					})
				})
			}

			generateServerTests("::1", 18080, generateInsecureClientTests)
		})

		When("the tls mode is set to tls", func() {
			var (
				tempDir         string
				privateKeyPath  string
				certificatePath string
			)

			BeforeEach(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "server-test-*")
				Expect(err).ToNot(HaveOccurred())

				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				Expect(err).ToNot(HaveOccurred())
				privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

				privateKeyPath = filepath.Join(tempDir, "key.pem")
				Expect(os.WriteFile(privateKeyPath, privateKeyPEM, 0600)).To(Succeed())

				certificateTemplate := x509.Certificate{
					SerialNumber: big.NewInt(1),
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
				certBytes, err := x509.CreateCertificate(rand.Reader, &certificateTemplate, &certificateTemplate, &privateKey.PublicKey, privateKey)
				Expect(err).ToNot(HaveOccurred())
				certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

				certificatePath = filepath.Join(tempDir, "cert.pem")
				Expect(os.WriteFile(certificatePath, certificatePEM, 0644)).To(Succeed())

				Expect(os.Setenv(string(config.HTTPServerTLSModeEnvName), string(config.HTTPServerTLSModeTLS))).To(Succeed())
				Expect(os.Setenv(string(config.HTTPServerKeyEnvName), privateKeyPath)).To(Succeed())
				Expect(os.Setenv(string(config.HTTPServerCertEnvName), certificatePath)).To(Succeed())
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})

			When("the server certificate files cannot be loaded", func() {
				It("should fail to start the server when certificate file is missing", func() {
					Expect(os.Remove(certificatePath)).To(Succeed())
					srv, err := server.New()
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to load the server certificates"))
				})

				It("should fail to start the server when key file is missing", func() {
					Expect(os.Remove(privateKeyPath)).To(Succeed())
					srv, err := server.New()
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to load the server certificates"))
				})

				It("should fail to start the server when certificate file is invalid", func() {
					Expect(os.WriteFile(certificatePath, []byte("invalid data"), 0644)).To(Succeed())
					srv, err := server.New()
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to load the server certificates"))
				})

				It("should fail to start the server when key file is invalid", func() {
					Expect(os.WriteFile(privateKeyPath, []byte("invalid data"), 0600)).To(Succeed())
					srv, err := server.New()
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to load the server certificates"))
				})
			})

			generateTLSClientTests := func(host string, port uint16) {
				When("an HTTPS client is created that verifies the server certificate without trusting it", func() {
					var (
						strictHttpClient *http.Client
					)

					BeforeEach(func() {
						strictHttpClient = &http.Client{
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: false,
								},
							},
						}
					})

					It("should fail to connect to the server", func() {
						request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s:%d%s", host, port, path), nil)
						Expect(err).NotTo(HaveOccurred())
						response, err := strictHttpClient.Do(request)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("certificate signed by unknown authority"))
						Expect(response).To(BeNil())
					})
				})

				When("an HTTPS client is created that verifies the server certificate and trusts it", func() {
					var (
						httpClient *http.Client
					)

					BeforeEach(func() {
						caCert, err := os.ReadFile(certificatePath)
						Expect(err).To(Not(HaveOccurred()))
						caCertPool := x509.NewCertPool()
						caCertPool.AppendCertsFromPEM(caCert)
						httpClient = &http.Client{
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: false,
									RootCAs:            caCertPool,
								},
							},
						}
					})

					It("should be able to get the root contents", func() {
						expectSuccessfulRootGet(httpClient, host, port, "https")
					})
				})

				When("an HTTPS client is created that doesn't verify the server certificate", func() {
					var (
						httpClient *http.Client
					)

					BeforeEach(func() {
						httpClient = &http.Client{
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: true,
								},
							},
						}
					})

					It("should be able to get the root contents", func() {
						expectSuccessfulRootGet(httpClient, host, port, "https")
					})
				})
			}

			generateServerTests("::1", 5093, generateTLSClientTests)
		})

		When("the tls mode is set to mutual_tls", func() {
			var (
				tempDir                  string
				serverPrivateKeyPath     string
				serverCertificatePath    string
				clientCACertPath         string
				clientPrivateKeyPath     string
				clientCertificatePath    string
				clientCertificateKeyPair tls.Certificate
			)

			BeforeEach(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "server-test-*")
				Expect(err).ToNot(HaveOccurred())

				caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				Expect(err).ToNot(HaveOccurred())

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
				Expect(err).ToNot(HaveOccurred())
				caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})

				clientCACertPath = filepath.Join(tempDir, "ca_cert.pem")
				Expect(os.WriteFile(clientCACertPath, caCertPEM, 0644)).To(Succeed())
				clientCaCertPaths := []string{clientCACertPath}
				clientCaCertPathsBytes, err := json.Marshal(clientCaCertPaths)
				Expect(err).ToNot(HaveOccurred())
				clientCaCertPathsStr := string(clientCaCertPathsBytes)

				serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				Expect(err).ToNot(HaveOccurred())
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
				Expect(err).ToNot(HaveOccurred())
				serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})

				serverPrivateKeyPath = filepath.Join(tempDir, "server_key.pem")
				Expect(os.WriteFile(serverPrivateKeyPath, serverPrivateKeyPEM, 0600)).To(Succeed())

				serverCertificatePath = filepath.Join(tempDir, "server_cert.pem")
				Expect(os.WriteFile(serverCertificatePath, serverCertPEM, 0644)).To(Succeed())

				clientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				Expect(err).ToNot(HaveOccurred())
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
				Expect(err).ToNot(HaveOccurred())
				clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertBytes})

				clientPrivateKeyPath = filepath.Join(tempDir, "client_key.pem")
				Expect(os.WriteFile(clientPrivateKeyPath, clientPrivateKeyPEM, 0600)).To(Succeed())

				clientCertificatePath = filepath.Join(tempDir, "client_cert.pem")
				Expect(os.WriteFile(clientCertificatePath, clientCertPEM, 0644)).To(Succeed())

				clientCertificateKeyPair, err = tls.LoadX509KeyPair(clientCertificatePath, clientPrivateKeyPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(os.Setenv(string(config.HTTPServerTLSModeEnvName), string(config.HTTPServerTLSModeMutualTLS))).To(Succeed())
				Expect(os.Setenv(string(config.HTTPServerKeyEnvName), serverPrivateKeyPath)).To(Succeed())
				Expect(os.Setenv(string(config.HTTPServerCertEnvName), serverCertificatePath)).To(Succeed())
				Expect(os.Setenv(string(config.HTTPServerClientCACertsEnvName), clientCaCertPathsStr)).To(Succeed())
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			})

			When("the server certificate files cannot be loaded", func() {
				It("should fail to start the server when server certificate file is missing", func() {
					Expect(os.Remove(serverCertificatePath)).To(Succeed())
					srv, err := server.New()
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to load the server certificates"))
				})

				It("should fail to start the server when client CA certificates cannot be loaded", func() {
					Expect(os.Remove(clientCACertPath)).To(Succeed())
					srv, err := server.New()
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to load client CA certificates"))
				})

				It("should fail to start the server when client CA certificate is invalid", func() {
					Expect(os.WriteFile(clientCACertPath, []byte("invalid data"), 0644)).To(Succeed())
					srv, err := server.New()
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to append client CA certificate"))
				})
			})

			generateMutualTLSClientTests := func(host string, port uint16) {
				When("an HTTPS client is created without a client certificate", func() {
					var (
						httpClient *http.Client
					)

					BeforeEach(func() {
						caCert, err := os.ReadFile(clientCACertPath)
						Expect(err).To(Not(HaveOccurred()))
						caCertPool := x509.NewCertPool()
						caCertPool.AppendCertsFromPEM(caCert)
						httpClient = &http.Client{
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: false,
									RootCAs:            caCertPool,
								},
							},
						}
					})

					It("should fail to connect to the server", func() {
						request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s:%d%s", host, port, path), nil)
						Expect(err).NotTo(HaveOccurred())
						response, err := httpClient.Do(request)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("tls: certificate required"))
						Expect(response).To(BeNil())
					})
				})

				When("an HTTPS client is created with a client certificate signed by the trusted CA", func() {
					var (
						httpClient *http.Client
					)

					BeforeEach(func() {
						caCert, err := os.ReadFile(serverCertificatePath)
						Expect(err).To(Not(HaveOccurred()))
						caCertPool := x509.NewCertPool()
						caCertPool.AppendCertsFromPEM(caCert)

						httpClient = &http.Client{
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: false,
									RootCAs:            caCertPool,
									Certificates:       []tls.Certificate{clientCertificateKeyPair},
								},
							},
						}
					})

					It("should be able to get the root contents", func() {
						expectSuccessfulRootGet(httpClient, host, port, "https")
					})
				})

				When("an HTTPS client is created with an invalid client certificate", func() {
					var (
						httpClient *http.Client
					)

					BeforeEach(func() {
						invalidClientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
						Expect(err).ToNot(HaveOccurred())

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
						Expect(err).ToNot(HaveOccurred())
						invalidClientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: invalidClientCertBytes})
						invalidClientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(invalidClientPrivateKey)})

						invalidClientCert, err := tls.X509KeyPair(invalidClientCertPEM, invalidClientKeyPEM)
						Expect(err).ToNot(HaveOccurred())

						caCert, err := os.ReadFile(serverCertificatePath)
						Expect(err).To(Not(HaveOccurred()))
						caCertPool := x509.NewCertPool()
						caCertPool.AppendCertsFromPEM(caCert)

						httpClient = &http.Client{
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: false,
									RootCAs:            caCertPool,
									Certificates:       []tls.Certificate{invalidClientCert},
								},
							},
						}
					})

					It("should fail to connect to the server", func() {
						request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s:%d%s", host, port, path), nil)
						Expect(err).NotTo(HaveOccurred())
						response, err := httpClient.Do(request)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("tls: certificate required"))
						Expect(response).To(BeNil())
					})
				})
			}

			generateServerTests("127.0.0.1", 4444, generateMutualTLSClientTests)
		})
	})
})
