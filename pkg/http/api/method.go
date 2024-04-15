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
	err := validation.Validate(httpMethod)
	if err != nil {
		panic(fmt.Sprintf("HTTP method '%s' is invalid (%s).", method, err.Error()))
	}
	return httpMethod
}

// String returns the method as a string.
func (m Method) String() string {
	return m.value
}
