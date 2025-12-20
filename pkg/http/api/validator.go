package api

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

// pathValidationTag is the validation tag name used to validate API paths.
const (
	pathValidationTag = "api_path"
)

var (
	// apiPathValidCharactersRegex matches the allowed characters in an API path string.
	apiPathValidCharactersRegex = regexp.MustCompile(`^[a-zA-Z0-9/{}]+$`)
)

// init adds a validator for the Path.
func init() {
	validation.MustRegisterValidator(pathValidationTag, validateAPIPathValue)
}

// pathValidationError returns a validation error response for an invalid API path.
func pathValidationError(params *validation.CallbackParameters, err error) *validation.CallbackResult {
	return validation.NewCallbackResult().SetError(validation.NewFieldError(params, err))
}

// validateAPIPathValue validates a value as an API path.
func validateAPIPathValue(params *validation.CallbackParameters) *validation.CallbackResult {
	path, err := apiPathFromValue(params.Value)
	if err != nil {
		return pathValidationError(params, err)
	}

	if err := validateAPIPathString(path); err != nil {
		return pathValidationError(params, err)
	}

	return nil
}

// apiPathFromValue extracts an API path string from a potentially dereferenceable value.
func apiPathFromValue(value reflect.Value) (string, error) {
	dereferenced := reflection.Dereference(value)
	if reflection.IsNil(dereferenced) {
		return "", errors.New("the value is nil")
	}
	if dereferenced.Kind() != reflect.String {
		return "", errors.New("the value must be a string")
	}
	return dereferenced.String(), nil
}

// validateAPIPathString validates an API path string.
func validateAPIPathString(path string) error {
	if len(path) == 0 {
		return errors.New("the path cannot be empty")
	}
	if path == "/" {
		return nil
	}
	if !apiPathValidCharactersRegex.MatchString(path) {
		return errors.New("the path contains invalid characters")
	}
	if !strings.HasPrefix(path, "/") {
		return errors.New("the path must start with '/'")
	}
	if strings.HasSuffix(path, "/") {
		return errors.New("the path cannot end with '/'")
	}

	parts := strings.Split(path, "/")
	if err := validateAPIPathParts(parts[1:]); err != nil {
		return fmt.Errorf("invalid path parts: %w", err)
	}

	return nil
}

// validateAPIPathParts validates the path parts after splitting on '/'.
func validateAPIPathParts(parts []string) error {
	seenParts := make(map[string]struct{}, len(parts)-1)
	for _, part := range parts {
		if err := validateAPIPathPart(part); err != nil {
			return fmt.Errorf("invalid path part: %w", err)
		}
		if _, foundPart := seenParts[part]; foundPart {
			return errors.New("the path parts must be unique")
		}
		seenParts[part] = struct{}{}
	}
	return nil
}

// validateAPIPathPart validates a single API path part.
func validateAPIPathPart(part string) error {
	if part == "" {
		return errors.New("the path parts cannot be empty")
	}
	if !strings.ContainsAny(part, "{}") {
		return nil
	}
	if !strings.HasPrefix(part, "{") || !strings.HasSuffix(part, "}") {
		return errors.New("the path parameters must start with '{' and end with '}'")
	}
	if strings.Count(part, "{") != 1 || strings.Count(part, "}") != 1 {
		return errors.New("the path parameters must have only one '{' and '}'")
	}
	if part == "{}" {
		return errors.New("the path parameters cannot be empty")
	}
	return nil
}
