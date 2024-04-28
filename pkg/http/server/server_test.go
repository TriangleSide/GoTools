package server_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/config"
	"intelligence/pkg/http/api"
	"intelligence/pkg/http/middleware"
	"intelligence/pkg/http/server"
)

type testHandler struct {
	Path       string
	Method     string
	Middleware []middleware.Middleware
	Handler    http.HandlerFunc
}

func (t testHandler) AcceptHTTPAPIBuilder(builder *api.HTTPAPIBuilder) {
	builder.MustRegister(api.NewPath(t.Path), api.NewMethod(t.Method), &api.Handler{
		Middleware: t.Middleware,
		Handler:    t.Handler,
	})
}

var _ = Describe("server", func() {
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

			When("a server certificate and key is generated", func() {
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
				})

				AfterEach(func() {
					Expect(os.RemoveAll(tempDir)).To(Succeed())
				})

				generateClientTests := func(host string, port uint16) {
					When("an HTTP client is created that verifies the server certificate without trusting it", func() {
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

					When("an HTTP client is created that verifies the server certificate and trusts it", func() {
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

						When("an HTTP request is made for the contents at root", func() {
							var (
								response *http.Response
							)

							BeforeEach(func() {
								request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s:%d%s", host, port, path), nil)
								Expect(err).NotTo(HaveOccurred())
								response, err = httpClient.Do(request)
								Expect(err).NotTo(HaveOccurred())
							})

							AfterEach(func() {
								Expect(response.Body.Close()).To(Succeed())
							})

							It("should return 200 OK and PONG in the body and the middleware to have been invoked", func() {
								Expect(response.StatusCode).To(Equal(http.StatusOK))
								Expect(response.Body).To(Not(BeNil()))
								responseBody, err := io.ReadAll(response.Body)
								Expect(err).NotTo(HaveOccurred())
								Expect(string(responseBody)).To(Equal(body))
								Expect(commonMwValue).To(Equal(commonMwValueSet))
								Expect(handlerMwValue).To(Equal(mwSetValue))
							})
						})
					})

					When("an HTTP client is created that doesn't verify the server certificate", func() {
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

						When("an HTTP request is made for the contents at root", func() {
							var (
								response *http.Response
							)

							BeforeEach(func() {
								request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s:%d%s", host, port, path), nil)
								Expect(err).NotTo(HaveOccurred())
								response, err = httpClient.Do(request)
								Expect(err).NotTo(HaveOccurred())
							})

							AfterEach(func() {
								Expect(response.Body.Close()).To(Succeed())
							})

							It("should return 200 OK and PONG in the body and the middleware to have been invoked", func() {
								Expect(response.StatusCode).To(Equal(http.StatusOK))
								Expect(response.Body).To(Not(BeNil()))
								responseBody, err := io.ReadAll(response.Body)
								Expect(err).NotTo(HaveOccurred())
								Expect(string(responseBody)).To(Equal(body))
								Expect(commonMwValue).To(Equal(commonMwValueSet))
								Expect(handlerMwValue).To(Equal(mwSetValue))
							})
						})
					})
				}

				generateServerTests := func(host string, port uint16) {
					When("an HTTP server is bound to IP "+host+" and port "+strconv.Itoa(int(port))+" with common middleware is started", func() {
						var (
							conf config.Server
							srv  *server.Server
						)

						BeforeEach(func() {
							conf = config.Server{
								ServerBindIP:       host,
								ServerBindPort:     port,
								ServerReadTimeout:  time.Minute,
								ServerWriteTimeout: time.Minute,
								ServerKey:          privateKeyPath,
								ServerCert:         certificatePath,
							}
							waitUntilReady := make(chan bool)
							srv = server.New(conf)
							go func() {
								err := srv.Run(ctx, commonMw, handlers, func() {
									close(waitUntilReady)
								})
								Expect(err).ToNot(HaveOccurred())
							}()
							<-waitUntilReady
						})

						AfterEach(func() {
							Expect(srv.Shutdown(ctx)).To(Succeed())
						})

						generateClientTests(host, port)
					})
				}

				generateServerTests("127.0.0.1", 4443)
				generateServerTests("::1", 4443)
			})
		})
	})
})
