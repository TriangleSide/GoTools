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

package string_utils

import (
	"strings"
	"unicode"
)

// CamelToUpperSnake converts a camelCase string to an upper case SNAKE_CASE format.
func CamelToUpperSnake(str string) string {
	var snake strings.Builder
	for i, r := range str {
		if i > 0 && unicode.IsUpper(r) && (i+1 < len(str) && unicode.IsLower(rune(str[i+1])) || unicode.IsLower(rune(str[i-1]))) {
			snake.WriteRune('_')
		}
		snake.WriteRune(unicode.ToUpper(r))
	}
	return snake.String()
}
