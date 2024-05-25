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

package logger_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"intelligence/pkg/logger"
)

type testFormatter struct {
	entry *logrus.Entry
}

func (f *testFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	f.entry = entry
	return nil, nil
}

var _ = Describe("formatter", func() {
	When("the UTC formatter is called for a log entry", func() {
		It("should set the timezone to UTC", func() {
			var buf bytes.Buffer
			testLogger := logrus.New()
			testLogger.SetLevel(logrus.InfoLevel)
			testLogger.Out = &buf
			testFormatter := &testFormatter{}
			testLogger.SetFormatter(&logger.UTCFormatter{
				Next: testFormatter,
			})
			testLogger.Info("hello world")
			Expect(testFormatter.entry).To(Not(BeNil()))
			Expect(testFormatter.entry.Time.Location().String()).To(Equal("UTC"))
		})
	})
})
