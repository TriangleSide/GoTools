package validation

import (
	"fmt"
	io "io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	AbsolutePathValidatorName Validator = "absolute_path"
)

// init registers the absolute_path validator that enforces absolute, valid, existing filesystem paths.
func init() {
	MustRegisterValidator(AbsolutePathValidatorName, func(params *CallbackParameters) *CallbackResult {
		result := NewCallbackResult()

		value, err := dereferenceAndNilCheck(params.Value)
		if err != nil {
			return result.WithError(NewViolation(params, err))
		}

		if value.Kind() != reflect.String {
			return result.WithError(fmt.Errorf("the value must be a string for the %s validator", AbsolutePathValidatorName))
		}

		strValue := value.String()
		if !filepath.IsAbs(strValue) {
			return result.WithError(NewViolation(params, fmt.Errorf("the path '%s' is not absolute", strValue)))
		}

		if fsPath := absolutePathToFSPath(strValue); fsPath != "" && !io.ValidPath(fsPath) {
			return result.WithError(NewViolation(params, fmt.Errorf("the path '%s' is not valid", strValue)))
		}

		if _, err := os.Stat(strValue); err != nil {
			return result.WithError(NewViolation(params, fmt.Errorf("the path '%s' is not accessible", strValue)))
		}

		return nil
	})
}

// absolutePathToFSPath normalizes an absolute path to the format expected by io.ValidPath.
func absolutePathToFSPath(path string) string {
	normalized := filepath.ToSlash(path)
	volume := filepath.VolumeName(normalized)
	if volume != "" {
		normalized = strings.TrimPrefix(normalized, volume)
	}

	return strings.TrimLeft(normalized, "/")
}
