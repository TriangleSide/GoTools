package validation

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	// AbsolutePathValidatorName is the name of the validator that enforces absolute, valid, existing filesystem paths.
	AbsolutePathValidatorName Validator = "absolute_path"
)

// init registers the absolute_path validator that enforces absolute, valid, existing filesystem paths.
func init() {
	MustRegisterValidator(AbsolutePathValidatorName, func(params *CallbackParameters) (*CallbackResult, error) {
		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return NewCallbackResult().AddFieldError(NewFieldError(params, err)), nil
		}

		if value.Kind() != reflect.String {
			return nil, fmt.Errorf("the value must be a string for the %s validator", AbsolutePathValidatorName)
		}

		strValue := value.String()
		if !filepath.IsAbs(strValue) {
			fieldErr := NewFieldError(params, fmt.Errorf("the path '%s' is not absolute", strValue))
			return NewCallbackResult().AddFieldError(fieldErr), nil
		}

		if fsPath := absolutePathToFSPath(strValue); fsPath != "" && !fs.ValidPath(fsPath) {
			fieldErr := NewFieldError(params, fmt.Errorf("the path '%s' is not valid", strValue))
			return NewCallbackResult().AddFieldError(fieldErr), nil
		}

		if _, err := os.Stat(strValue); err != nil {
			fieldErr := NewFieldError(params, fmt.Errorf("the path '%s' is not accessible", strValue))
			return NewCallbackResult().AddFieldError(fieldErr), nil //nolint:nilerr // returning field error
		}

		return nil, nil //nolint:nilnil // nil, nil means validation passed
	})
}

// absolutePathToFSPath normalizes an absolute path to the format expected by io.ValidPath.
func absolutePathToFSPath(path string) string {
	normalized := filepath.ToSlash(path)
	return strings.TrimLeft(normalized, "/")
}
