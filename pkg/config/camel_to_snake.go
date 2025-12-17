package config

import (
	"strings"
	"unicode"
)

// camelToSnake converts a camelCase string to an upper case SNAKE_CASE format.
func camelToSnake(str string) string {
	var snake strings.Builder

	runes := []rune(str)
	for i, currentRune := range runes {
		isUpper := unicode.IsUpper(currentRune)
		notFirstRune := i > 0
		hasNextRune := i+1 < len(runes)
		nextRuneIsLower := hasNextRune && unicode.IsLower(runes[i+1])
		prevRuneIsLower := notFirstRune && unicode.IsLower(runes[i-1])
		shouldInsertUnderscore := notFirstRune && isUpper && (nextRuneIsLower || prevRuneIsLower)
		if shouldInsertUnderscore {
			snake.WriteByte('_')
		}
		snake.WriteRune(unicode.ToUpper(currentRune))
	}

	return snake.String()
}
