package attribute

// String creates a new string attribute with the given key and value.
func String(key string, value string) *Attribute {
	return &Attribute{
		key:         key,
		attrType:    TypeString,
		stringValue: value,
	}
}

// Bool creates a new boolean attribute with the given key and value.
func Bool(key string, value bool) *Attribute {
	return &Attribute{
		key:       key,
		attrType:  TypeBool,
		boolValue: value,
	}
}

// Int creates a new integer attribute with the given key and value.
func Int(key string, value int64) *Attribute {
	return &Attribute{
		key:      key,
		attrType: TypeInt,
		intValue: value,
	}
}

// Float creates a new float attribute with the given key and value.
func Float(key string, value float64) *Attribute {
	return &Attribute{
		key:        key,
		attrType:   TypeFloat,
		floatValue: value,
	}
}
