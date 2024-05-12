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

package metrics_test

import (
	"context"
	"os"
	"strconv"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/config"
	"intelligence/pkg/metrics"
	"intelligence/pkg/utils/ptr"
)

var _ = Describe("metrics", func() {
	AfterEach(func() {
		unsetMetricsEnvironmentVariables()
	})

	When("normal configurations for the metrics client and server are set", func() {
		var (
			ctx            context.Context
			metric         *metrics.Metric
			serverBindPort uint16 = 35000
		)

		BeforeEach(func() {
			ctx = context.Background()

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
				metricsServer, err := metrics.NewServer()
				Expect(err).To(HaveOccurred())
				Expect(metricsServer).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsKey'"))
			})
		})

		When("the bind IP environment variable is set to a random value that is incorrectly formatted", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsBindIPEnvName), "!@#$%^&*()_+")).To(Succeed())
			})

			It("should return an error when creating a server", func() {
				server, err := metrics.NewServer()
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
				server, err := metrics.NewServer()
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
				server, err := metrics.NewServer()
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
				server, err := metrics.NewServer()
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
				server, err := metrics.NewServer()
				Expect(err).To(HaveOccurred())
				Expect(server).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsReadThreads' with validator 'lte' and parameter(s) '32'"))
			})
		})

		When("a metrics server and client are created", func() {
			var (
				metricsClient *metrics.Client
				metricsServer *metrics.Server
			)

			BeforeEach(func() {
				var err error
				metricsServer, err = metrics.NewServer()
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsServer).ToNot(BeNil())
				metricsClient, err = metrics.NewClient()
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsClient).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(metricsClient.Close()).To(Succeed())
			})

			It("should be able to send a metric even if the server is not started", func() {
				Expect(metricsClient.Send([]*metrics.Metric{metric})).To(Succeed())
			})

			It("should fail to send metrics after the client is closed", func() {
				Expect(metricsClient.Close()).To(Succeed())
				err := metricsClient.Send([]*metrics.Metric{metric})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("metrics client is closed"))
			})

			When("the metrics server is started", func() {
				var (
					metricsChan chan *metrics.Metric
				)

				BeforeEach(func() {
					metricsChan = metricsServer.Run(ctx)
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
							Expect(metricsServer.Run(ctx)).ToNot(BeNil())
						}).Should(Panic())
					})
				})

				When("a metric is sent by the client", func() {
					BeforeEach(func() {
						Expect(metricsClient.Send([]*metrics.Metric{metric})).To(Succeed())
					})

					It("should be received by the server", func() {
						receivedMetrics := make([]*metrics.Metric, 0)
						Eventually(func() bool {
							for {
								select {
								case receivedMetric := <-metricsChan:
									receivedMetrics = append(receivedMetrics, receivedMetric)
								default:
									return len(receivedMetrics) == 1
								}
							}
						}).WithPolling(time.Millisecond).WithTimeout(time.Second * 10).Should(BeTrue())
						Expect(receivedMetrics).To(HaveLen(1))
						Expect(receivedMetrics[0].Namespace).To(Equal(metric.Namespace))
						Expect(*receivedMetrics[0].Measurement).To(BeNumerically("~", 12.34, 0.001))
						Expect(receivedMetrics[0].Scopes).To(Equal(metric.Scopes))
						Expect(receivedMetrics[0].Timestamp.Round(time.Microsecond)).To(Equal(metric.Timestamp.Round(time.Microsecond)))
					})
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
