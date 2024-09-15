package server_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
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

var _ = Describe("http server", func() {
	AfterEach(func() {
		unsetEnvironmentVariables()
	})

	When("an endpoint handler is created that returns PONG on requests to root and has middleware", func() {
		const (
			path       = "/"
			body       = "PONG"
			mwSetValue = "set"
		)

		var (
			ctx            context.Context
			handlerMwValue string
			handlerMw      []middleware.Middleware
			handlers       []api.HTTPEndpointHandler
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
		})

		When("common middleware is created", func() {
			const (
				commonMwValueSet = "commonMwValueSet"
			)

			var (
				commonMwValue string
				commonMw      []middleware.Middleware
			)

			BeforeEach(func() {
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

			generateServerRunErrorCases := func() {
				It("should fail if the environment variables fail to be parsed", func() {
					srv, err := server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
						return nil, errors.New("config error")
					}))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not load configuration (config error)"))
					Expect(srv).To(BeNil())
				})

				It("should fail if the server address is incorrectly formatted", func() {
					srv, err := server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
						cfg, err := envprocessor.ProcessAndValidate[config.HTTPServer]()
						Expect(err).ToNot(HaveOccurred())
						cfg.HTTPServerBindIP = "not_an_ip"
						return cfg, nil
					}))
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to create the network listener"))
				})

				It("should fail if the server keys are missing", func() {
					srv, err := server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
						cfg, err := envprocessor.ProcessAndValidate[config.HTTPServer]()
						Expect(err).ToNot(HaveOccurred())
						cfg.HTTPServerTLS = true
						cfg.HTTPServerKey = ""
						cfg.HTTPServerCert = ""
						return cfg, nil
					}))
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to load the server certificate"))
				})

				It("should fail if the port is is already bound", func() {
					const ip = "::1"
					const port = 6789
					ln, err := tcplistener.New(ip, port)
					Expect(err).ToNot(HaveOccurred())
					defer func() {
						Expect(ln.Close()).To(Succeed())
					}()
					srv, err := server.New(server.WithConfigProvider(func() (*config.HTTPServer, error) {
						cfg, err := envprocessor.ProcessAndValidate[config.HTTPServer]()
						Expect(err).ToNot(HaveOccurred())
						cfg.HTTPServerBindIP = ip
						cfg.HTTPServerBindPort = port
						return cfg, nil
					}))
					Expect(err).NotTo(HaveOccurred())
					err = srv.Run(commonMw, handlers, func() {})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("address already in use"))
				})

				When("the TCP listener is closed unexpectedly when the server is running", func() {
					var (
						listener net.Listener
						srv      *server.Server
					)

					BeforeEach(func() {
						var err error
						listener, err = tcplistener.New("::1", 6789)
						Expect(err).ToNot(HaveOccurred())
						srv, err = server.New(server.WithListenerProvider(func(string, uint16) (tcp.Listener, error) {
							return listener, nil
						}))
						Expect(err).NotTo(HaveOccurred())
					})

					AfterEach(func() {
						Expect(srv.Shutdown(ctx)).To(Succeed())
					})

					It("should return an error", func() {
						srvErrChan := make(chan error, 1)
						waitUntilReady := make(chan bool)
						go func() {
							srvErrChan <- srv.Run(commonMw, handlers, func() {
								close(waitUntilReady)
							})
						}()
						<-waitUntilReady
						Expect(listener.Close()).To(Succeed())
						err := <-srvErrChan
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("error encountered while serving http requests"))
					})
				})
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
						Expect(err).ToNot(HaveOccurred())

						go func() {
							srvErrChan <- srv.Run(commonMw, handlers, func() {
								close(waitUntilReady)
							})
						}()
						<-waitUntilReady
					})

					AfterEach(func() {
						Expect(srv.Shutdown(ctx)).To(Succeed())
						Expect(<-srvErrChan).To(Not(HaveOccurred()))
					})

					It("should panic when started again", func() {
						Expect(func() {
							Expect(srv.Run(commonMw, handlers, func() {})).To(Succeed())
						}).Should(PanicWith(ContainSubstring("HTTP server can only be run once per instance")))
					})

					It("should be able to be shutdown multiple times", func() {
						for i := 0; i < 3; i++ {
							Expect(srv.Shutdown(ctx)).To(Succeed())
						}
					})

					clientTests(host, port)
				})
			}

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

			When("a server certificate and key is generated for TLS", func() {
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
					Expect(os.WriteFile(privateKeyPath, privateKeyPEM, 0644)).To(Succeed())

					certificateTemplate := x509.Certificate{
						SerialNumber: big.NewInt(1),
						Subject: pkix.Name{
							Organization: []string{"Server Tests Inc."},
						},
						NotBefore:             time.Now(),
						NotAfter:              time.Now().Add(24 * time.Hour), // 1 day validity
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

					Expect(os.Setenv(string(config.HTTPServerTLSEnvName), "true")).To(Succeed())
					Expect(os.Setenv(string(config.HTTPServerKeyEnvName), privateKeyPath)).To(Succeed())
					Expect(os.Setenv(string(config.HTTPServerCertEnvName), certificatePath)).To(Succeed())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(tempDir)).To(Succeed())
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

						AfterEach(func() {
							strictHttpClient.CloseIdleConnections()
						})

						It("should fail to connect to the server", func() {
							request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s:%d%s", host, port, path), nil)
							Expect(err).NotTo(HaveOccurred())
							response, err := strictHttpClient.Do(request)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("failed to verify certificate"))
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

						AfterEach(func() {
							httpClient.CloseIdleConnections()
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

						AfterEach(func() {
							httpClient.CloseIdleConnections()
						})

						It("should be able to get the root contents", func() {
							expectSuccessfulRootGet(httpClient, host, port, "https")
						})
					})
				}

				generateServerRunErrorCases()
				generateServerTests("127.0.0.1", 4443, generateTLSClientTests)
				generateServerTests("::1", 4443, generateTLSClientTests)
			})

			When("tls is turned off", func() {
				BeforeEach(func() {
					Expect(os.Setenv(string(config.HTTPServerTLSEnvName), "false")).To(Succeed())
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

						AfterEach(func() {
							strictHttpClient.CloseIdleConnections()
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

						AfterEach(func() {
							httpClient.CloseIdleConnections()
						})

						It("should be able to get the root contents", func() {
							expectSuccessfulRootGet(httpClient, host, port, "http")
						})
					})
				}

				generateServerRunErrorCases()
				generateServerTests("127.0.0.1", 18080, generateInsecureClientTests)
				generateServerTests("::1", 18080, generateInsecureClientTests)
			})
		})
	})
})
