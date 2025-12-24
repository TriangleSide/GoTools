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

const (
	// pathValidationTag is the validation tag name used to validate endpoint paths.
	pathValidationTag = "api_endpoint_path"
)

var (
	// endpointPathValidCharactersRegex matches the allowed characters in an endpoint path string.
	endpointPathValidCharactersRegex = regexp.MustCompile(`^[a-zA-Z0-9/{}]+$`)
)

// init adds a validator for the endpoint paths.
func init() {
	validation.MustRegisterValidator(pathValidationTag, validateEndpointPathValue)
}

// pathValidationFieldError returns a validation field error result for an invalid endpoint path.
func pathValidationFieldError(params *validation.CallbackParameters, err error) (*validation.CallbackResult, error) {
	return validation.NewCallbackResult().AddFieldError(validation.NewFieldError(params, err)), nil
}

// validateEndpointPathValue validates a value as an endpoint path.
func validateEndpointPathValue(params *validation.CallbackParameters) (*validation.CallbackResult, error) {
	path, err := endpointPathFromValue(params.Value)
	if err != nil {
		return pathValidationFieldError(params, err)
	}

	if err := validateEndpointPathString(path); err != nil {
		return pathValidationFieldError(params, err)
	}

	return validation.NewCallbackResult().PassValidation(), nil
}

// endpointPathFromValue extracts an endpoints path string from a potentially dereferenceable value.
func endpointPathFromValue(value reflect.Value) (string, error) {
	dereferenced := reflection.Dereference(value)
	if reflection.IsNil(dereferenced) {
		return "", errors.New("the value is nil")
	}
	if dereferenced.Kind() != reflect.String {
		return "", errors.New("the value must be a string")
	}
	return dereferenced.String(), nil
}

// validateEndpointPathString validates an endpoints path string.
func validateEndpointPathString(path string) error {
	if len(path) == 0 {
		return errors.New("the path cannot be empty")
	}
	if path == "/" {
		return nil
	}
	if !endpointPathValidCharactersRegex.MatchString(path) {
		return errors.New("the path contains invalid characters")
	}
	if !strings.HasPrefix(path, "/") {
		return errors.New("the path must start with '/'")
	}
	if strings.HasSuffix(path, "/") {
		return errors.New("the path cannot end with '/'")
	}

	parts := strings.Split(path, "/")
	if err := validateEndpointPathParts(parts[1:]); err != nil {
		return fmt.Errorf("invalid path parts: %w", err)
	}

	return nil
}

// validateEndpointPathParts validates that the path parts are correctly formatted.
func validateEndpointPathParts(parts []string) error {
	seenParts := make(map[string]struct{}, len(parts)-1)
	for _, part := range parts {
		if err := validateEndpointPathPart(part); err != nil {
			return fmt.Errorf("invalid path part: %w", err)
		}
		if _, foundPart := seenParts[part]; foundPart {
			return errors.New("the path parts must be unique")
		}
		seenParts[part] = struct{}{}
	}
	return nil
}

// validateEndpointPathPart validates a single endpoint path part.
func validateEndpointPathPart(part string) error {
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
