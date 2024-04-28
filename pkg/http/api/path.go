package api

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"intelligence/pkg/validation"
)

const (
	PathTag = "api_path"
)

// init adds a validator for the Path.
func init() {
	isValidCharacters := regexp.MustCompile(`^[a-zA-Z0-9/{}]+$`).MatchString
	errMsgForValidation := func(value any) string {
		path, ok := value.(string)
		if !ok {
			return "path must be a string"
		}
		if len(path) == 0 {
			return "path cannot be empty"
		}
		if path == "/" {
			return ""
		}
		if !isValidCharacters(path) {
			return "path contains invalid characters"
		}
		if !strings.HasPrefix(path, "/") {
			return "path must start with '/'"
		}
		if strings.HasSuffix(path, "/") {
			return "path cannot end with '/'"
		}
		parts := strings.Split(path, "/")
		parameters := map[string]bool{}
		for i := 1; i < len(parts); i++ {
			part := parts[i]
			if part == "" {
				return "path parts cannot be empty"
			}
			if _, foundPart := parameters[part]; foundPart {
				return "path part must be unique"
			}
			parameters[part] = true
			if strings.Contains(part, "{") || strings.Contains(part, "}") {
				if !strings.HasPrefix(part, "{") || !strings.HasSuffix(part, "}") {
					return "path parameters must start with '{' and end with '}'"
				}
				if strings.Count(part, "{") != 1 || strings.Count(part, "}") != 1 {
					return "path parameters have only one '{' and '}'"
				}
				if len(part) <= 2 {
					return "path parameters cannot be empty"
				}
			}
		}
		return ""
	}
	validation.RegisterValidation(PathTag, func(field validator.FieldLevel) bool {
		if field.Field().Type().Kind() == reflect.Ptr && field.Field().IsNil() {
			return false
		}
		path := field.Field().String()
		return errMsgForValidation(path) == ""
	}, func(err validator.FieldError) string {
		return errMsgForValidation(err.Value())
	})
}

// Path represents the path to an HTTP endpoint.
// For example: /library/book/{bookId}
type Path struct {
	value string `validate:"api_path"`
}

// NewPath allocates, configures, and validates a Path.
// This function panics if the path is not correctly formatted.
func NewPath(apiPath string) Path {
	path := Path{
		value: apiPath,
	}
	err := validation.Struct(path)
	if err != nil {
		panic(fmt.Sprintf("Path '%s' is invalid (%s).", path, err.Error()))
	}
	return path
}

// String returns the Path as a string.
func (path Path) String() string {
	return path.value
}
