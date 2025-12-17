package config

import (
	"strings"
	"unicode"
)

// camelToSnake converts a camelCase string to an upper case SNAKE_CASE format.
func camelToSnake(str string) string {
	var snake strings.Builder

	for i, currentRune := range str {
		isUpper := unicode.IsUpper(currentRune)
		notFirstByte := i > 0
		hasNextByte := i+1 < len(str)
		nextByteIsLower := hasNextByte && unicode.IsLower(rune(str[i+1]))
		prevByteIsLower := notFirstByte && unicode.IsLower(rune(str[i-1]))
		shouldInsertUnderscore := notFirstByte && isUpper && (nextByteIsLower || prevByteIsLower)
		if shouldInsertUnderscore {
			snake.WriteRune('_')
		}
		snake.WriteRune(unicode.ToUpper(currentRune))
	}

	return snake.String()
}
