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
