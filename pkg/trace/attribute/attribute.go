package attribute

import (
	"strconv"
)

// Attribute represents a key-value pair with a specific type.
type Attribute struct {
	key string

	attrType Type

	intValue    int64
	stringValue string
	floatValue  float64
	boolValue   bool
}

// Key returns the attribute's key.
func (a *Attribute) Key() string {
	return a.key
}

// Type returns the attribute's type.
func (a *Attribute) Type() Type {
	return a.attrType
}

// IntValue returns the attribute's value as an int64.
func (a *Attribute) IntValue() int64 {
	return a.intValue
}

// StringValue returns the attribute's value as a string.
func (a *Attribute) StringValue() string {
	return a.stringValue
}

// FloatValue returns the attribute's value as a float64.
func (a *Attribute) FloatValue() float64 {
	return a.floatValue
}

// BoolValue returns the attribute's value as a bool.
func (a *Attribute) BoolValue() bool {
	return a.boolValue
}

// AsString returns a string representation of the attribute's value.
func (a *Attribute) AsString() string {
	switch a.attrType {
	case TypeInt:
		return strconv.FormatInt(a.intValue, 10)
	case TypeFloat:
		return strconv.FormatFloat(a.floatValue, 'g', -1, 64)
	case TypeBool:
		return strconv.FormatBool(a.boolValue)
	default:
		return a.stringValue
	}
}
