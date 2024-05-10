package metrics_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/config"
	"intelligence/pkg/metrics"
)

var _ = Describe("metrics client", func() {
	AfterEach(func() {
		unsetMetricsEnvironmentVariables()
	})

	When("a metric client is created without the needed environment variable configuration", func() {
		It("should return an error", func() {
			client, err := metrics.NewClient()
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
			client, err := metrics.NewClient()
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		When("the hostname environment variable is set to a value that is incorrectly formatted", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsHostEnvName), "!@#$%^&*()_+")).To(Succeed())
			})

			It("should return an error when creating a client", func() {
				client, err := metrics.NewClient()
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to format the metrics server address"))
			})
		})

		When("the hostname environment variable is set to a value that doesnt exist", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsHostEnvName), "doesnotexist.doesnotexist")).To(Succeed())
			})

			It("should return an error when creating a client", func() {
				client, err := metrics.NewClient()
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to resolve the metrics server address"))
			})
		})

		When("the encryption key environment variable is set to an empty value", func() {
			BeforeEach(func() {
				Expect(os.Setenv(string(config.MetricsKeyEnvName), "")).To(Succeed())
			})

			It("should return an error when creating a client", func() {
				client, err := metrics.NewClient()
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("validation failed on field 'MetricsKey'"))
			})
		})
	})
})
