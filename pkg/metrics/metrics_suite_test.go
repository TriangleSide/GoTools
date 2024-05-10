package metrics_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/config"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics test suite.")
}

func unsetMetricsEnvironmentVariables() {
	Expect(os.Unsetenv(string(config.MetricsKeyEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsHostEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsPortEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsBindIPEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsOsBufferSizeEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsReadBufferSizeEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsReadThreadsEnvName))).To(Succeed())
}
