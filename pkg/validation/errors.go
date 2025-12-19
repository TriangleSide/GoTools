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

// Error ensures Violation has the error interface.
