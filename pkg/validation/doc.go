/*
Package validation provides struct and variable validation using struct tags.

Use this package to validate Go values against constraints defined in struct
field tags. The package supports validating entire structs with the Struct
function or individual variables with the Var function.

Validation rules are specified using the "validate" struct tag. Multiple rules
can be combined with commas. For example:

	type User struct {
		Name  string `validate:"required,min=1,max=100"`
		Age   int    `validate:"required,gte=0,lte=150"`
		Email string `validate:"required"`
	}

Built-in validators include required, omitempty, gt, gte, lt, lte, len, min,
max, oneof, dive, ip_addr, absolute_path, filepath, and required_if. Custom
validators can be registered to extend the validation capabilities. Aliases
can also be registered to create reusable combinations of validators.

Validation errors are returned as joined errors containing field-specific
information including the field name, validator name, and the reason for
failure.
*/
package validation
