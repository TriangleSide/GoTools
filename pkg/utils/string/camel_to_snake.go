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
