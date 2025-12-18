package validation

import (
	"errors"
	"fmt"
	"strings"
)

// Violation represents a failure for a specific validator.
type Violation struct {
	parameters *CallbackParameters
	err        error
}

// NewViolation instantiates a *Violation.
func NewViolation(params *CallbackParameters, err error) *Violation {
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
	return &Violation{
		parameters: params,
		err:        fmt.Errorf(builder.String(), err),
	}
}

// Error ensures Violation has the error interface.
func (v *Violation) Error() string {
	return v.err.Error()
}

// Unwrap returns the underlying wrapped error.
func (v *Violation) Unwrap() error {
	if v == nil || v.err == nil {
		return nil
	}
	return errors.Unwrap(v.err)
}

// Violations represents a list of violations.
type Violations struct {
	violations []*Violation
}

// NewViolations instantiates a *Violations struct.
func NewViolations() *Violations {
	return &Violations{
		violations: make([]*Violation, 0, 1),
	}
}

// AddViolations appends other violations.
func (v *Violations) AddViolations(others *Violations) {
	if others != nil && len(others.violations) > 0 {
		v.violations = append(v.violations, others.violations...)
	}
}

// AddViolation appends another violation to this list of violations.
func (v *Violations) AddViolation(other *Violation) {
	if other != nil {
		v.violations = append(v.violations, other)
	}
}

// NilIfEmpty returns nil if the violation list is empty.
func (v *Violations) NilIfEmpty() error {
	if len(v.violations) == 0 {
		return nil
	}
	return v
}

// Error ensures Violations has the error interface.
func (v *Violations) Error() string {
	errorStrings := make([]string, 0, len(v.violations))
	for _, violation := range v.violations {
		errorStrings = append(errorStrings, violation.Error())
	}
	return strings.Join(errorStrings, "; ")
}

// Unwrap returns the underlying violations as errors.
func (v *Violations) Unwrap() []error {
	if v == nil || len(v.violations) == 0 {
		return nil
	}

	errs := make([]error, 0, len(v.violations))
	for _, violation := range v.violations {
		errs = append(errs, violation)
	}

	return errs
}
