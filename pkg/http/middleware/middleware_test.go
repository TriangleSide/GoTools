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

package middleware_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/http/middleware"
)

var _ = Describe("middleware", func() {
	When("an http handler is created", func() {
		const (
			handlerInvocationName = "handler"
		)

		var (
			mwInvocations []string
			handler       http.HandlerFunc
		)

		BeforeEach(func() {
			mwInvocations = []string{}
			handler = func(w http.ResponseWriter, req *http.Request) {
				mwInvocations = append(mwInvocations, handlerInvocationName)
			}
		})

		When("the middleware chain is created and invoked with a nil middleware list and the handler", func() {
			BeforeEach(func() {
				mwChain := middleware.CreateChain(nil, handler)
				mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
			})

			It("should only call the handler", func() {
				Expect(mwInvocations).To(Equal([]string{handlerInvocationName}))
			})
		})

		When("the middleware chain is created and invoked with an empty middleware list and the handler", func() {
			BeforeEach(func() {
				mw := make([]middleware.Middleware, 0)
				mwChain := middleware.CreateChain(mw, handler)
				mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
			})

			It("should only call the handler", func() {
				Expect(mwInvocations).To(Equal([]string{handlerInvocationName}))
			})
		})

		When("the middleware chain is created and invoked with two middleware and the handler", func() {
			const (
				mw1InvocationName = "mw1"
				mw2InvocationName = "mw2"
			)

			BeforeEach(func() {
				mwList := []middleware.Middleware{
					func(next http.HandlerFunc) http.HandlerFunc {
						return func(writer http.ResponseWriter, request *http.Request) {
							mwInvocations = append(mwInvocations, mw1InvocationName)
							next(writer, request)
						}
					},
					func(next http.HandlerFunc) http.HandlerFunc {
						return func(writer http.ResponseWriter, request *http.Request) {
							mwInvocations = append(mwInvocations, mw2InvocationName)
							next(writer, request)
						}
					},
				}
				mwChain := middleware.CreateChain(mwList, handler)
				mwChain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
			})

			It("should have executed the middleware and handler in order", func() {
				Expect(mwInvocations).To(Equal([]string{mw1InvocationName, mw2InvocationName, handlerInvocationName}))
			})
		})

	})
})
