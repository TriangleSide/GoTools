package attribute

// Type represents an attribute's data type.
type Type int

const (
	// TypeString represents a string attribute type.
	TypeString Type = iota + 1

	// TypeInt represents an integer attribute type.
	TypeInt

	// TypeFloat represents a float attribute type.
	TypeFloat

	// TypeBool represents a boolean attribute type.
	TypeBool
)

// String returns a string representation of the attribute type.
func (t Type) String() string {
	switch t {
	case TypeString:
		return "String"
	case TypeInt:
		return "Int"
	case TypeFloat:
		return "Float"
	case TypeBool:
		return "Bool"
	default:
		return "Unknown"
	}
}
