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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/logger"
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
