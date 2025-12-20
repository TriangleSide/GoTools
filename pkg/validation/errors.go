package validation

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// errValueIsNil is returned when a value is nil while validation requires a non-nil value.
	errValueIsNil = errors.New("the value is nil")
)

// FieldError represents a failure for a specific validator.
type FieldError struct {
	err error
}

// NewFieldError instantiates a *FieldError.
func NewFieldError(params *CallbackParameters, err error) *FieldError {
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
	return &FieldError{
		err: fmt.Errorf(builder.String(), err),
	}
}

// Error ensures FieldError has the error interface.
func (v *FieldError) Error() string {
	return v.err.Error()
}

// Unwrap returns the underlying wrapped error.
func (v *FieldError) Unwrap() error {
	if v == nil || v.err == nil {
		return nil
	}
	return errors.Unwrap(v.err)
}
