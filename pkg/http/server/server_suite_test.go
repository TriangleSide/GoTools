package server_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/config"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP server test suite.")
}

func unsetEnvironmentVariables() {
	Expect(os.Unsetenv(string(config.HTTPServerBindPortEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.HTTPServerCertEnvName))).To(Succeed())
	Expect(os.Unsetenv(string(config.HTTPServerKeyEnvName))).To(Succeed())
}
