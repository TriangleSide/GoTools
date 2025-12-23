package endpoints

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/TriangleSide/GoTools/pkg/reflection"
	"github.com/TriangleSide/GoTools/pkg/validation"
)

// pathValidationTag is the validation tag name used to validate route paths.
const (
	pathValidationTag = "route_path"
)

var (
	// routePathValidCharactersRegex matches the allowed characters in a route path string.
	routePathValidCharactersRegex = regexp.MustCompile(`^[a-zA-Z0-9/{}]+$`)
)

// init adds a validator for the Path.
func init() {
	validation.MustRegisterValidator(pathValidationTag, validateRoutePathValue)
}

// pathValidationFieldError returns a validation field error result for an invalid route path.
func pathValidationFieldError(params *validation.CallbackParameters, err error) (*validation.CallbackResult, error) {
	return validation.NewCallbackResult().AddFieldError(validation.NewFieldError(params, err)), nil
}

// validateRoutePathValue validates a value as a route path.
func validateRoutePathValue(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
	path, err := routePathFromValue(params.Value)
	if err != nil {
		return pathValidationFieldError(params, err)
	}

	if err := validateRoutePathString(path); err != nil {
		return pathValidationFieldError(params, err)
	}

	return validation.NewCallbackResult().PassValidation(), nil
}

// routePathFromValue extracts a route path string from a potentially dereferenceable value.
func routePathFromValue(value reflect.Value) (string, error) {
	dereferenced := reflection.Dereference(value)
	if reflection.IsNil(dereferenced) {
		return "", errors.New("the value is nil")
	}
	if dereferenced.Kind() != reflect.String {
		return "", errors.New("the value must be a string")
	}
	return dereferenced.String(), nil
}

// validateRoutePathString validates a route path string.
func validateRoutePathString(path string) error {
	if len(path) == 0 {
		return errors.New("the path cannot be empty")
	}
	if path == "/" {
		return nil
	}
	if !routePathValidCharactersRegex.MatchString(path) {
		return errors.New("the path contains invalid characters")
	}
	if !strings.HasPrefix(path, "/") {
		return errors.New("the path must start with '/'")
	}
	if strings.HasSuffix(path, "/") {
		return errors.New("the path cannot end with '/'")
	}

	parts := strings.Split(path, "/")
	if err := validateRoutePathParts(parts[1:]); err != nil {
		return fmt.Errorf("invalid path parts: %w", err)
	}

	return nil
}

// validateRoutePathParts validates the path parts after splitting on '/'.
func validateRoutePathParts(parts []string) error {
	seenParts := make(map[string]struct{}, len(parts)-1)
	for _, part := range parts {
		if err := validateRoutePathPart(part); err != nil {
			return fmt.Errorf("invalid path part: %w", err)
		}
		if _, foundPart := seenParts[part]; foundPart {
			return errors.New("the path parts must be unique")
		}
		seenParts[part] = struct{}{}
	}
	return nil
}

// validateRoutePathPart validates a single route path part.
func validateRoutePathPart(part string) error {
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
