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

package ptr

import (
	"reflect"
)

// Of allocates a new instance of the given type, copies the value into it, and returns it.
// This can be used as a utility to make pointers to static values.
// For example: Of[uint](123) returns a uint pointer containing the value 123.
func Of[T any](val T) *T {
	if reflect.TypeOf(val).Kind() == reflect.Ptr {
		panic("type cannot be a pointer")
	}
	valPtr := new(T)
	*valPtr = val
	return valPtr
}
