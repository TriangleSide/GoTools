package validation

import (
	"errors"
	"fmt"
	"strings"
)

// ViolationError represents a failure for a specific validator.
type ViolationError struct {
	parameters *CallbackParameters
	err        error
}

// NewViolationError instantiates a *ViolationError.
func NewViolationError(params *CallbackParameters, err error) *ViolationError {
	var builder strings.Builder
	builder.WriteString("validation failed")
	if params.IsStructValidation {
		builder.WriteString(" on field '")
		builder.WriteString(params.StructFieldName)
		builder.WriteString("'")
	}
	builder.WriteString(" with validator '")
	builder.WriteString(string(params.Validator))
	builder.WriteString("'")
	if params.Parameters != "" {
		builder.WriteString(" and parameters '")
		builder.WriteString(params.Parameters)
		builder.WriteString("'")
	}
	builder.WriteString(" because %w")
	return &ViolationError{
		parameters: params,
		err:        fmt.Errorf(builder.String(), err),
	}
}

// Error ensures ViolationError has the error interface.
func (v *ViolationError) Error() string {
	return v.err.Error()
}

// Unwrap returns the underlying wrapped error.
func (v *ViolationError) Unwrap() error {
	if v == nil || v.err == nil {
		return nil
	}
	return errors.Unwrap(v.err)
}

// ViolationsError represents a list of violations.
type ViolationsError struct {
	violations []*ViolationError
}

// NewViolationsError instantiates a *ViolationsError struct.
func NewViolationsError() *ViolationsError {
	return &ViolationsError{
		violations: make([]*ViolationError, 0, 1),
	}
}

// AddViolations appends other violations.
func (v *ViolationsError) AddViolations(others *ViolationsError) {
	if others != nil && len(others.violations) > 0 {
		v.violations = append(v.violations, others.violations...)
	}
}

// AddViolation appends another violation to this list of violations.
func (v *ViolationsError) AddViolation(other *ViolationError) {
	if other != nil {
		v.violations = append(v.violations, other)
	}
}

// NilIfEmpty returns nil if the violation list is empty.
func (v *ViolationsError) NilIfEmpty() error {
	if len(v.violations) == 0 {
		return nil
	}
	return v
}

// Error ensures ViolationsError has the error interface.
func (v *ViolationsError) Error() string {
	errorStrings := make([]string, 0, len(v.violations))
	for _, violation := range v.violations {
		errorStrings = append(errorStrings, violation.Error())
	}
	return strings.Join(errorStrings, "; ")
}

// Unwrap returns the underlying violations as errors.
func (v *ViolationsError) Unwrap() []error {
	if v == nil || len(v.violations) == 0 {
		return nil
	}

	errs := make([]error, 0, len(v.violations))
	for _, violation := range v.violations {
		errs = append(errs, violation)
	}

	return errs
}
