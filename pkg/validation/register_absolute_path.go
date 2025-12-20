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
	MustRegisterValidator(AbsolutePathValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.SetError(NewFieldError(params, err))
		}

		if value.Kind() != reflect.String {
			return result.SetError(fmt.Errorf("the value must be a string for the %s validator", AbsolutePathValidatorName))
		}

		strValue := value.String()
		if !filepath.IsAbs(strValue) {
			return result.SetError(NewFieldError(params, fmt.Errorf("the path '%s' is not absolute", strValue)))
		}

		if fsPath := absolutePathToFSPath(strValue); fsPath != "" && !fs.ValidPath(fsPath) {
			return result.SetError(NewFieldError(params, fmt.Errorf("the path '%s' is not valid", strValue)))
		}

		if _, err := os.Stat(strValue); err != nil {
			return result.SetError(NewFieldError(params, fmt.Errorf("the path '%s' is not accessible", strValue)))
		}

		return nil
	})
}

// absolutePathToFSPath normalizes an absolute path to the format expected by io.ValidPath.
func absolutePathToFSPath(path string) string {
	normalized := filepath.ToSlash(path)
	return strings.TrimLeft(normalized, "/")
}
