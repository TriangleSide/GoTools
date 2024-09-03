package metrics_client_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/config"
	"github.com/TriangleSide/GoBase/pkg/crypto/symmetric"
	"github.com/TriangleSide/GoBase/pkg/metrics"
	metricsclient "github.com/TriangleSide/GoBase/pkg/metrics/client"
	"github.com/TriangleSide/GoBase/pkg/network/udp"
	udpclient "github.com/TriangleSide/GoBase/pkg/network/udp/client"
)

type udpFailWriter struct {
	udp.Conn
}

func (w *udpFailWriter) Write([]byte) (int, error) {
	return 0, errors.New("write fail")
}

type udpWrongLengthWriter struct {
	udp.Conn
}

func (w *udpWrongLengthWriter) Write(b []byte) (int, error) {
	return len(b) / 2, nil
}

type failEncryptor struct {
	symmetric.Encryptor
}

func (e *failEncryptor) Encrypt([]byte) ([]byte, error) {
	return nil, errors.New("encrypt fail")
}

var _ = Describe("metrics client", func() {
	AfterEach(func() {
		unsetEnvironmentVariables()
	})

	When("a metric client is created without the needed environment variable configuration", func() {
		It("should return an error", func() {
			client, err := metricsclient.New()
			Expect(err).To(HaveOccurred())
			Expect(client).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to get the metrics client configuration"))
		})
	})

	When("defaults are set for the environment variables", func() {
		BeforeEach(func() {
			Expect(os.Setenv(string(config.MetricsKeyEnvName), "encryption_key")).To(Succeed())
			Expect(os.Setenv(string(config.MetricsHostEnvName), "::1")).To(Succeed())
			Expect(os.Setenv(string(config.MetricsPortEnvName), "12345")).To(Succeed())
		})

		It("should be able to create a new metrics client", func() {
			client, err := metricsclient.New()
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		When("creating the connection to the server fails", func() {
			It("should return an error", func() {
				metricsClient, err := metricsclient.New(metricsclient.WithUDPClientProvider(func(remoteHost string, remotePort uint16) (udp.Conn, error) {
					return nil, errors.New("failed to create UDP connection")
				}))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create the metrics client (failed to create UDP connection)"))
				Expect(metricsClient).To(BeNil())
			})
		})

		When("processing the options fails", func() {
			It("should return an error", func() {
				metricsClient, err := metricsclient.New(func(clientConfig *metricsclient.Config) error {
					return errors.New("error")
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to configure metrics client (error)"))
				Expect(metricsClient).To(BeNil())
			})
		})

		When("the hostname environment variable is set to a value that is incorrectly formatted", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsHostEnvName), "!@#$%^&*()_+")).To(Succeed())
			})

			It("should return an error when creating a client", func() {
				client, err := metricsclient.New()
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to format the UDP address"))
			})
		})

		When("the hostname environment variable is set to a value that doesnt exist", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsHostEnvName), "doesnotexist.doesnotexist")).To(Succeed())
			})

			It("should return an error when creating a client", func() {
				client, err := metricsclient.New()
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to resolve the UDP address"))
			})
		})

		When("the encryption key environment variable is set to an empty value", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsKeyEnvName), "")).To(Succeed())
			})

			It("should return an error when creating a client", func() {
				client, err := metricsclient.New()
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsKey'"))
			})
		})

		When("the encryptor fails to be created", func() {
			It("should return an error when creating a metrics client", func() {
				client, err := metricsclient.New(metricsclient.WithEncryptorProvider(func(key string) (symmetric.Encryptor, error) {
					return nil, errors.New("encryptor error")
				}))
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to create the encryptor (encryptor error)"))
			})
		})

		When("the encryptor fails to encrypt", func() {
			It("should return an error when sending a metric", func() {
				client, err := metricsclient.New(metricsclient.WithEncryptorProvider(func(key string) (symmetric.Encryptor, error) {
					encryptor, err := symmetric.New("key")
					Expect(err).ToNot(HaveOccurred())
					return &failEncryptor{
						Encryptor: encryptor,
					}, nil
				}))
				Expect(err).ToNot(HaveOccurred())
				err = client.Send([]*metrics.Metric{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to encrypt the metrics (encrypt fail)"))
			})
		})

		When("a metrics client is created with a UDP connection that fails to write", func() {
			It("should return an error when sending a metric", func() {
				metricsClient, err := metricsclient.New(metricsclient.WithUDPClientProvider(func(remoteHost string, remotePort uint16) (udp.Conn, error) {
					udpConn, err := udpclient.New(remoteHost, remotePort)
					Expect(err).ToNot(HaveOccurred())
					return &udpFailWriter{
						Conn: udpConn,
					}, nil
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsClient).ToNot(BeNil())
				err = metricsClient.Send([]*metrics.Metric{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to send the metrics to the server (write fail)"))
				Expect(metricsClient.Close()).To(Succeed())
			})
		})

		When("a metrics client is created with a UDP connection that doesn't write the whole data", func() {
			It("should return an error when sending a metric", func() {
				metricsClient, err := metricsclient.New(metricsclient.WithUDPClientProvider(func(remoteHost string, remotePort uint16) (udp.Conn, error) {
					udpConn, err := udpclient.New(remoteHost, remotePort)
					Expect(err).ToNot(HaveOccurred())
					return &udpWrongLengthWriter{
						Conn: udpConn,
					}, nil
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsClient).ToNot(BeNil())
				err = metricsClient.Send([]*metrics.Metric{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("only sent"))
				Expect(err.Error()).To(ContainSubstring("bytes to the metrics server"))
				Expect(metricsClient.Close()).To(Succeed())
			})
		})

		When("the marshal function fails when sending a metric", func() {
			It("should return an error", func() {
				metricsClient, err := metricsclient.New(metricsclient.WithMarshallingFunc(func(any) ([]byte, error) {
					return nil, errors.New("marshal error")
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsClient).ToNot(BeNil())
				err = metricsClient.Send([]*metrics.Metric{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to marshal the metrics (marshal error)"))
				Expect(metricsClient.Close()).To(Succeed())
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

			It("should fail to send metrics after the client is closed", func() {
				Expect(metricsClient.Close()).To(Succeed())
				err := metricsClient.Send([]*metrics.Metric{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("metrics client is closed"))
			})

			It("should succeed to send a metric", func() {
				Expect(metricsClient.Send([]*metrics.Metric{})).To(Succeed())
			})
		})
	})
})
