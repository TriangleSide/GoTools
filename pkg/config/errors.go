package config

import (
	"fmt"
)

// FieldAssignmentError represents a failure to assign a value to a configuration field.
type FieldAssignmentError struct {
	FieldName string
	Value     string
	Err       error
}

// Error ensures FieldAssignmentError implements the error interface.
func (e *FieldAssignmentError) Error() string {
	return fmt.Sprintf("failed to assign value %s to field %s: %s", e.Value, e.FieldName, e.Err.Error())
}

// Unwrap returns the underlying wrapped error.
func (e *FieldAssignmentError) Unwrap() error {
	if e == nil || e.Err == nil {
		return nil
	}
	return e.Err
}
