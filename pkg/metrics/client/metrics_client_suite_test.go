package metrics_client_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/config"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics client test suite.")
}

func unsetEnvironmentVariables() {
	Expect(os.Unsetenv(string(config.MetricsKeyEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsHostEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsPortEnvName))).To(Succeed())
}
