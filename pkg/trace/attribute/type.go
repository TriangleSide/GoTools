package attribute

// Type represents an attributes data type.
type Type int

const (
	// TypeString represents a string attribute type.
	TypeString Type = iota + 1

	// TypeInt represents a integer attribute type.
	TypeInt

	// TypeFloat represents a float attribute type.
	TypeFloat

	// TypeBool represents a boolean attribute type.
	TypeBool
)
