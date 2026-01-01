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

// NilSourceFuncError indicates an attempt to register a nil SourceFunc for a processor.
type NilSourceFuncError struct {
	ProcessorName string
}

// Error ensures NilSourceFuncError implements the error interface.
func (e *NilSourceFuncError) Error() string {
	return fmt.Sprintf("processor %q requires a non-nil sourcing function", e.ProcessorName)
}

// ProcessorAlreadyRegisteredError indicates an attempt to register a processor with a name that is already in use.
type ProcessorAlreadyRegisteredError struct {
	ProcessorName string
}

// Error ensures ProcessorAlreadyRegisteredError implements the error interface.
func (e *ProcessorAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("processor with name %q already registered", e.ProcessorName)
}
