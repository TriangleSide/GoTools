package stringcase_test

import (
	"testing"

	"github.com/TriangleSide/GoTools/pkg/stringcase"
	"github.com/TriangleSide/GoTools/pkg/test/assert"
)

func TestStringCase(t *testing.T) {
	t.Parallel()

	t.Run("camel to snake", func(t *testing.T) {
		t.Parallel()
		subTests := []struct {
			value    string
			expected string
		}{
			{"", ""},
			{"a", "A"},
			{"12345", "12345"},
			{"1a", "1A"},
			{"1aSplit", "1A_SPLIT"},
			{"1a1Split", "1A1_SPLIT"},
			{"MyCamelCase", "MY_CAMEL_CASE"},
			{"myCamelCase", "MY_CAMEL_CASE"},
			{"CAMELCase", "CAMEL_CASE"},
		}
		for _, st := range subTests {
			assert.Equals(t, stringcase.CamelToSnake(st.value), st.expected)
		}
	})
}
