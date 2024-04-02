package logger_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/logger"
)

var _ = Describe("logger", func() {
	When("there is no log entry in the context", func() {
		var (
			ctx context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()
		})

		It("should return a new log entry when fetched", func() {
			Expect(logger.LogEntry(ctx)).To(Not(BeNil()))
		})

		When("a field is added to log entry", func() {
			const (
				fieldKey   = "test"
				fieldValue = "value"
			)

			BeforeEach(func() {
				logger.WithField(&ctx, fieldKey, fieldValue)
			})

			It("should have the field in the log entry returned from the function", func() {
				entry := logger.WithField(&ctx, fieldKey, fieldValue)
				value, ok := entry.Data[fieldKey]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal(fieldValue))
			})

			It("should have the field in the log entry in the context", func() {
				_ = logger.WithField(&ctx, fieldKey, fieldValue)
				entry := logger.LogEntry(ctx)
				value, ok := entry.Data[fieldKey]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal(fieldValue))
			})
		})
	})
})
