package structs

import (
	"reflect"
)

// FieldMetadata is the metadata extracted from a structs field.
type FieldMetadata struct {
	reflectType reflect.Type
	tags        map[string]string
	anonymous   []string
}

// Type returns the fields type.
func (f *FieldMetadata) Type() reflect.Type {
	return f.reflectType
}

// Tags returns a map of tag names to its respective tag content.
func (f *FieldMetadata) Tags() map[string]string {
	return f.tags
}

// Anonymous returns the anonymous struct name chain before getting to a field.
//
//	type DeepExample struct {
//	  DeepField string
//	}
//
//	type Example struct {
//	  DeepExample
//	  Field string
//	}
//
// If calling Anonymous on the struct Example and the field DeepField, ["DeepExample"] would be returned.
func (f *FieldMetadata) Anonymous() []string {
	return f.anonymous
}
