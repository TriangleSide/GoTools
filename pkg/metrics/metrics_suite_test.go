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
