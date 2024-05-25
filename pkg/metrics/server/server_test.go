// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

package metrics_server_test

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/config"
	"intelligence/pkg/config/envprocessor"
	"intelligence/pkg/crypto/symmetric"
	"intelligence/pkg/metrics"
	metricsclient "intelligence/pkg/metrics/client"
	metricsserver "intelligence/pkg/metrics/server"
	"intelligence/pkg/network/udp"
	udplistener "intelligence/pkg/network/udp/listener"
	"intelligence/pkg/utils/ptr"
)

type udpFailSetReadBuffer struct {
	udp.Conn
}

func (w *udpFailSetReadBuffer) SetReadBuffer(int) error {
	return errors.New("set read buffer fail")
}

type udpHugeReader struct {
	udp.Conn
}

func (w *udpHugeReader) Read([]byte) (n int, err error) {
	return 1 << 30, nil
}

type failDecryptor struct {
	symmetric.Encryptor
}

func (e *failDecryptor) Decrypt([]byte) ([]byte, error) {
	return nil, errors.New("decrypt fail")
}

var _ = Describe("metrics", func() {
	AfterEach(func() {
		unsetEnvironmentVariables()
	})

	When("environment variables with normal values are set for the metrics client and server", func() {
		var (
			metric         *metrics.Metric
			serverBindPort uint16 = 35000
		)

		BeforeEach(func() {
			metric = &metrics.Metric{
				Namespace: "namespace",
				Scopes: map[string]string{
					"type": "value",
				},
				Measurement: ptr.Of[float32](12.34),
				Timestamp:   time.Now(),
			}

			serverBindPort++

			Expect(os.Setenv(string(config.MetricsKeyEnvName), "encryption_key_"+strconv.Itoa(int(serverBindPort)))).To(Succeed())
			Expect(os.Setenv(string(config.MetricsHostEnvName), "::1")).To(Succeed())
			Expect(os.Setenv(string(config.MetricsPortEnvName), strconv.Itoa(int(serverBindPort)))).To(Succeed())
			Expect(os.Setenv(string(config.MetricsBindIPEnvName), "::1")).To(Succeed())
		})

		When("the key environment variable is set to an empty value", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsKeyEnvName), "")).To(Succeed())
			})

			It("should return an error when creating a server", func() {
				metricsServer, err := metricsserver.New()
				Expect(err).To(HaveOccurred())
				Expect(metricsServer).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsKey'"))
			})
		})

		When("the server fails to format the networking details", func() {
			It("should fail to create a server", func() {
				metricsServer, err := metricsserver.New(metricsserver.WithConfigProvider(func() (*config.MetricsServer, error) {
					conf, err := envprocessor.ProcessAndValidate[config.MetricsServer]()
					Expect(err).NotTo(HaveOccurred())
					conf.MetricsBindIP = "300.300.300.300"
					return conf, nil
				}))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to format the UDP address"))
				Expect(metricsServer).To(BeNil())
			})
		})

		When("the bind IP environment variable is set to a random value that is incorrectly formatted", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsBindIPEnvName), "!@#$%^&*()_+")).To(Succeed())
			})

			It("should return an error when creating a server", func() {
				server, err := metricsserver.New()
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsBindIP' with validator 'ip_addr'"))
			})
		})

		When("the bind IP environment variable is set to an incorrectly formatted IP address", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsBindIPEnvName), "127.0.0.256")).To(Succeed())
			})

			It("should return an error when creating a server", func() {
				server, err := metricsserver.New()
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsBindIP' with validator 'ip_addr'"))
			})
		})

		When("the os read buffer environment variable is set to an invalid value", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsOsBufferSizeEnvName), "123")).To(Succeed())
			})

			It("should return an error when creating a server", func() {
				server, err := metricsserver.New()
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsOSBufferSize' with validator 'gte' and parameter(s) '4096'"))
			})
		})

		When("the read buffer environment variable is set to an invalid value", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsReadBufferSizeEnvName), "123")).To(Succeed())
			})

			It("should return an error when creating a server", func() {
				server, err := metricsserver.New()
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsReadBufferSize' with validator 'gte' and parameter(s) '4096'"))
			})
		})

		When("the threads environment variable is set to an invalid value", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsReadThreadsEnvName), "123")).To(Succeed())
			})

			It("should return an error when creating a server", func() {
				server, err := metricsserver.New()
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsReadThreads' with validator 'lte' and parameter(s) '32'"))
			})
		})

		When("the server fails to set the read buffer", func() {
			It("should fail to create a server", func() {
				metricsServer, err := metricsserver.New(metricsserver.WithUDPListenerProvider(func(localHost string, localPort uint16) (udp.Conn, error) {
					conn, err := udplistener.New(localHost, localPort)
					Expect(err).To(Not(HaveOccurred()))
					return &udpFailSetReadBuffer{
						Conn: conn,
					}, nil
				}))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to set the read buffer size"))
				Expect(err.Error()).To(ContainSubstring("set read buffer fail"))
				Expect(metricsServer).To(BeNil())
			})
		})

		When("the encryptor fails to be created", func() {
			It("should return an error when creating a metrics server", func() {
				server, err := metricsserver.New(metricsserver.WithEncryptorProvider(func(key string) (symmetric.Encryptor, error) {
					return nil, errors.New("encryptor error")
				}))
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to create the encryptor (encryptor error)"))
			})
		})

		When("an option return an error", func() {
			It("should return an error", func() {
				server, err := metricsserver.New(func(serverConfig *metricsserver.Config) error {
					return errors.New("option error")
				})
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to configure metrics server (option error)"))
			})
		})

		When("a metrics server is started with an unmarshalling func that fails", func() {
			It("should return an error", func() {
				received := atomic.Bool{}
				received.Store(false)
				metricsServer, err := metricsserver.New(metricsserver.WithUnmarshallingFunc(func(data []byte, v any) error {
					return errors.New("unmarshalling error")
				}), metricsserver.WithErrorHandler(func(err error) {
					defer GinkgoRecover()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to unmarshall the data from the socket (unmarshalling error)"))
					received.Store(true)
				}))
				Expect(err).ToNot(HaveOccurred())
				metricsClient, err := metricsclient.New()
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsClient.Send([]*metrics.Metric{metric})).To(Succeed())
				Expect(metricsClient.Close()).To(Succeed())
				_ = metricsServer.Run()
				Eventually(func() bool {
					return received.Load()
				}).WithPolling(time.Millisecond * 5).WithTimeout(time.Second * 5).Should(BeTrue())
				Expect(metricsServer.Shutdown()).To(Succeed())
			})
		})

		When("the listener is closed while the server is running", func() {
			It("should return an error and the server should get shutdown eventually", func() {
				received := atomic.Bool{}
				received.Store(false)
				serverListener, err := udplistener.New("::1", serverBindPort)
				Expect(err).ToNot(HaveOccurred())
				metricsServer, err := metricsserver.New(metricsserver.WithUDPListenerProvider(func(string, uint16) (udp.Conn, error) {
					return serverListener, nil
				}), metricsserver.WithErrorHandler(func(err error) {
					defer GinkgoRecover()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("use of closed network connection"))
					received.Store(true)
				}))
				Expect(err).ToNot(HaveOccurred())
				_ = metricsServer.Run()
				Expect(serverListener.Close()).To(Succeed())
				Eventually(func() bool {
					return metricsServer.Running()
				}).WithPolling(time.Millisecond).WithTimeout(time.Second * 5).Should(BeFalse())
				Eventually(func() bool {
					return received.Load()
				}).WithPolling(time.Millisecond).WithTimeout(time.Second * 5).Should(BeTrue())
				Expect(metricsServer.Shutdown()).To(Succeed())
			})
		})

		When("when a server is run and the decryption fails", func() {
			It("should return an error eventually", func() {
				received := atomic.Bool{}
				received.Store(false)
				metricsServer, err := metricsserver.New(metricsserver.WithEncryptorProvider(func(key string) (symmetric.Encryptor, error) {
					enc, err := symmetric.New(key)
					Expect(err).To(Not(HaveOccurred()))
					return &failDecryptor{
						Encryptor: enc,
					}, nil
				}), metricsserver.WithErrorHandler(func(err error) {
					defer GinkgoRecover()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("decrypt fail"))
					received.Store(true)
				}))
				Expect(err).ToNot(HaveOccurred())
				_ = metricsServer.Run()
				metricsClient, err := metricsclient.New()
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsClient.Send([]*metrics.Metric{metric})).To(Succeed())
				Expect(metricsClient.Close()).To(Succeed())
				Eventually(func() bool {
					return received.Load()
				}).WithPolling(time.Millisecond).WithTimeout(time.Second * 5).Should(BeTrue())
				Expect(metricsServer.Shutdown()).To(Succeed())
			})
		})

		When("when a server is run and the data read is too large", func() {
			It("should return an error eventually", func() {
				received := atomic.Bool{}
				received.Store(false)
				metricsServer, err := metricsserver.New(metricsserver.WithUDPListenerProvider(func(localHost string, localPort uint16) (udp.Conn, error) {
					conn, err := udplistener.New(localHost, localPort)
					Expect(err).To(Not(HaveOccurred()))
					return &udpHugeReader{
						Conn: conn,
					}, nil
				}), metricsserver.WithErrorHandler(func(err error) {
					defer GinkgoRecover()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("data received from the socket is too large"))
					received.Store(true)
				}))
				Expect(err).ToNot(HaveOccurred())
				_ = metricsServer.Run()
				metricsClient, err := metricsclient.New()
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsClient.Send([]*metrics.Metric{metric})).To(Succeed())
				Expect(metricsClient.Close()).To(Succeed())
				Eventually(func() bool {
					return received.Load()
				}).WithPolling(time.Millisecond).WithTimeout(time.Second * 5).Should(BeTrue())
				Expect(metricsServer.Shutdown()).To(Succeed())
			})
		})

		When("when a server is run and clients sent more metrics then the server can process", func() {
			It("should return an error eventually", func() {
				Expect(os.Setenv(string(config.MetricsQueueSizeEnvName), "1")).To(Succeed())
				received := atomic.Bool{}
				received.Store(false)
				metricsServer, err := metricsserver.New(metricsserver.WithErrorHandler(func(err error) {
					defer GinkgoRecover()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("metric queue is full"))
					received.Store(true)
				}))
				Expect(err).ToNot(HaveOccurred())
				_ = metricsServer.Run()
				metricsClient, err := metricsclient.New()
				Expect(err).ToNot(HaveOccurred())
				for i := 0; i < 2; i++ {
					Expect(metricsClient.Send([]*metrics.Metric{metric})).To(Succeed())
				}
				Expect(metricsClient.Close()).To(Succeed())
				Eventually(func() bool {
					return received.Load()
				}).WithPolling(time.Millisecond).WithTimeout(time.Second * 5).Should(BeTrue())
				Expect(metricsServer.Shutdown()).To(Succeed())
			})
		})

		When("a metrics server is created and started", func() {
			var (
				metricsServer *metricsserver.Server
				metricsChan   chan *metrics.Metric
			)

			BeforeEach(func() {
				var err error
				metricsServer, err = metricsserver.New()
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsServer).ToNot(BeNil())
				metricsChan = metricsServer.Run()
				Expect(metricsServer.Running()).To(BeTrue())
			})

			AfterEach(func() {
				Expect(metricsServer.Running()).To(BeTrue())
				Expect(metricsServer.Shutdown()).To(Succeed())
				Expect(metricsServer.Running()).To(BeFalse())
			})

			When("run is called again", func() {
				It("should panic", func() {
					Expect(func() {
						Expect(metricsServer.Run()).ToNot(BeNil())
					}).Should(PanicWith(ContainSubstring("metrics server can only be run once per instance")))
				})
			})

			When("another metrics server is created on the same address", func() {
				It("should return an error", func() {
					metricsServer, err := metricsserver.New()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("address already in use"))
					Expect(metricsServer).To(BeNil())
				})
			})

			When("a metrics client is created", func() {
				var (
					metricsClient *metricsclient.Client
				)

				BeforeEach(func() {
					var err error
					metricsClient, err = metricsclient.New()
					Expect(err).ToNot(HaveOccurred())
					Expect(metricsClient).ToNot(BeNil())
				})

				AfterEach(func() {
					Expect(metricsClient.Close()).To(Succeed())
				})

				When("many metrics are sent by the client in a batch", func() {
					var (
						metricsList []*metrics.Metric
					)

					BeforeEach(func() {
						for i := 0; i < 3; i++ {
							metricsCopy := *metric
							metricsCopy.Namespace = strconv.Itoa(i)
							metricsList = append(metricsList, &metricsCopy)
						}
						Expect(metricsClient.Send(metricsList)).To(Succeed())
					})

					It("should be received by the server", func() {
						receivedMetrics := make([]*metrics.Metric, 0)
						Eventually(func() bool {
							for {
								select {
								case receivedMetric := <-metricsChan:
									receivedMetrics = append(receivedMetrics, receivedMetric)
								default:
									return len(receivedMetrics) == 3
								}
							}
						}).WithPolling(time.Millisecond).WithTimeout(time.Second * 10).Should(BeTrue())
						Expect(receivedMetrics).To(HaveLen(3))

						namespaces := make(map[string]struct{})
						for _, receivedMetric := range receivedMetrics {
							namespaces[receivedMetric.Namespace] = struct{}{}
						}

						Expect(namespaces).To(HaveLen(len(metricsList)))
						for _, sentMetric := range metricsList {
							_, found := namespaces[sentMetric.Namespace]
							Expect(found).To(BeTrue())
						}
					})
				})

				When("many metrics are sent by the client one by one", func() {
					BeforeEach(func() {
						for i := 0; i < 3; i++ {
							metricsCopy := *metric
							metricsCopy.Namespace = strconv.Itoa(i)
							Expect(metricsClient.Send([]*metrics.Metric{&metricsCopy})).To(Succeed())
						}
					})

					It("should be received by the server", func() {
						receivedMetrics := make([]*metrics.Metric, 0)
						Eventually(func() bool {
							for {
								select {
								case receivedMetric := <-metricsChan:
									receivedMetrics = append(receivedMetrics, receivedMetric)
								default:
									return len(receivedMetrics) == 3
								}
							}
						}).WithPolling(time.Millisecond).WithTimeout(time.Second * 10).Should(BeTrue())
						Expect(receivedMetrics).To(HaveLen(3))

						namespaces := make(map[string]struct{})
						for _, receivedMetric := range receivedMetrics {
							namespaces[receivedMetric.Namespace] = struct{}{}
						}

						Expect(namespaces).To(HaveLen(3))
						for i := 0; i < 3; i++ {
							_, found := namespaces[strconv.Itoa(i)]
							Expect(found).To(BeTrue())
						}
					})
				})

				When("metrics are sent by the client concurrently", func() {
					It("should be received by the server", func() {
						const threadCount int = 2
						const metricsEach int = 250

						wg := &sync.WaitGroup{}

						for threadIndex := 0; threadIndex < threadCount; threadIndex++ {
							wg.Add(1)
							go func(threadIndex int) {
								defer func() {
									GinkgoRecover()
								}()
								defer func() {
									wg.Done()
								}()
								for metricIndex := 0; metricIndex < metricsEach; metricIndex++ {
									metricCopy := *metric
									metricCopy.Namespace = "namespace_" + strconv.Itoa(threadIndex) + "_" + strconv.Itoa(metricIndex)
									Expect(metricsClient.Send([]*metrics.Metric{&metricCopy})).To(Succeed())
								}
							}(threadIndex)
						}

						countReceived := 0
						Eventually(func() bool {
							Expect(metricsServer.Running()).To(BeTrue())
							for {
								select {
								case <-metricsChan:
									countReceived++
								default:
									return countReceived == threadCount*metricsEach
								}
							}
						}).WithPolling(time.Millisecond).WithTimeout(time.Second * 10).Should(BeTrue())

						wg.Wait()
					})
				})
			})
		})
	})
})
