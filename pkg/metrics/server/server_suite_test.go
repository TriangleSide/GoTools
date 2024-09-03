package metrics_server_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TriangleSide/GoBase/pkg/config"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics server test suite.")
}

func unsetEnvironmentVariables() {
	Expect(os.Unsetenv(string(config.MetricsKeyEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsPortEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsHostEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsBindIPEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsQueueSizeEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsOsBufferSizeEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsReadBufferSizeEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.MetricsReadThreadsEnvName))).To(Succeed())
}
