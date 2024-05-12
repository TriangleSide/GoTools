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

package api

import (
	"fmt"

	"intelligence/pkg/validation"
)

// Method is a command used by a client to indicate the desired action to be performed
// on a specified resource within a server as part of the HTTP protocol.
type Method struct {
	value string `validate:"required,oneof=GET POST HEAD PUT PATCH DELETE CONNECT OPTIONS TRACE"`
}

// NewMethod allocates, configures and validates a Method.
// This function panics if the method is not supported.
func NewMethod(method string) Method {
	httpMethod := Method{
		value: method,
	}
	err := validation.Struct(httpMethod)
	if err != nil {
		panic(fmt.Sprintf("HTTP method '%s' is invalid (%s).", method, err.Error()))
	}
	return httpMethod
}

// String returns the method as a string.
func (m Method) String() string {
	return m.value
}
